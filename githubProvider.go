package main

import (
	"github.com/google/uuid"
	"github.com/markbates/goth"
)

type UserStatus int8

const (
	// 1 ~ 29 User Active 	
    Active   UserStatus = 1
	// 40 ~ 49 warning
	Pending UserStatus = 40
	// 50 ~ 59 error
    Inactive UserStatus = 50
    Suspended UserStatus = 51
)

type UserRole int8

const (
	NormalUser UserRole = iota
	GmUser
	Admin
)

type User struct {
    ID              		uuid.UUID 	`json:"id"`
    NickName        		string    	`json:"nickname"`
    Email	        		*string   	`json:"email"`
    ImgURL          		string    	`json:"img_url"`
	AccessToken     		*string    	`json:"access_token"`
    RefreshToken    		*string   	`json:"refresh_token"`
	Status					UserStatus	`json:"status"`
	Role					UserRole	`json:"role"`
    Provider        		string    	`json:"provider"`
    ProviderUserID  		string    	`json:"provider_user_id"`
    ProviderAccessToken     string    	`json:"provider_access_token"`
    ProviderRefreshToken    *string   	`json:"provider_refresh_token"`
	Terms					bool		`json:"terms_accepted"`
}

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