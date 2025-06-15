package main

import (
	"fmt"
	"go.uber.org/zap"
	"os"
)

type AuthProviders struct {
	ProviderKeys    map[string]string
	ProviderSecrets map[string]string
}

type DatadogSettings struct {
	AgentHost          string
	TraceAgentHostname string
	ServiceName        string
	ServiceEnvironment string
	Version            string
}

type Environment struct {
	RedirectUrl        string
	Auths              AuthProviders
	SessionSecret      string
	DatabaseConn       string
	ServerPort         string
	MetricsPort        string
	AccessTokenSecret  string
	RefreshTokenSecret string
	DatadogSettings    DatadogSettings
}

func checkEnvVariable(label string) string {
	env := os.Getenv(label)
	if env == "" {
		logger.Error("Setup Project Error | Not found variable",
			zap.String(label, env))
		os.Exit(1)
	}
	return env
}

func initEnvironments() *Environment {
	redirectUrl := checkEnvVariable("REDIRECT_URL")
	githubKey := checkEnvVariable("AUTH_GITHUB_KEY")
	googleKey := checkEnvVariable("AUTH_GOOGLE_KEY")
	githubSecret := checkEnvVariable("AUTH_GITHUB_SECRET")
	googleSecret := checkEnvVariable("AUTH_GOOGLE_SECRET")
	sessionSecret := checkEnvVariable("SESSION_SECRET")
	accessTokenSecret := checkEnvVariable("ACCESS_TOKEN_SECRET")
	refreshTokenSecret := checkEnvVariable("REFRESH_TOKEN_SECRET")
	dbHost := checkEnvVariable("DB_HOST")
	dbPort := checkEnvVariable("DB_PORT")
	dbUser := checkEnvVariable("DB_USER")
	dbPassword := checkEnvVariable("DB_PASSWORD")
	dbName := checkEnvVariable("DB_NAME")
	serverPort := checkEnvVariable("SERVER_PORT")
	metricsPort := checkEnvVariable("METRICS_PORT")

	datadogSettings := DatadogSettings{
		AgentHost:          checkEnvVariable("DD_AGENT_HOST"),
		TraceAgentHostname: checkEnvVariable("DD_TRACE_AGENT_HOSTNAME"),
		ServiceName:        checkEnvVariable("DD_SERVICE"),
		ServiceEnvironment: checkEnvVariable("DD_ENV"),
		Version:            checkEnvVariable("DD_VERSION"),
	}

	databaseString := fmt.Sprintf(`host=%s port=%s user=%s password=%s dbname=%s sslmode=%s`,
		dbHost, dbPort, dbUser, dbPassword, dbName, "disable")

	envs := Environment{
		RedirectUrl: redirectUrl,
		Auths: AuthProviders{
			ProviderKeys: map[string]string{
				"github": githubKey,
				"google": googleKey,
			},
			ProviderSecrets: map[string]string{
				"github": githubSecret,
				"google": googleSecret,
			},
		},
		SessionSecret:      sessionSecret,
		DatabaseConn:       databaseString,
		ServerPort:         serverPort,
		MetricsPort:        metricsPort,
		AccessTokenSecret:  accessTokenSecret,
		RefreshTokenSecret: refreshTokenSecret,
		DatadogSettings:    datadogSettings,
	}

	logger.Info("Environment variables loaded successfully.")
	return &envs
}
