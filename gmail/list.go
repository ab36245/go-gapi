package gmail

import (
	"fmt"
	"strings"
	"time"
)

type ListOptions struct {
	After  time.Time
	Before time.Time
	Labels []string
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
