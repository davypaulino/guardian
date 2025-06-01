package main

import (
    "time"
	"os"

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