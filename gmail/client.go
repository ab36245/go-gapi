package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"iter"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"github.com/ab36245/go-gapi/auth"
)

func New(options Options) (*Client, error) {
	var gservice *gmail.Service
	{
		auth, err := auth.New(options.Auth)
		if err != nil {
			return nil, err
		}
		http, err := auth.Client()
		if err != nil {
			return nil, err
		}
		ctx := context.Background()
		gservice, err = gmail.NewService(ctx, option.WithHTTPClient(http))
		if err != nil {
			return nil, fmt.Errorf("can't create gmail service: %w", err)
		}
	}

	var email string
	{
		req := gservice.Users.GetProfile("me")
		res, err := req.Do()
		if err != nil {
			return nil, fmt.Errorf("can't get profile: %w", err)
		}
		email = res.EmailAddress
	}
	fmt.Printf("email address is %q\n", email)

	var labels Labels
	{
		req := gservice.Users.Labels.List("me")
		res, err := req.Do()
		if err != nil {
			return nil, fmt.Errorf("can't get labels: %w", err)
		}

		for _, glabel := range res.Labels {
			labels = append(labels, Label{glabel})
		}
	}

	return &Client{
		gservice: gservice,
		email:    email,
		labels:   labels,
	}, nil
}

type Client struct {
	gservice *gmail.Service
	email    string
	labels   Labels
}

func (c *Client) Labels() Labels {
	return c.labels
}

func (c *Client) List(options ListOptions) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		q := options.Q()
		pageToken := ""
		for {
			req := c.gservice.Users.Messages.List("me")
			if q != "" {
				req = req.Q(q)
			}
			if pageToken != "" {
				req = req.PageToken(pageToken)
			}
			res, err := req.Do()
			if err != nil {
				yield("", fmt.Errorf("can't get messages: %w", err))
				return
			}
			if len(res.Messages) == 0 {
				return
			}
			for _, msg := range res.Messages {
				if !yield(msg.Id, nil) {
					return
				}
			}
			if res.NextPageToken == "" {
				return
			}
			pageToken = res.NextPageToken
		}
	}
}

func (c *Client) Msg(id string) (*Msg, error) {
	var gmessage *gmail.Message
	{
		req := c.gservice.Users.Messages.Get("me", id)
		var err error
		gmessage, err = req.Do()
		if err != nil {
			return nil, fmt.Errorf("can't get message id %s: %w", id, err)
		}
	}

	msg := &Msg{
		client:   c,
		gmessage: gmessage,
	}

	{
		var getParts func(gpart *gmail.MessagePart) error
		getParts = func(gpart *gmail.MessagePart) error {
			if gpart.PartId == "" || gpart.Body.Size > 0 {
				part := Part{
					msg:   msg,
					gpart: gpart,
				}
				for _, gheader := range gpart.Headers {
					header := Header{
						Name:  gheader.Name,
						Value: gheader.Value,
					}
					part.headers = append(part.headers, header)
				}
				msg.parts = append(msg.parts, part)
			}
			if len(gpart.Parts) > 0 {
				for _, subpart := range gpart.Parts {
					if err := getParts(subpart); err != nil {
						return err
					}
				}
			}
			return nil
		}
		if err := getParts(gmessage.Payload); err != nil {
			return nil, err
		}
	}

	return msg, nil
}

func (c *Client) Raw(id string) ([]byte, error) {
	var encoded string
	{
		var gmessage *gmail.Message
		req := c.gservice.Users.Messages.Get("me", id)
		req.Format("RAW")
		var err error
		gmessage, err = req.Do()
		if err != nil {
			return nil, fmt.Errorf("message %s: can't get message: %w", id, err)
		}
		encoded = gmessage.Raw
	}

	var decoded []byte
	{
		var err error
		decoded, err = base64.URLEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("message %s: can't decode: %w", id, err)
		}
	}
	return decoded, nil
}
