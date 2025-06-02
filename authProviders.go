package main

import (
	"os"
	"sort"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
}

var UserAccount = map[string]func(goth.User) User{
    "github": NewGithubUser,
}

var providerIndex = func () *ProviderIndex {
	goth.UseProviders(
		github.New(os.Getenv("AUTH_GITHUB_KEY"), os.Getenv("AUTH_GITHUB_SECRET"), "http://localhost:3001/auth/github/callback"),
		google.New(os.Getenv("AUTH_GOOGLE_KEY"), os.Getenv("AUTH_GOOGLE_SECRET"), "http://localhost:3001/auth/google/callback"),
	)

	m := map[string]string{
		"github": "Github",
		"google": "Google",
	}

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	providerIndex := &ProviderIndex{Providers: keys, ProvidersMap: m}
	return providerIndex
}()