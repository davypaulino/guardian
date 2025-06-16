package main

import (
	"errors"
	"fmt"
	"github.com/google/uuid"

	"github.com/markbates/goth"
)

var (
	ErrInvalidRefreshToken       = errors.New("invalid refresh token provided")
	ErrRefreshTokenMismatch      = errors.New("refresh token mismatch for user")
	ErrAccessTokenMismatch       = errors.New("access token mismatch for user")
	ErrTokenHasBeenRevokedOrUsed = errors.New("token has been revoked or already used") // If you track JTI or single-use tokens
	ErrUserNotFound              = errors.New("user associated with token not found")
	ErrTokenGenerationFailure    = errors.New("failed to generate new tokens")
	ErrTokenUpdateFailure        = errors.New("failed to update tokens in database")
	ErrRefreshTokenExpired       = errors.New("refresh token has expired") // From jwt_utils.ValidateToken
	ErrAuthenticationFailed      = errors.New("authentication failed")     // General auth error
	ErrUnexpectedTokenValidation = errors.New("unexpected token validation error")
)

func SyncUserProvider(user goth.User) (User, error) {
	newUser := UserAccount[user.Provider](user)
	err := CreateUserOrUpdateProviderTokens(newUser)
	if err != nil {
		return User{}, err
	}

	newUser, err = GetUserByProviderId(newUser.ProviderUserID)
	if err != nil {
		return User{}, err
	}

	token, refresh, _ := GenerateTokens(newUser)
	err = UpdateUserTokens(token, refresh, newUser.ID)
	if err != nil {
		return User{}, err
	}
	newUser.AccessToken = &token
	newUser.RefreshToken = &refresh

	return newUser, nil
}

func RenewAccessToken(oldAccess, oldRefresh string) (UserTokenResponse, error) {
	claims, err := ValidateToken(oldRefresh, Refresh)
	if err != nil {
		return UserTokenResponse{}, fmt.Errorf("%w: %v", ErrInvalidRefreshToken, err)
	}

	userID, ok := (*claims)["sub"].(string)
	if !ok || userID == "" {
		return UserTokenResponse{}, fmt.Errorf("%w: user ID (sub) not found in refresh token claims", ErrInvalidRefreshToken)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return UserTokenResponse{}, fmt.Errorf("%w: invalid user ID format from token: %v", ErrInvalidRefreshToken, err)
	}

	user, err := GetUserByUserId(userUUID)
	if err != nil {
		return UserTokenResponse{}, fmt.Errorf("%w: %v", ErrUserNotFound, err)
	}

	if user.RefreshToken == nil || *user.RefreshToken != oldRefresh {
		return UserTokenResponse{}, ErrRefreshTokenMismatch
	}

	if user.AccessToken == nil || *user.AccessToken != oldAccess {
		return UserTokenResponse{}, ErrAccessTokenMismatch
	}

	access, refresh, err := GenerateTokens(user)
	if err != nil {
		return UserTokenResponse{}, fmt.Errorf("%w: %v", ErrTokenGenerationFailure, err)
	}

	err = UpdateUserTokens(access, refresh, user.ID)
	if err != nil {
		return UserTokenResponse{}, fmt.Errorf("%w: %v", ErrTokenUpdateFailure, err)
	}

	return UserTokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}
