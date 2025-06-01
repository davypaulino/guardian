SRC := 	main.go \
		authProviders.go \
		database.go \
		githubProvider.go \
		tokens.go

all:
	go run $(SRC)