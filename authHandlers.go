package main

import (
	"fmt"
	"net/http"
	"time"

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

	callbackQueryParams := r.URL.Query()
	stateFromCallback := callbackQueryParams.Get("state")

	gClientsSessions.Lock()
	sessionData, found := gClientsSessions.data[stateFromCallback]
	if !found || time.Now().After(sessionData.ExpiresAt) {
		gClientsSessions.Unlock()
		logger.Error("Callback session state not found or expired.", zap.String("state", stateFromCallback))
		http.Error(w, "Authentication session expired or invalid.", http.StatusUnauthorized)
		return
	}
	delete(gClientsSessions.data, stateFromCallback)
	gClientsSessions.Unlock()

	redirectURL := fmt.Sprintf("%shome?access_token=%s&refresh_token=%s",
		sessionData.RedirectURL, *newUser.AccessToken, *newUser.RefreshToken)

	logger.Info("Redirect to", zap.String("redirectURL", sessionData.RedirectURL))
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
