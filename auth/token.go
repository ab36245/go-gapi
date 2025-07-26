package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"golang.org/x/oauth2"
)

func tokenCreate(config *oauth2.Config) (*oauth2.Token, error) {
	code, err := codeFromWeb(config)
	if err != nil {
		return nil, err
	}
	token, err := config.Exchange(context.TODO(), code)
	if err != nil {
		return nil, fmt.Errorf("can't exchange code for token: %w", err)
	}
	fmt.Printf("tokenCreate: expiry %v\n", token.Expiry)
	return token, nil
}

func tokenLoad(path string) (*oauth2.Token, error) {
	bytes, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("can't read token file '%s': %w", path, err)
	}
	token := &oauth2.Token{}
	if err := json.Unmarshal(bytes, token); err != nil {
		return nil, fmt.Errorf("can't decode token: %w", err)
	}
	fmt.Printf("tokenLoad: expiry %v\n", token.Expiry)
	return token, nil
}

func tokenRefresh(config *oauth2.Config, token *oauth2.Token) (*oauth2.Token, error) {
	fmt.Printf("tokenRefresh: old expiry %v\n", token.Expiry)
	token, err := config.TokenSource(context.TODO(), token).Token()
	if err != nil {
		return nil, fmt.Errorf("can't refresh token: %s", err)
	}
	fmt.Printf("tokenRefresh: new expiry %v\n", token.Expiry)
	return token, nil
}

func tokenSave(path string, token *oauth2.Token) error {
	bytes, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("can't encode token: %s", err)
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("can't create token file '%s': %s", path, err)
	}
	defer file.Close()
	if _, err := file.Write(bytes); err != nil {
		return fmt.Errorf("can't write token file '%s': %s", path, err)
	}
	return nil
}
