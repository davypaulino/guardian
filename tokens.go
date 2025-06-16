package main

import (
	"fmt"
	"github.com/google/uuid"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateTokens(user User) (string, string, error) {

	accessClaims := jwt.MapClaims{
		"sub":        user.ID,
		"role":       user.Role,
		"nickname":   user.NickName,
		"status":     user.Status,
		"exp":        time.Now().Add(time.Hour).UTC().Unix(),
		"iat":        time.Now().UTC().Unix(),
		"jti":        uuid.New().String(),
		"token_type": "access",
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccessToken, err := accessToken.SignedString([]byte(environments.AccessTokenSecret))
	if err != nil {
		return "", "", err
	}

	refreshClaims := jwt.MapClaims{
		"sub":        user.ID,
		"exp":        time.Now().Add(time.Hour * 24 * 7).UTC().Unix(),
		"iat":        time.Now().UTC().Unix(),
		"jti":        uuid.New().String(),
		"token_type": "refresh",
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(environments.RefreshTokenSecret))
	if err != nil {
		return "", "", err
	}

	return signedAccessToken, signedRefreshToken, nil
}

type TokenType string

const (
	Access  TokenType = "accessToken"
	Refresh TokenType = "refreshToken"
)

func ValidateToken(tokenString string, tt TokenType) (*jwt.MapClaims, error) {
	var secret = []byte(environments.AccessTokenSecret)
	if tt == Refresh {
		secret = []byte(environments.RefreshTokenSecret)
	}

	if len(secret) == 0 {
		return nil, fmt.Errorf("missing secret key")
	}

	parseOptions := []jwt.ParserOption{
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(5 * time.Second),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	}, parseOptions...)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid claims")
	}
	return &claims, nil
}
