package main

func UpdateUserTokenss() {
	token, refresh, _ := GenerateTokens(user)
	user.AccessToken = &token
	user.RefreshToken = &refresh

	_, err := db.Exec(`
	UPDATE users SET 
		access_token = $1,
		refresh_token = $2,
		updated_at = NOW()
	WHERE id = $3`,
	user.AccessToken, user.RefreshToken, user.ID)
}