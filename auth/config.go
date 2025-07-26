package auth

import (
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func configLoad(path string, scopes []string) (*oauth2.Config, error) {
	if path == "" {
		return &oauth2.Config{}, nil
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("can't read config file '%s': %w", path, err)
	}
	// allscopes := append(scopes, "https://www.googleapis.com/auth/userinfo.email")
	allscopes := scopes
	config, err := google.ConfigFromJSON(bytes, allscopes...)
	if err != nil {
		return nil, fmt.Errorf("can't decode config: %w", err)
	}
	return config, nil
}
