package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/markbates/goth/gothic"
	"go.uber.org/zap"
)

var db *sql.DB

var logger *zap.Logger

func initLogger() {
    logger, _ = zap.NewProduction()
    defer logger.Sync()
}

type UserRequest struct {
    ID        	uuid.UUID 	`json:"userId"`
    Nickname  	string 		`json:"nickname"`
    AvatarURL 	string 		`json:"avatar_url"`
	Terms		bool		`json:"terms_accepted"`
}

type UserTokenResponse struct {
	AccessToken 	string 	`json:"access_token"`
	RefreshToken 	string 	`json:"refresh_token"`
}

func postUserRegister(w http.ResponseWriter, r *http.Request) {
	correlationId := r.Header.Get("X-Correlation-Id")
	method := "postUserRegister"

	if r.Method == http.MethodOptions {
		logger.Info("Starting Process", zap.String("http:method", r.Method), zap.String("method", method), zap.String("correlation_id", correlationId))
		defer logger.Info("Finished Process", zap.String("http:method", r.Method), zap.String("method", method), zap.String("correlation_id", correlationId))
	
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodPut {
		logger.Info("Starting Process", zap.String("http:method", r.Method), zap.String("method", method), zap.String("correlation_id", correlationId))
		defer logger.Info("Finished Process", zap.String("http:method", r.Method), zap.String("method", method), zap.String("correlation_id", correlationId))
	
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Warn("Error on Read Body", zap.String("method", method), zap.Error(err))
			http.Error(w, "Erro ao ler o corpo da requisição", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		
		var user UserRequest
		err = json.Unmarshal(body, &user)
		if err != nil {
			logger.Warn("Error on Convert Body", zap.String("method", method), zap.Error(err))
			http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
			return
		}
		
		if !user.Terms {
			logger.Warn("Error Terms no assign", zap.String("method", method), zap.Error(err))
			http.Error(w, "Termos não aceitos", http.StatusBadRequest)
			return
		}

		newUser, err := GetUserByUserId(user.ID)
		newUser.NickName = user.Nickname
		newUser.ImgURL = user.AvatarURL
		newUser.Terms = user.Terms
		newUser.Status = Active

		token, refresh, err := GenerateTokens(newUser)
		if err != nil {
			logger.Error("Error on Generate Tokens", zap.String("method", method), zap.Error(err))
			http.Error(w, "Erro ao gerar tokens", http.StatusInternalServerError)
			return
		}

		response := UserTokenResponse{
			AccessToken:  token,
			RefreshToken: refresh,
		}
	
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func getUserInfo(w http.ResponseWriter, r *http.Request) {
	correlationId := r.Header.Get("X-Correlation-Id")
	userId := r.URL.Query().Get("userId")
	logger.Info("Starting | Get User Info", zap.String("userId", userId), zap.String("correlation_id", correlationId))
	defer logger.Info("Finished | Get User Info", zap.String("correlation_id", correlationId))
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	var user struct {
		ID			string	`json:"id"`
		Name      	string 	`json:"nickname"`
		Terms     	string 	`json:"accepted_terms"`
		AvatarURL 	string 	`json:"img_url"`
	}
	user.ID = userId
	if userId == "" || userId == "undefined" {
		logger.Warn("User Id Param not found.")
		http.Error(w, "Not found User Id query param", http.StatusBadRequest)
		return
	}

	err := db.QueryRow(`
        SELECT id, nickname, terms_accepted, avatar_url 
        FROM users WHERE id = $1`, user.ID).Scan(&user.ID, &user.Name, &user.Terms, &user.AvatarURL)
	
	if err == sql.ErrNoRows {
		logger.Error("Error on db", zap.Error(err))
		http.Error(w, "User not found or invalid token", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		logger.Error("Database query error", zap.Error(err), zap.String("correlation_id", correlationId))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func logoutHandler(res http.ResponseWriter, req *http.Request) {
	gothic.Logout(res, req)
	res.Header().Set("Location", "/")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func providerAuthHandler(res http.ResponseWriter, req *http.Request) {

	// try to get the user without re-authenticating
	if _, err := gothic.CompleteUserAuth(res, req); err == nil {
		callbackHandler(res, req)
	} else {
		gothic.BeginAuthHandler(res, req)
	}
}

func main() {
	Init()
	initLogger()
	fmt.Println("SESSION_SECRET:", os.Getenv("SESSION_SECRET"))

	http.HandleFunc("/auth/{provider}/callback",
		configMiddlewares(callbackHandler, corsMiddleware))
	http.HandleFunc("/logout/{provider}",
		configMiddlewares(logoutHandler, corsMiddleware))
	http.HandleFunc("/auth/{provider}",
		configMiddlewares(providerAuthHandler, corsMiddleware))

	http.HandleFunc("/users/1", getUserInfo)
	http.HandleFunc("/users", getUserInfo)
	http.HandleFunc("/register",
		configMiddlewares(postUserRegister,
			corsMiddleware,
			authMiddleware))

	log.Println("listening on localhost:3001")
	log.Fatal(http.ListenAndServe(":3001", nil))
}
