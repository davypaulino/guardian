SRC := 	main.go \
		authProviders.go \
		database.go \
		githubProvider.go \
		tokens.go \
		middlewares.go

all:
	go run $(SRC)