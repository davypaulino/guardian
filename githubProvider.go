package main

import (
	"github.com/google/uuid"
	"github.com/markbates/goth"
)

func NewGithubUser(user goth.User) User {
	var nickName = user.NickName
	if (nickName == "") {
		if n, ok := user.RawData["login"].(string); ok {
			nickName = n
		}
	}
	
	if (nickName == "") {
		if n, ok := user.RawData["twitter_username"].(string); ok {
			nickName = n;
		}
	}

	var email *string
	if (user.Email != "") {
		email = &user.Email
	}

	if (email == nil) {
		if mail, ok := user.RawData["login"].(string); ok {
			email = &mail;
		}
	}

	var url = user.AvatarURL
	if (url == "") {
		if avatar, ok := user.RawData["avatar_url"].(string); ok {
			url = avatar
		}
	}

	newUser := User{
		ID: uuid.New(),
		Provider: user.Provider,
		ProviderUserID: user.UserID,
		NickName: nickName,
		Email: email,
		ImgURL: url,
		ProviderAccessToken: user.AccessToken,
		ProviderRefreshToken: &user.RefreshToken,
		Status: Pending,
		Role: NormalUser,
		Terms: false,
	}
	return newUser
}