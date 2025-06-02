SRC := 	main.go \
		authProviders.go \
		database.go \
		githubProvider.go \
		tokens.go \
		middlewares.go \
		authRepository.go \
		authService.go \
		authHandlers.go \
		userModel.go

all:
	go run $(SRC)