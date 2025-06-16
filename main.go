package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	urlp "net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	httptrace "github.com/DataDog/dd-trace-go/contrib/net/http/v2"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/google/uuid"
	"github.com/markbates/goth/gothic"
)

var gAppAllowedHosts = []string{
	"http://localhost:3000/",
	"https://localhost:8443/",
}

var gClientsSessions = struct {
	sync.Mutex
	data map[string]struct {
		RedirectURL string
		ExpiresAt   time.Time
	}
}{
	data: make(map[string]struct {
		RedirectURL string
		ExpiresAt   time.Time
	}),
}

func initAppHosts() {
	allowedHostsStr := os.Getenv("ALLOWED_FRONTEND_HOSTS")
	if allowedHostsStr != "" {
		gAppAllowedHosts = strings.Split(allowedHostsStr, ",")
	}

	go cleanupExpiredSessions()
}

func cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		gClientsSessions.Lock()
		for state, entry := range gClientsSessions.data {
			if time.Now().After(entry.ExpiresAt) {
				delete(gClientsSessions.data, state)
			}
		}
		gClientsSessions.Unlock()
		logger.Debug("Cleaned up expired client sessions.")
	}
}

func isHostAllowed(host string) bool {
	for _, allowedHost := range gAppAllowedHosts {
		if allowedHost == host {
			return true
		}
	}
	return false
}

var db *sql.DB
var logger *zap.Logger
var environments *Environment
var providerIndex *ProviderIndex

func initLogger() {
	logger, _ = zap.NewProduction()
	defer logger.Sync()
}

type UserRequest struct {
	ID        uuid.UUID `json:"userId"`
	Nickname  string    `json:"nickname"`
	AvatarURL string    `json:"avatar_url"`
	Terms     bool      `json:"terms_accepted"`
}

type UserTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserTokenRequest struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
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

		token, refresh, err := UpdateUserRegister(newUser)
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

func putRenewTokens(w http.ResponseWriter, r *http.Request) {
	correlationId := r.Header.Get("X-Correlation-Id")
	method := "putRenewTokens"

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

		var tokens UserTokenRequest
		err = json.Unmarshal(body, &tokens)
		if err != nil {
			logger.Warn("Error on Convert Body", zap.String("method", method), zap.Error(err))
			http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
			return
		}

		response, err := RenewAccessToken(tokens.AccessToken, tokens.RefreshToken)
		if err != nil {
			logger.Warn("Error on Convert Body", zap.String("method", method), zap.Error(err))
			http.Error(w, "Token invalido", http.StatusBadRequest)
			return
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
		ID        string `json:"id"`
		Name      string `json:"nickname"`
		Terms     string `json:"accepted_terms"`
		AvatarURL string `json:"img_url"`
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
	referer := req.Header.Get("Referer")

	if !isHostAllowed(referer) {
		logger.Warn("Unauthorized host attempting authentication.",
			zap.String("referer", referer))
		res.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(res, "Unauthorized request origin.")
		return
	}

	if _, err := gothic.CompleteUserAuth(res, req); err == nil {
		callbackHandler(res, req)
	} else {
		// gothic.BeginAuthHandler(res, req)
		urlStr, err := gothic.GetAuthURL(res, req)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(res, err)
			return
		}

		parsedURL, parseErr := urlp.Parse(urlStr)
		if parseErr != nil {
			logger.Error("Failed to parse Auth URL string", zap.Error(parseErr), zap.String("url", urlStr))
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(res, "Internal server error during URL parsing.")
			return
		}

		queryParams := parsedURL.Query()
		state := queryParams.Get("state")

		gClientsSessions.Lock()
		gClientsSessions.data[state] = struct {
			RedirectURL string
			ExpiresAt   time.Time
		}{
			RedirectURL: referer,
			ExpiresAt:   time.Now().Add(5 * time.Minute),
		}
		gClientsSessions.Unlock()

		logger.Info("Auth Flow Initiated",
			zap.String("GothAuthURL", urlStr),
			zap.String("GothState", state),
			zap.String("FrontendReturnURL", referer),
			zap.String("RequestHost", req.Host),
			zap.String("RequestURI", req.RequestURI),
		)

		http.Redirect(res, req, urlStr, http.StatusTemporaryRedirect)
	}
}

func main() {
	initLogger()
	environments = initEnvironments()
	tracer.Start(
		tracer.WithEnv(environments.DatadogSettings.ServiceEnvironment),
		tracer.WithService(environments.DatadogSettings.ServiceName),
		tracer.WithHostname(environments.DatadogSettings.AgentHost),
		tracer.WithServiceVersion(environments.DatadogSettings.Version),
		tracer.WithAgentURL(environments.DatadogSettings.TraceAgentHostname),
		tracer.WithGlobalTag("team", "pung-guardian"),
	)

	defer tracer.Stop()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	go func() {
		<-sigChan
		tracer.Stop()
	}()

	db, _ = initDatabase()
	providerIndex = initProviders()

	prefix := "/api/v1/guardian"

	apiMux := httptrace.NewServeMux(httptrace.WithService("pung-guardian"))

	apiMux.HandleFunc(prefix+"/auth/{provider}/callback",
		configMiddlewares(callbackHandler, corsMiddleware))

	apiMux.HandleFunc(prefix+"/logout/{provider}",
		configMiddlewares(logoutHandler, corsMiddleware))

	apiMux.HandleFunc(prefix+"/auth/{provider}",
		configMiddlewares(providerAuthHandler, corsMiddleware))

	apiMux.HandleFunc(prefix+"/auth/refresh",
		configMiddlewares(putRenewTokens, corsMiddleware))

	apiMux.HandleFunc(prefix+"/users",
		configMiddlewares(getUserInfo, corsMiddleware, authMiddleware))

	apiMux.HandleFunc(prefix+"/register",
		configMiddlewares(postUserRegister, corsMiddleware, authMiddleware))

	logger.Info("Starting server", zap.String("port", environments.ServerPort))
	log.Fatal(http.ListenAndServe(":"+environments.ServerPort, apiMux))
}
