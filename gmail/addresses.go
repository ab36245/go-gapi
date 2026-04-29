package gmail

import (
	"strings"
)

type Addresses []Address

func (as Addresses) String() string {
	if len(as) == 0 {
		return ""
	}
	var s strings.Builder
	s.WriteString(as[0].String())
	for i := 1; i < len(as); i++ {
		s.WriteString("," + as[i].String())
	}
	return s.String()
}
