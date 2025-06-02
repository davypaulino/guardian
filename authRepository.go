package main

import "github.com/google/uuid"

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
		logger.Error("Error on get user by Provider")
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
		logger.Error("Error on get user by Provider")
		return User{}, err
	}

	return user, nil
}

