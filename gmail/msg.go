package gmail

import (
	"fmt"
	"iter"
	"net/mail"
	"slices"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"
)

func msgGet(client *Client, id string) (*Msg, error) {
	req := client.service.Users.Messages.Get("me", id)
	res, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("can't get message id %s: %w", id, err)
	}
	return &Msg{
		client:   client,
		gmessage: res,
	}, nil
}

type Msg struct {
	client   *Client
	gmessage *gmail.Message
}

func (m *Msg) Body() ([]byte, error) {
	return m.main().Body()
}

func (m *Msg) Date() time.Time {
	date, err := mail.ParseDate(m.Header("date"))
	if err != nil {
		return time.Time{}
	}
	return date.Local()
}

func (m *Msg) Client() *Client {
	return m.client
}

func (m *Msg) Filename() string {
	return m.main().Filename()
}

func (m *Msg) From() string {
	return m.Header("from")
}

func (m *Msg) FromAddress() string {
	from := m.From()
	address, err := mail.ParseAddress(from)
	if err != nil {
		return ""
	}
	return address.Address
}

func (m *Msg) FromName() string {
	from := m.From()
	address, err := mail.ParseAddress(from)
	if err != nil {
		return ""
	}
	return address.Name
}

func (m *Msg) Header(name string) string {
	return m.main().Header(name)
}

func (m *Msg) Headers() []PartHeader {
	return m.main().Headers()
}

func (m *Msg) Id() string {
	return m.gmessage.Id
}

func (m *Msg) Labels() Labels {
	return m.client.labels.ByIds(m.gmessage.LabelIds)
}

func (m *Msg) MimeType() string {
	return m.main().MimeType()
}

func (m *Msg) Subject() string {
	return m.Header("subject")
}

func (m *Msg) Parts() iter.Seq2[*Part, error] {
	var walk func(gpart *gmail.MessagePart) bool
	return func(yield func(*Part, error) bool) {
		walk = func(gpart *gmail.MessagePart) bool {
			part := m.part(gpart)
			if !yield(part, nil) {
				return false
			}
			for _, gpart := range gpart.Parts {
				if !walk(gpart) {
					return false
				}
			}
			return true
		}
		walk(m.gmessage.Payload)
	}
}

func (m *Msg) String() string {
	s := fmt.Sprintf("Id: %s\n", m.Id())

	s += "Labels:\n"
	for _, label := range m.Labels() {
		s += fmt.Sprintf("  %s\n", label)
	}

	s += "Parts:\n"
	for p := range m.Parts() {
		for _, l := range strings.Split(p.String(), "\n") {
			s += fmt.Sprintf("  %s\n", l)
		}
	}
	return s
}

func (m *Msg) To() string {
	return m.Header("to")
}

func (m *Msg) ToAddress() string {
	from := m.To()
	address, err := mail.ParseAddress(from)
	if err != nil {
		return ""
	}
	return address.Address
}

func (m *Msg) ToName() string {
	from := m.From()
	address, err := mail.ParseAddress(from)
	if err != nil {
		return ""
	}
	return address.Name
}

func (m *Msg) UpdateLabels(add []string, remove []string) (*Msg, error) {
	var addIds []string
	for _, id := range add {
		if slices.Contains(remove, id) {
			// Label is in both add and remove
			continue
		}
		if m.Labels().ById(id).IsValid() {
			// Message already has the label to be added
			continue
		}
		if strings.EqualFold(id, "sent") {
			// API request cannot adjust the SENT label
			continue
		}
		addIds = append(addIds, id)
	}
	var removeIds []string
	for _, id := range remove {
		if slices.Contains(add, id) {
			// Label is in both remove and add
			continue
		}
		if m.Labels().ById(id).IsInvalid() {
			// Message already does not have the label to be removed
			continue
		}
		if strings.EqualFold(id, "sent") {
			// API request cannot adjust the SENT label
			continue
		}
		removeIds = append(removeIds, id)
	}
	if len(addIds) == 0 && len(removeIds) == 0 {
		// Nothing to do
		return m, nil
	}
	mod := gmail.ModifyMessageRequest{
		AddLabelIds:    addIds,
		RemoveLabelIds: removeIds,
	}
	req := m.client.service.Users.Messages.Modify("me", m.Id(), &mod)
	res, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("can't modify message id %s: %w", m.Id(), err)
	}
	return msgGet(m.client, res.Id)
}

func (m *Msg) main() *Part {
	return m.part(m.gmessage.Payload)
}

func (m *Msg) part(gpart *gmail.MessagePart) *Part {
	return &Part{
		gpart: gpart,
		msg:   m,
	}
}
