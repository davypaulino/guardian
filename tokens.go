package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte(os.Getenv("ACCESS_TOKEN_SECRET"))    
var refreshSecretKey = []byte("REFRESH_TOKEN_SECRET")

func GenerateTokens(user User) (string, string, error) {

    accessClaims := jwt.MapClaims{
        "sub":  user.ID,
        "role": user.Role,
		"nickname": user.NickName,
		"status": user.Status,
        "exp":  time.Now().Add(time.Hour).Unix(),
        "iat":  time.Now().Unix(),
    }
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    signedAccessToken, err := accessToken.SignedString(secretKey)
    if err != nil {
        return "", "", err
    }

    refreshClaims := jwt.MapClaims{
        "sub": user.ID,
        "exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
        "iat": time.Now().Unix(),
    }
    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    signedRefreshToken, err := refreshToken.SignedString(refreshSecretKey)

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
    var secret = secretKey
    if (tt == Refresh) {
        secret = refreshSecretKey
    }

    if len(secret) == 0 {
        return nil, fmt.Errorf("missing secret key")
    }

    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return secret, nil
    })

    if err != nil {
        return nil, fmt.Errorf("invalid token: %v", err)
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        return nil, fmt.Errorf("invalid claims")
    }
    return &claims, nil
}
