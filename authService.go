package main

import "github.com/markbates/goth"

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