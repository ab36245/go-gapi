package gmail

import (
	"fmt"
	"net/mail"
)

type Address struct {
	Address string `json:"address"`
	Name    string `json:"name,omitempty"`
}

func (a Address) String() string {
	if a.Name == "" {
		return a.Address
	}
	return fmt.Sprintf("%s <%s>", a.Name, a.Address)
}

func addressFromParsed(parsed *mail.Address) Address {
	return Address{
		Address: parsed.Address,
		Name:    parsed.Name,
	}
}
