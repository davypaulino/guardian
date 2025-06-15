package main

import (
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

func initProviders() *ProviderIndex {
	goth.UseProviders(
		github.New(environments.Auths.ProviderKeys["github"], environments.Auths.ProviderSecrets["github"], environments.RedirectUrl+"/api/v1/guardian/auth/github/callback"),
		google.New(environments.Auths.ProviderKeys["google"], environments.Auths.ProviderSecrets["google"], environments.RedirectUrl+"/api/v1/guardian/auth/google/callback"),
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

	return &ProviderIndex{Providers: keys, ProvidersMap: m}
}
