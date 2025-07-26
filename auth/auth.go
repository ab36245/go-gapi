package auth

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	goauth2 "google.golang.org/api/oauth2/v1"
	"google.golang.org/api/option"
)

func New(options Options) (*Auth, error) {
	config, err := configLoad(options.ConfigFile, options.Scopes)
	if err != nil {
		return nil, err
	}
	return &Auth{
		config:  config,
		options: options,
	}, nil
}

type Auth struct {
	config       *oauth2.Config
	emailAddress string
	options      Options
}

func (a *Auth) Id() (string, error) {
	client, err := a.Client()
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	service, err := goauth2.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", fmt.Errorf("can't create oauth2 service: %w", err)
	}
	req := service.Userinfo.V2.Me.Get()
	res, err := req.Do()
	if err != nil {
		return "", fmt.Errorf("can't get userinfo: %w", err)
	}
	return res.Email, nil
}

func (a *Auth) Client() (*http.Client, error) {
	token, err := a.Token()
	if err != nil {
		return nil, err
	}
	client := a.config.Client(context.Background(), token)
	return client, nil
}

func (a *Auth) Token() (*oauth2.Token, error) {
	token, err := tokenLoad(a.options.TokenFile)
	if err != nil {
		return nil, err
	}
	if token != nil && token.Valid() {
		return token, nil
	}
	if token == nil {
		newToken, err := tokenCreate(a.config)
		if err != nil {
			return nil, err
		}
		token = newToken
	} else {
		newToken, err := tokenRefresh(a.config, token)
		if err != nil {
			return nil, err
		}
		token = newToken
	}
	if err := tokenSave(a.options.TokenFile, token); err != nil {
		return nil, err
	}
	return token, err
}
