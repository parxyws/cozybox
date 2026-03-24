package service

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/parxyws/cozybox/internal/dto"
	"github.com/parxyws/cozybox/internal/models"
	"github.com/parxyws/cozybox/internal/tools/contextkey"
	"github.com/parxyws/cozybox/internal/tools/mail"
	"github.com/parxyws/cozybox/internal/tools/util"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db      *gorm.DB
	mailer  *mail.Mailer
	authRds *redis.Client
	jwt     util.TokenMaker
}

const (
	OTP_LENGTH = 6
	OTP_EXPIRY = 2 * time.Minute
)

type UserServiceInterface interface{}

func NewUserService(db *gorm.DB, mailer *mail.Mailer, authRds *redis.Client, jwt util.TokenMaker) *UserService {
	return &UserService{db: db, mailer: mailer, authRds: authRds, jwt: jwt}
}

func (u *UserService) Register(ctx context.Context, req *dto.RegisterUserRequest) (*dto.RegisterResponse, error) {
	currentTime := time.Now()
	identifier := base64.StdEncoding.EncodeToString([]byte(req.Email))
	referenceId := fmt.Sprintf("%s-%s", identifier, ulid.Make().String()) // generate random string for email verification
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to generate password")
	}

	request := &models.User{
		Id:         ulid.Make().String(),
		Name:       req.Name,
		Username:   req.Username,
		Email:      req.Email,
		Password:   string(hashedPassword),
		IsVerified: false,
		CreatedAt:  currentTime,
		UpdatedAt:  currentTime,
	}

	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	res := tx.Where("email = ? AND deleted_at IS NULL", req.Email).FirstOrCreate(request)
	if res.Error != nil {
		return nil, fmt.Errorf("failed to register user: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return nil, errors.New("user already exists")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	marshaledData, err := json.Marshal(dto.UserResponse{
		Id:       request.Id,
		Name:     request.Name,
		Username: request.Username,
		Email:    request.Email,
	})
	// TODO: provide more generic error message
	if err != nil {
		return nil, errors.New("failed to marshal user data")
	}

	otp, err := util.GenerateRandomInteger(OTP_LENGTH)
	if err != nil {
		return nil, errors.New("failed to generate otp")
	}

	// Store OTP and reference data in Redis using Pipeline (single round-trip)
	pipe := u.authRds.Pipeline()
	pipe.Set(ctx, fmt.Sprintf("ref-%s", referenceId), marshaledData, OTP_EXPIRY)
	pipe.Set(ctx, fmt.Sprintf("otp-%s", referenceId), otp, OTP_EXPIRY)
	if _, err = pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to store registration data: %w", err)
	}

	// TODO: send email verification
	go func() {
		if err := u.mailer.SendOTP(request.Email, "Verify Your Email", fmt.Sprintf("Please verify your email address %s", otp)); err != nil {
			fmt.Printf("failed to send otp to %s: %v", request.Email, err)
		}
	}()

	return &dto.RegisterResponse{
		ReferenceId: referenceId,
	}, nil
}

