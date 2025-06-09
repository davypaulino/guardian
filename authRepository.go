package main

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func UpdateUserTokens(accessToken, refreshToken string, userId uuid.UUID) error {
	_, err := db.Exec(`
		UPDATE users SET 
			access_token = $1,
			refresh_token = $2,
			updated_at = NOW()
		WHERE id = $3`,
		accessToken, refreshToken, userId)

	if err != nil {
		logger.Error("Error on Update Tokens")
		return err
	}

	return nil
}

func GetUserByProviderId(providerUserId string) (User, error) {
	var user User

	err := db.QueryRow(`
	SELECT id, nickname, email, avatar_url,
		access_token, refresh_token, status,
		role, terms_accepted
	FROM users
	WHERE provider_user_id = $1`,
	providerUserId).Scan(&user.ID, &user.NickName, &user.Email, &user.ImgURL,
		&user.AccessToken, &user.RefreshToken, &user.Status,
		&user.Role, &user.Terms)

	if err != nil {
		logger.Error("Error on get user by Provider", zap.Error(err))
		return User{}, err
	}

	return user, nil
}

func GetUserByUserId(userId uuid.UUID) (User, error) {
	var user User

	err := db.QueryRow(`
	SELECT id, nickname, email, avatar_url,
		access_token, refresh_token, status,
		role, terms_accepted
	FROM users
	WHERE id = $1`,
	userId).Scan(&user.ID, &user.NickName, &user.Email, &user.ImgURL,
		&user.AccessToken, &user.RefreshToken, &user.Status,
		&user.Role, &user.Terms)

	if err != nil {
		logger.Error("Error on get user by Provider", zap.Error(err))
		return User{}, err
	}

	return user, nil
}

func CreateUserOrUpdateProviderTokens(user User) error {
	_, err := db.Exec(`
		INSERT INTO users (id, provider, provider_user_id, nickname,
			email, avatar_url, provider_access_token,
			provider_refresh_token, updated_at, status, "role", terms_accepted) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (provider_user_id) DO UPDATE SET 
			provider_access_token = $7,
			provider_refresh_token = $8,
			updated_at = NOW();`,
	user.ID, user.Provider, user.ProviderUserID,
	user.NickName, user.Email, user.ImgURL,
	user.ProviderAccessToken, user.ProviderRefreshToken,
	nil, user.Status, user.Role, user.Terms)

	if err != nil {
		logger.Error("Error on create user or update provider", zap.Error(err))
		return err
	}

	return nil
}

func UpdateUserRegister(user User) (string, string, error) {
	
	token, refresh, err := GenerateTokens(user)
	if err != nil {
		logger.Error("Error on Generate Tokens", zap.Error(err))
		return "", "", err
	}

	_, err = db.Exec(`
		UPDATE users SET 
			nickname = $1,
			avatar_url = $2,
			terms_accepted = $3,
			status = $4,
			access_token = $5,
			refresh_token = $6,
			updated_at = NOW()
		WHERE id = $7`,
		&user.NickName,
		&user.ImgURL,
		&user.Terms,
		&user.Status,
		&token,
		&refresh,
		&user.ID)

	if err != nil {
		logger.Error("Error on Update User", zap.Error(err))
		return "", "", err
	}

	return token, refresh, nil
}