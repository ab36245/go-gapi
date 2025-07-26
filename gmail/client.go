package gmail

import (
	"context"
	"fmt"
	"iter"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"github.com/ab36245/go-pkgs/gapi/auth"
)

func New(options Options) (*Client, error) {
	auth, err := auth.New(options.Auth)
	if err != nil {
		return nil, err
	}
	http, err := auth.Client()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	service, err := gmail.NewService(ctx, option.WithHTTPClient(http))
	if err != nil {
		return nil, fmt.Errorf("can't create gmail service: %w", err)
	}

	var email string
	{
		req := service.Users.GetProfile("me")
		res, err := req.Do()
		if err != nil {
			return nil, fmt.Errorf("can't get profile: %w", err)
		}
		email = res.EmailAddress
	}
	fmt.Printf("email address is %q\n", email)

	var labels []Label
	{
		req := service.Users.Labels.List("me")
		res, err := req.Do()
		if err != nil {
			return nil, fmt.Errorf("can't get labels: %w", err)
		}

		for _, glabel := range res.Labels {
			labels = append(labels, Label{glabel})
		}
	}

	client := &Client{
		email:   email,
		labels:  labels,
		service: service,
	}

	return client, nil
}

type Client struct {
	email   string
	labels  Labels
	service *gmail.Service
}

func (c *Client) Labels() Labels {
	return c.labels
}

func (c *Client) List(options ListOptions) iter.Seq2[string, error] {
	return listGet(c, options)
}

func (c *Client) Get(id string) (*Msg, error) {
	return msgGet(c, id)
}