func (u *UserService) VerifyEmail(ctx context.Context, req *dto.VerifyEmailRequest) (*dto.VerifyEmailResponse, error) {
	trimmedReferenceID := strings.TrimSpace(req.ReferenceId)

	// otp, err := u.authRds.Get(ctx, fmt.Sprintf("otp-%s", referenceId)).Result()
	// if err == redis.Nil {
	// 	return nil, errors.New("otp not found")
	// } else if err != nil {
	// 	return nil, errors.New("failed to check otp mapping")
	// }

	// if otp != req.Otp {
	// 	return nil, errors.New("otp not match")
	// }

	if err := u.verifyOTP(ctx, trimmedReferenceID, req.Otp); err != nil {
		return nil, err
	}

	data, err := u.authRds.Get(ctx, fmt.Sprintf("ref-%s", trimmedReferenceID)).Result()
	if err == redis.Nil {
		return nil, errors.New("reference id not found")
	} else if err != nil {
		return nil, errors.New("failed to get reference mappings")
	}

	var user dto.UserResponse
	err = json.Unmarshal([]byte(data), &user)
	if err != nil {
		return nil, errors.New("failed to unmarshal user data")
	}

	var userModel models.User
	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()
	res := tx.Where("id = ? AND deleted_at IS NULL", user.Id).First(&userModel)
	if res.Error != nil {
		return nil, fmt.Errorf("failed to get user: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return nil, errors.New("user not found")
	}

	userModel.IsVerified = true
	userModel.UpdatedAt = time.Now()
	res = tx.Updates(&userModel)
	if res.Error != nil {
		return nil, errors.New("failed to verify email")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &dto.VerifyEmailResponse{
		ReferenceId: trimmedReferenceID,
	}, nil
}

func (u *UserService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.UserAutheticateResponse, error) {
	var userModel models.User

	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	res := tx.Where("email = ? AND deleted_at IS NULL", req.Email).First(&userModel)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, errors.New("invalid email or password")
		}
		return nil, fmt.Errorf("failed to get user: %w", res.Error)
	}

	err := bcrypt.CompareHashAndPassword([]byte(userModel.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !userModel.IsVerified {
		return nil, errors.New("email is not verified")
	}

	sessionID := ulid.Make().String()

	accessToken, err := u.jwt.CreateAccessToken(userModel.Id, sessionID, 15*time.Minute)
	if err != nil {
		return nil, errors.New("failed to create access token")
	}

	refreshToken, err := u.jwt.GenerateRefreshToken()
	if err != nil {
		return nil, errors.New("failed to create refresh token")
	}

	// Hash refresh token before saving to database
	hashedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash refresh token")
	}

	session := models.UserSession{
		SessionID:    sessionID,
		UserID:       userModel.Id,
		RefreshToken: string(hashedRefreshToken),
		ClientIP:     "",                                 // TODO: Get from context/request
		UserAgent:    "",                                 // TODO: Get from context/request
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 Days
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, errors.New("failed to marshal session")
	}

	err = u.authRds.Set(ctx, fmt.Sprintf("session:%s", sessionID), sessionJSON, 7*24*time.Hour).Err()
	if err != nil {
		return nil, errors.New("failed to store session in redis")
	}

	userModel.LastLogin = time.Now()
	if err := tx.Save(&userModel).Error; err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return &dto.UserAutheticateResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		User: dto.UserResponse{
			Id:       userModel.Id,
			Name:     userModel.Name,
			Username: userModel.Username,
			Email:    userModel.Email,
		},
	}, nil
}

