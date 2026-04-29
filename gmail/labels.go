package gmail

import (
	"fmt"
	"strings"
)

func getLabels(client *Client) (Labels, error) {
	req := client.gservice.Users.Labels.List("me")
	res, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("can't get labels: %w", err)
	}

	var labels []Label
	for _, glabel := range res.Labels {
		labels = append(labels, Label{glabel})
	}
	return labels, nil
}

type Labels []Label

func (ls Labels) ById(id string) Label {
	for _, label := range ls {
		if label.Id() == id {
			return label
		}
	}
	return Label{}
}

func (ls Labels) ByIds(ids []string) Labels {
	var labels Labels
	for _, id := range ids {
		label := ls.ById(id)
		if label.IsValid() {
			labels = append(labels, label)
		}
	}
	return labels
}

func (ls Labels) ByName(name string) Label {
	for _, label := range ls {
		if strings.EqualFold(label.Name(), name) {
			return label
		}
	}
	return Label{}
}

func (ls Labels) ByNames(names []string) Labels {
	var labels Labels
	for _, name := range names {
		label := ls.ByName(name)
		if label.IsValid() {
			labels = append(labels, label)
		}
	}
	return labels
}

func (ls Labels) Ids() []string {
	var ids []string
	for _, label := range ls {
		ids = append(ids, label.Id())
	}
	return ids
}

func (ls Labels) Names() []string {
	var names []string
	for _, label := range ls {
		names = append(names, label.Name())
	}
	return names
}
