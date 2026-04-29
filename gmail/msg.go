package gmail

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"

	"github.com/ab36245/go-mimeext"
)

type Msg struct {
	client   *Client
	gmessage *gmail.Message
	parts    []Part
}

func (m Msg) Cc() (Addresses, error) {
	return m.parseAddresses("Cc")
}

func (m Msg) Client() *Client {
	return m.client
}

func (m Msg) Date() (time.Time, error) {
	return m.parseDate("Date")
}

func (m Msg) DeliveredTo() string {
	return m.Header("Delivered-To")
}

func (m Msg) From() (Addresses, error) {
	return m.parseAddresses("From")
}

func (m Msg) Header(name string) string {
	return m.parts[0].Header(name)
}

func (m Msg) Headers() Headers {
	return m.parts[0].Headers()
}

func (m Msg) Id() string {
	return m.gmessage.Id
}

func (m Msg) InReplyTo() string {
	return m.Header("In-Reply-To")
}

func (m Msg) ListUnsubscribe() string {
	return m.Header("List-Unsubscribe")
}

func (m Msg) Labels() Labels {
	return m.client.Labels().ByIds(m.gmessage.LabelIds)
}

func (m Msg) MessageId() string {
	return m.Header("Message-ID")
}

func (m Msg) Parts() []Part {
	return m.parts
}

func (m Msg) ReplyTo() (Addresses, error) {
	return m.parseAddresses("Reply-To")
}

func (m Msg) Subject() string {
	return m.Header("Subject")
}

func (m Msg) To() (Addresses, error) {
	return m.parseAddresses("To")
}

func (m Msg) SaveTo(dir string) error {
	dir = filepath.Join(dir, m.Id())
	if err := os.MkdirAll(dir, 0700); err != nil {
		return m.wrapError(err, "can't create directory %q", dir)
	}

	info := MsgInfo{}

	info.Id = m.Id()

	from, err := m.From()
	if err != nil {
		return err
	}
	info.From = from

	to, err := m.To()
	if err != nil {
		return err
	}
	info.To = to

	cc, err := m.Cc()
	if err != nil {
		return err
	}
	info.Cc = cc

	info.Subject = m.Subject()

	date, err := m.Date()
	if err != nil {
		return err
	}
	info.Date = date

	for _, part := range m.Parts() {
		if part.Size() == 0 {
			continue
		}

		pinfo := PartInfo{}
		pinfo.Id = part.Id()
		pinfo.ContentDisposition = part.ContentDisposition()
		pinfo.ContentTransferEncoding = part.ContentTransferEncoding()
		pinfo.ContentType = part.ContentType()
		pinfo.FileName = part.FileName()
		pinfo.MimeType = part.MimeType()
		pinfo.MimeVersion = part.MimeVersion()
		pinfo.Size = part.Size()

		info.Parts = append(info.Parts, pinfo)
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return m.wrapError(err, "can't encode message info")
	}

	path := filepath.Join(dir, "info.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		return m.wrapError(err, "can't write message info to %q", path)
	}

	links := map[string][]string{}
	for n, part := range m.Parts() {
		if part.Size() == 0 {
			continue
		}
		body, err := part.Body()
		if err != nil {
			return err
		}
		ext := mimeext.MimeToExt(part.MimeType())
		name := fmt.Sprintf("part%02d%s", n, ext)
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, body, 0600); err != nil {
			return m.wrapError(err, "can't write file %q", path)
		}

		lname := part.FileName()
		if lname != "" {
			links[lname] = append(links[lname], path)
		}
	}

	for lname, paths := range links {
		if len(paths) == 0 {
			return m.wrapError(err, "LOGIC ERROR! attachment %q links to 0 parts!!", lname)
		}
		if len(paths) > 1 {
			fmt.Printf("%s: ignoring attachment %q which links to %d parts\n", m.id(), lname, len(paths))
			continue
		}
		path := paths[0]
		lpath := filepath.Join(dir, lname)
		if err := os.Link(path, lpath); err != nil {
			return m.wrapError(err, "can't create link %q -> %q", path, lpath)
		}
	}

	return nil
}

func (m *Msg) UpdateLabels(add []string, remove []string) (*Msg, error) {
	labels := m.client.Labels()

	var addIds []string
	for _, id := range add {
		if slices.Contains(remove, id) {
			// Label is in both add and remove
			continue
		}
		if labels.ById(id).IsValid() {
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
		if labels.ById(id).IsInvalid() {
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
	req := m.client.gservice.Users.Messages.Modify("me", m.Id(), &mod)
	res, err := req.Do()
	if err != nil {
		return nil, m.wrapError(err, "can't modify message", m.Id())
	}
	return m.client.Msg(res.Id)
}

func (m Msg) id() string {
	return fmt.Sprintf("message %s", m.Id())
}

func (m Msg) parseAddress(name string) (Address, error) {
	value := m.Header(name)
	if value == "" {
		return Address{}, nil
	}

	// This is a horrible hack to handle some weird address lines that include charset=
	re := regexp.MustCompile(`charset=\S+`)
	value = re.ReplaceAllString(value, "")

	parsed, err := mail.ParseAddress(value)
	if err != nil {
		return Address{}, m.wrapError(err, "can't parse %s address", name)
	}
	return Address{
		Address: parsed.Address,
		Name:    parsed.Name,
	}, nil
}

func (m Msg) parseAddresses(name string) (Addresses, error) {
	value := m.Header(name)
	if value == "" {
		return nil, nil
	}

	// This is a horrible hack to handle some weird address lines that include charset=
	re := regexp.MustCompile(`charset=\S+`)
	value = re.ReplaceAllString(value, "")

	parsed, err := mail.ParseAddressList(value)
	if err != nil {
		return nil, m.wrapError(err, "can't parse %s address(es)", name)
	}
	var addresses Addresses
	for _, p := range parsed {
		addresses = append(addresses, Address{
			Address: p.Address,
			Name:    p.Name,
		})
	}
	return addresses, nil
}

func (m Msg) parseDate(name string) (time.Time, error) {
	value := m.Header(name)
	date, err := mail.ParseDate(value)
	if err != nil {
		return time.Time{}, m.wrapError(err, "can't parse %s header", name)
	}
	return date.Local(), nil
}

func (m Msg) wrapError(err error, str string, args ...any) error {
	id := m.id()
	str = fmt.Sprintf(str, args...)
	return fmt.Errorf("%s: %s: %w", id, str, err)
}

type MsgInfo struct {
	Id      string     `json:"id"`
	From    Addresses  `json:"from,omitzero"`
	To      Addresses  `json:"to,omitempty"`
	Cc      Addresses  `json:"cc,omitempty"`
	Subject string     `json:"subject,omitempty"`
	Date    time.Time  `json:"date,omitzero"`
	Parts   []PartInfo `json:"parts,omitempty"`
}

type PartInfo struct {
	Id                      string `json:"part-id,omitempty"`
	ContentDisposition      string `json:"content-disposition,omitempty"`
	ContentTransferEncoding string `json:"content-transfer-encoding,omitempty"`
	ContentType             string `json:"content-type,omitempty"`
	FileName                string `json:"filename,omitempty"`
	MimeType                string `json:"mime-type,omitempty"`
	MimeVersion             string `json:"mime-version,omitempty"`
	Size                    int64  `json:"size,omitempty"`
}
