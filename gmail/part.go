package gmail

import (
	"encoding/base64"
	"fmt"
	"strings"

	"google.golang.org/api/gmail/v1"
)

type Part struct {
	msg     *Msg
	gpart   *gmail.MessagePart
	headers Headers
}

func (p Part) Body() ([]byte, error) {
	data := p.gpart.Body.Data
	{
		attachmentId := p.gpart.Body.AttachmentId
		if attachmentId != "" {
			service := p.msg.client.gservice.Users.Messages.Attachments
			body, err := service.Get("me", p.msg.Id(), attachmentId).Do()
			if err != nil {
				return nil, p.wrap(err, "can't get attachment id %s", attachmentId)
			}
			data = body.Data
		}
	}

	var bytes []byte
	{
		var err error
		bytes, err = base64.URLEncoding.DecodeString(data)
		if err != nil {
			return nil, p.wrap(err, "can't decode message body")
		}
	}

	return bytes, nil
}

func (p Part) ContentDisposition() string {
	return p.Header("Content-Disposition")
}

func (p Part) ContentTransferEncoding() string {
	return p.Header("Content-Transfer-Encoding")
}

func (p Part) ContentType() string {
	return p.Header("Content-Type")
}

func (p Part) FileName() string {
	name := p.gpart.Filename
	// Sanitize the file name
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	return name
}

func (p Part) Header(name string) string {
	for _, h := range p.headers {
		if h.MatchesName(name) {
			return h.Value
		}
	}
	return ""
}

func (p Part) Headers() Headers {
	return p.headers
}

func (p Part) Id() string {
	return p.gpart.PartId
}

func (p Part) MimeType() string {
	return p.gpart.MimeType
}

func (p Part) MimeVersion() string {
	return p.Header("MimeVersion")
}

func (p Part) Size() int64 {
	return p.gpart.Body.Size
}

func (p Part) id() string {
	id := p.Id()
	if id == "" {
		id = "main"
	}
	return fmt.Sprintf("%s, part %s", p.msg.id(), id)
}

func (p Part) wrap(err error, str string, args ...any) error {
	id := p.id()
	str = fmt.Sprintf(str, args...)
	return fmt.Errorf("%s: %s: %w", id, str, err)
}