func (u *UserService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.UserAutheticateResponse, error) {
	sessionJSON, err := u.authRds.Get(ctx, fmt.Sprintf("session:%s", req.SessionID)).Result()
	if err == redis.Nil {
		return nil, errors.New("invalid or expired session")
	} else if err != nil {
		return nil, errors.New("failed to retrieve session")
	}

	var session models.UserSession
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		return nil, errors.New("failed to unmarshal session")
	}

	// Verify the refresh token matches
	err = bcrypt.CompareHashAndPassword([]byte(session.RefreshToken), []byte(req.RefreshToken))
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if time.Now().After(session.ExpiresAt) {
		_ = u.authRds.Del(ctx, fmt.Sprintf("session:%s", req.SessionID))
		return nil, errors.New("refresh token expired")
	}

	// Get user details
	var userModel models.User
	res := u.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", session.UserID).First(&userModel)
	if res.Error != nil {
		return nil, errors.New("user not found")
	}

	// ---- TOKEN ROTATION ----

	// Delete old session
	_ = u.authRds.Del(ctx, fmt.Sprintf("session:%s", req.SessionID))

	// Generate new session & tokens
	newSessionID := ulid.Make().String()

	newAccessToken, err := u.jwt.CreateAccessToken(userModel.Id, newSessionID, 15*time.Minute)
	if err != nil {
		return nil, errors.New("failed to create access token")
	}

	newRefreshToken, err := u.jwt.GenerateRefreshToken()
	if err != nil {
		return nil, errors.New("failed to create refresh token")
	}

	hashedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(newRefreshToken), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash refresh token")
	}

	newSession := models.UserSession{
		SessionID:    newSessionID,
		UserID:       userModel.Id,
		RefreshToken: string(hashedRefreshToken),
		ClientIP:     session.ClientIP, // preserve original client info
		UserAgent:    session.UserAgent,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}

	newSessionJSON, err := json.Marshal(newSession)
	if err != nil {
		return nil, errors.New("failed to marshal new session")
	}

	err = u.authRds.Set(ctx, fmt.Sprintf("session:%s", newSessionID), newSessionJSON, 7*24*time.Hour).Err()
	if err != nil {
		return nil, errors.New("failed to store new session in redis")
	}

	return &dto.UserAutheticateResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		User: dto.UserResponse{
			Id:       userModel.Id,
			Name:     userModel.Name,
			Username: userModel.Username,
			Email:    userModel.Email,
		},
	}, nil
}

func (u *UserService) Logout(ctx context.Context, sessionID string) error {
	err := u.authRds.Del(ctx, fmt.Sprintf("session:%s", sessionID)).Err()
	if err != nil {
		return errors.New("failed to delete session")
	}
	return nil
}

func (u *UserService) GetProfile(ctx context.Context) (*dto.UserResponse, error) {
	var userModel models.User
	id := ctx.Value(contextkey.UserID).(string)
	if err := u.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&userModel).Error; err != nil {
		return nil, errors.New("user not found")
	}
	return &dto.UserResponse{
		Id:       userModel.Id,
		Name:     userModel.Name,
		Username: userModel.Username,
		Email:    userModel.Email,
	}, nil
}

func (u *UserService) UpdateProfile(ctx context.Context, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	id := ctx.Value(contextkey.UserID).(string)
	currentTime := time.Now()

	var request *models.User
	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&request).Error; err != nil {
		return nil, errors.New("user not found")
	}

	request.Name = req.Name
	request.UpdatedAt = currentTime

	if err := tx.Updates(&request).Error; err != nil {
		return nil, errors.New("failed to update user")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("failed to commit")
	}

	return &dto.UserResponse{
		Id:       request.Id,
		Name:     request.Name,
		Username: request.Username,
		Email:    request.Email,
	}, nil

}

func (u *UserService) UpdatePassword(ctx context.Context, req *dto.UpdatePasswordRequest) (*dto.UserResponse, error) {
	currentTime := time.Now()
	id := ctx.Value(contextkey.UserID).(string)
	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	var userModel models.User
	if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&userModel).Error; err != nil {
		return nil, errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userModel.Password), []byte(req.CurrentPassword)); err != nil {
		return nil, errors.New("invalid current password")
	}

	newhasedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	userModel.Password = string(newhasedPassword)
	userModel.UpdatedAt = currentTime

	if err := tx.Updates(&userModel).Error; err != nil {
		return nil, errors.New("failed to update password")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("failed to commit")
	}

	return &dto.UserResponse{
		Id:       userModel.Id,
		Name:     userModel.Name,
		Username: userModel.Username,
		Email:    userModel.Email,
	}, nil

}

