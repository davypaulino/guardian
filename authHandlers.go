package main

import (
	"fmt"
	"net/http"

	"github.com/markbates/goth/gothic"
	"go.uber.org/zap"
)

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		logger.Warn("Error on Complete User Auth | Unauthorized", zap.Error(err))
		http.Error(w, "Authentication failed: ", http.StatusUnauthorized)
		return
	}
	
	if user.AccessToken == "" {
		logger.Error("No Access Token Generated", zap.Error(err))
		http.Error(w, "No access token received", http.StatusInternalServerError)
		return
	}

	newUser, err := SyncUserProvider(user)
	if err != nil {
		http.Error(w, "Error on Create account.", http.StatusInternalServerError)
		return
	}
	
	redirectURL := fmt.Sprintf("http://localhost:3000/home?token=%s", *newUser.AccessToken)	
	http.Redirect(w, r, redirectURL, http.StatusFound)
}