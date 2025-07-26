package gmail

import (
	"encoding/base64"
	"fmt"
	"strings"

	"google.golang.org/api/gmail/v1"
)

type Part struct {
	gpart *gmail.MessagePart
	msg   *Msg
}

func (p *Part) Body() ([]byte, error) {
	mid := p.msg.Id()
	id := p.gpart.Body.AttachmentId
	data := p.gpart.Body.Data
	if id != "" {
		service := p.msg.client.service.Users.Messages.Attachments
		req := service.Get("me", p.msg.Id(), id)
		res, err := req.Do()
		if err != nil {
			return nil, fmt.Errorf("can't get attachment (msg %s, id %s): %w", mid, id, err)
		}
		data = res.Data
	}
	bytes, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("can't decode message body (msg %s id %s): %w", mid, id, err)
	}
	return bytes, nil
}

func (p *Part) Filename() string {
	return p.gpart.Filename
}

func (p *Part) Header(name string) string {
	for _, gheader := range p.gpart.Headers {
		if strings.EqualFold(gheader.Name, name) {
			return gheader.Value
		}
	}
	return ""
}

func (p *Part) Headers() []PartHeader {
	var headers []PartHeader
	for _, gheader := range p.gpart.Headers {
		header := PartHeader{
			Name:  gheader.Name,
			Value: gheader.Value,
		}
		headers = append(headers, header)
	}
	return headers
}

func (p *Part) Id() string {
	return p.gpart.PartId
}

func (p *Part) MimeType() string {
	return p.gpart.MimeType
}

func (p *Part) String() string {
	s := ""
	s += fmt.Sprintf("Id: %s\n", p.Id())
	s += fmt.Sprintf("Filename: %s\n", p.Filename())
	s += "Headers:\n"
	for _, h := range p.Headers() {
		s += fmt.Sprintf("  %s: %s\n", h.Name, h.Value)
	}
	s += fmt.Sprintf("MimeType: %s\n", p.MimeType())
	return s
}

type PartHeader struct {
	Name  string
	Value string
}