func (u *UserService) UpdateEmail(ctx context.Context, req *dto.UpdateEmailRequest) (*dto.UpdateEmailResponse, error) {
	id := ctx.Value(contextkey.UserID).(string)
	identifier := base64.StdEncoding.EncodeToString([]byte(id))
	referenceId := fmt.Sprintf("%s-%s", identifier, ulid.Make().String())

	var count int64
	if err := u.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", req.NewEmail).Count(&count).Error; err != nil {
		return nil, errors.New("failed to check email")
	}

	if count > 0 {
		return nil, errors.New("email already exists")
	}

	otp, err := util.GenerateRandomInteger(OTP_LENGTH)
	if err != nil {
		return nil, errors.New("failed to generate otp")
	}

	data := map[string]interface{}{
		"user_id": id,
		"email":   req.NewEmail,
	}

	marshaledData, err := json.Marshal(data)
	if err != nil {
		return nil, errors.New("failed to marshal user payload")
	}

	pipe := u.authRds.Pipeline()
	pipe.Set(ctx, fmt.Sprintf("ref-%s", referenceId), marshaledData, OTP_EXPIRY)
	pipe.Set(ctx, fmt.Sprintf("otp-%s", referenceId), otp, OTP_EXPIRY)
	if _, err = pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to store registration data: %w", err)
	}

	go func() {
		if err := u.mailer.SendOTP(req.NewEmail, "Update Email", fmt.Sprintf("Please verify your email address %s", otp)); err != nil {
			fmt.Printf("failed to send otp to %s: %v", req.NewEmail, err)
		}
	}()

	return &dto.UpdateEmailResponse{ReferenceId: referenceId}, nil
}

func (u *UserService) CommitUpdateEmail(ctx context.Context, req *dto.CommitUpdateEmailRequest) (*dto.UserResponse, error) {
	id := ctx.Value(contextkey.UserID).(string)

	if err := u.verifyOTP(ctx, req.ReferenceId, req.Otp); err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := u.authRds.Get(ctx, fmt.Sprintf("ref-%s", req.ReferenceId)).Scan(&data); err != nil {
		return nil, errors.New("failed to get user data")
	}

	if data["user_id"].(string) != id {
		return nil, errors.New("user not found")
	}

	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	var userModel models.User
	if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&userModel).Error; err != nil {
		return nil, errors.New("user not found")
	}

	userModel.Email = data["email"].(string)
	userModel.UpdatedAt = time.Now()

	if err := tx.Updates(&userModel).Error; err != nil {
		return nil, errors.New("failed to update user")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("failed to commit update")
	}

	if err := u.authRds.Del(ctx, fmt.Sprintf("ref-%s", req.ReferenceId)).Err(); err != nil {
		return nil, errors.New("failed to delete ref")
	}

	return &dto.UserResponse{
		Id:       userModel.Id,
		Name:     userModel.Name,
		Username: userModel.Username,
		Email:    userModel.Email,
	}, nil
}

func (u *UserService) DeleteAccount(ctx context.Context, req *dto.DeleteAccountRequest) error {
	id := ctx.Value(contextkey.UserID).(string)
	currentTime := time.Now()

	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	var userModel models.User
	if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&userModel).Error; err != nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userModel.Password), []byte(req.CurrentPassword)); err != nil {
		return errors.New("invalid current password")
	}

	userModel.DeletedAt = sql.NullTime{Time: currentTime, Valid: true}

	if err := tx.Updates(&userModel).Error; err != nil {
		return errors.New("failed to delete user")
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("failed to commit")
	}

	return nil
}

func (u *UserService) verifyOTP(ctx context.Context, reference string, otp string) error {
	trimmedReference := strings.TrimSpace(reference)
	trimmedOTP := strings.TrimSpace(otp)

	otp, err := u.authRds.Get(ctx, fmt.Sprintf("otp-%s", trimmedReference)).Result()
	if err == redis.Nil {
		return errors.New("otp not found")
	} else if err != nil {
		return errors.New("failed to check otp mapping")
	}

	if otp != trimmedOTP {
		return errors.New("otp not match")
	}

	if err := u.authRds.Del(ctx, fmt.Sprintf("otp-%s", trimmedReference)).Err(); err != nil {
		return errors.New("failed to delete otp")
	}

	return nil
}
