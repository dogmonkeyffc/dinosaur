package metadata

import (
	"errors"
	"strings"
)

var ErrSettings = errors.New("invalid metadata settings")
var ErrMissingCredential = errors.New("TMDB credential is not configured")

type Settings struct {
	CredentialType string `json:"credential_type"`
	Language       string `json:"language"`
	Credential     string `json:"-"`
}

type SettingsUpdate struct {
	CredentialType string `json:"credential_type"`
	Language       string `json:"language"`
	Credential     string `json:"credential"`
}

func DefaultSettings() Settings {
	return Settings{CredentialType: "bearer", Language: "zh-CN"}
}

// Apply retains the saved secret when an edit leaves the credential blank.
func (s Settings) Apply(update SettingsUpdate) (Settings, error) {
	if update.Language != "" {
		s.Language = update.Language
	}
	if update.CredentialType != "" && update.CredentialType != s.CredentialType {
		if strings.TrimSpace(update.Credential) == "" && s.Credential != "" {
			return Settings{}, ErrSettings
		}
		s.CredentialType = update.CredentialType
	}
	if secret := strings.TrimSpace(update.Credential); secret != "" {
		s.Credential = secret
	}
	if s.CredentialType != "bearer" && s.CredentialType != "api_key" {
		return Settings{}, ErrSettings
	}
	if s.Language != "zh-CN" && s.Language != "zh-TW" && s.Language != "zh-HK" && s.Language != "en-US" {
		return Settings{}, ErrSettings
	}
	if len(s.Credential) > 4096 {
		return Settings{}, ErrSettings
	}
	for _, char := range s.Credential {
		if char < 33 || char > 126 {
			return Settings{}, ErrSettings
		}
	}
	return s, nil
}
