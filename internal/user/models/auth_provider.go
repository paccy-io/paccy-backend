package models

type AuthProvider string

const (
	ProviderEmail  AuthProvider = "email"
	ProviderGoogle AuthProvider = "google"
	ProviderGitHub AuthProvider = "github"
)

func IsValidAuthProvider(p string) bool {
	switch AuthProvider(p) {
	case ProviderEmail, ProviderGoogle, ProviderGitHub:
		return true
	default:
		return false
	}
}