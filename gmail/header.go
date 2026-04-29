package gmail

import (
	"fmt"
	"slices"
	"strings"
)

type Header struct {
	Name  string
	Value string
}

func (h Header) MatchesName(name string) bool {
	return strings.EqualFold(h.Name, name)
}

func (h Header) MatchesNames(names []string) bool {
	return slices.ContainsFunc(names, h.MatchesName)
}

func (h Header) MatchesPrefix(prefix string) bool {
	return strings.HasPrefix(h.Name, prefix)
}

func (h Header) MatchesPrefixes(prefixes []string) bool {
	return slices.ContainsFunc(prefixes, h.MatchesPrefix)
}

func (h Header) String() string {
	return fmt.Sprintf("%s: %q", h.Name, h.Value)
}
