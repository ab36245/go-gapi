package gmail

import (
	"fmt"
	"iter"
	"strings"
	"time"
)

type ListOptions struct {
	After  time.Time
	Before time.Time
	Labels []string
}

func listGet(client *Client, options ListOptions) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		q := options.Q()
		pageToken := ""
		for {
			req := client.service.Users.Messages.List("me")
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

func (o ListOptions) Q() string {
	var q []string

	if !o.After.IsZero() {
		after := o.After.Format("2006/1/2")
		q = append(q, fmt.Sprintf("after:%s", after))
	}

	if !o.Before.IsZero() {
		before := o.Before.Format("2006/1/2")
		q = append(q, fmt.Sprintf("before:%s", before))
	}

	if len(o.Labels) > 0 {
		in := ""
		for _, label := range o.Labels {
			if in != "" {
				in += " OR "
			}
			in += fmt.Sprintf("in:%q", label)
		}
		q = append(q, in)
	}

	return strings.Join(q, " ")
}
