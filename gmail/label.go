package gmail

import (
	"fmt"

	"google.golang.org/api/gmail/v1"
)

type Label struct {
	glabel *gmail.Label
}

func (l Label) Id() string {
	if l.IsInvalid() {
		return ""
	}
	return l.glabel.Id
}

func (l Label) IsInvalid() bool {
	return !l.IsValid()
}

func (l Label) IsSystem() bool {
	return l.IsValid() && l.glabel.Type == "system"
}

func (l Label) IsUser() bool {
	return l.IsValid() && l.glabel.Type == "user"
}

func (l Label) IsValid() bool {
	return l.glabel != nil
}

func (l Label) Name() string {
	if l.IsInvalid() {
		return ""
	}
	return l.glabel.Name
}

func (l Label) String() string {
	return fmt.Sprintf("%s (%s)", l.Name(), l.Id())
}
