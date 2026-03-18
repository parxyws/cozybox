package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	golangjwt "github.com/golang-jwt/jwt/v5"
)

const (
	ISS = "cozybox"
)

type TokenMaker interface {
	CreateAccessToken(userID string, sessionID string, duration time.Duration) (string, error)
	VerifyAccessToken(token string) (*CustomClaims, error)
	GenerateRefreshToken() (string, error)
}

type JWTMaker struct {
	secretKey []byte
}

type CustomClaims struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	golangjwt.RegisteredClaims
}

func NewJWTMaker(secretKey string) (TokenMaker, error) {
	if len(secretKey) < 32 {
		return nil, errors.New("invalid key size: must be at least 32 characters")
	}
	return &JWTMaker{
		secretKey: []byte(secretKey),
	}, nil
}

func (maker *JWTMaker) CreateAccessToken(userID string, sessionID string, duration time.Duration) (string, error) {
	claims := CustomClaims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: golangjwt.RegisteredClaims{
			Issuer:    ISS,
			Subject:   userID,
			Audience:  []string{"cozybox-api"},
			ExpiresAt: golangjwt.NewNumericDate(time.Now().Add(duration)),
			NotBefore: golangjwt.NewNumericDate(time.Now()),
			IssuedAt:  golangjwt.NewNumericDate(time.Now()),
			ID:        sessionID,
		},
	}

	jwtToken := golangjwt.NewWithClaims(golangjwt.SigningMethodHS256, claims)
	return jwtToken.SignedString(maker.secretKey)
}

func (maker *JWTMaker) VerifyAccessToken(token string) (*CustomClaims, error) {
	keyFunc := func(token *golangjwt.Token) (interface{}, error) {
		_, ok := token.Method.(*golangjwt.SigningMethodHMAC)
		if !ok {
			return nil, errors.New("invalid authentication token marker")
		}
		return maker.secretKey, nil
	}

	jwtToken, err := golangjwt.ParseWithClaims(token, &CustomClaims{}, keyFunc,
		golangjwt.WithLeeway(5*time.Second),
		golangjwt.WithIssuer(ISS),
		golangjwt.WithAudience("cozybox-api"),
		golangjwt.WithValidMethods([]string{golangjwt.SigningMethodHS256.Alg()}),
		golangjwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}

	claims, ok := jwtToken.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("invalid authentication token marker")
	}

	return claims, nil
}

func (maker *JWTMaker) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
