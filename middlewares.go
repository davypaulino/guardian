package main

import (
	"net/http"

	"go.uber.org/zap"
)

func configMiddlewares(handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		correlationId := r.Header.Get("X-Correlation-Id")
		logger.Info("Starting | Auth ", zap.String("correlation_id", correlationId))
		defer logger.Info("Finished | Auth ", zap.String("correlation_id", correlationId))

		if authHeader == "" {
			logger.Warn("Invalid Token", zap.String("token", authHeader), zap.String("correlation_id", correlationId))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := authHeader[len("Bearer "):]
		claims, err := ValidateToken(token, Access)

		if err != nil {
			logger.Warn("Invalid Token", zap.Error(err), zap.String("correlation_id", correlationId))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		id, err := claims.GetSubject()
		if err != nil {
			logger.Warn("Problem on Get Sub", zap.Any("claims", claims), zap.String("correlation_id", correlationId))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		user := User{}

		err = db.QueryRow(`
		SELECT id, nickname, email, avatar_url,
			access_token, refresh_token, status,
			role, terms_accepted
		FROM users
		WHERE id = $1`,
			id).Scan(&user.ID, &user.NickName, &user.Email, &user.ImgURL,
			&user.AccessToken, &user.RefreshToken, &user.Status,
			&user.Role, &user.Terms)

		if err != nil {
			logger.Error("Error On Database", zap.Error(err), zap.String("correlation_id", correlationId))
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if *user.AccessToken != token {
			logger.Warn("Invalid Token", zap.String("token", token), zap.String("correlation_id", correlationId))
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		correlationId := r.Header.Get("X-Correlation-Id")
		logger.Info("Starting | Cors ", zap.String("correlation_id", correlationId))
		defer logger.Info("Finished | Cors ", zap.String("correlation_id", correlationId))

		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Correlation-Id")
		next.ServeHTTP(w, r)
	}
}
