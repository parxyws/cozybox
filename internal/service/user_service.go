package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/parxyws/cozybox/internal/dto"
	"github.com/parxyws/cozybox/internal/models"
	"github.com/parxyws/cozybox/internal/tools/jwt"
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
	jwt     jwt.TokenMaker
}

const (
	OTP_LENGTH = 6
	OTP_EXPIRY = 5 * time.Minute
)

type UserServiceInterface interface{}

func NewUserService(db *gorm.DB, mailer *mail.Mailer, authRds *redis.Client, jwt jwt.TokenMaker) *UserService {
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
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	res := tx.Where("email = ? AND deleted_at IS NULL", req.Email).FirstOrCreate(request)
	if res.Error != nil {
		return nil, fmt.Errorf("failed to register user: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return nil, errors.New("user already exists")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
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
		err := u.mailer.SendOTP(request.Email, "Verify Your Email", fmt.Sprintf("Please verify your email address %s", otp))
		if err != nil {
			fmt.Printf("failed to send otp to %s: %v", request.Email, err)
		}
	}()

	return &dto.RegisterResponse{
		ReferenceId: referenceId,
	}, nil
}

func (u *UserService) VerifyEmail(ctx context.Context, req *dto.VerifyEmailRequest) (*dto.VerifyEmailResponse, error) {
	referenceId := strings.TrimSpace(req.ReferenceId)

	otp, err := u.authRds.Get(ctx, fmt.Sprintf("otp-%s", referenceId)).Result()
	if err == redis.Nil {
		return nil, errors.New("otp not found")
	} else if err != nil {
		return nil, errors.New("failed to check otp mapping")
	}

	if otp != req.Otp {
		return nil, errors.New("otp not match")
	}

	data, err := u.authRds.Get(ctx, fmt.Sprintf("ref-%s", referenceId)).Result()
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
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
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
		tx.Rollback()
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &dto.VerifyEmailResponse{
		ReferenceId: referenceId,
	}, nil
}

func (u *UserService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.UserAutheticateResponse, error) {
	var userModel models.User

	tx := u.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

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
		tx.Rollback()
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
	id := ctx.Value("user_id").(string)
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
	return nil, nil
}

func (u *UserService) UpdatePassword(ctx context.Context, req *dto.UpdatePasswordRequest) (*dto.UserResponse, error) {
	return nil, nil
}

func (u *UserService) UpdateEmail(ctx context.Context, req *dto.UpdateEmailRequest) (*dto.UserResponse, error) {
	return nil, nil
}

func (u *UserService) DeleteAccount(ctx context.Context, req *dto.DeleteAccountRequest) error {
	return nil
}
