package rules

import (
	"fmt"
	"strings"

	"github.com/ab36245/go-pkgs/source"
)

type Rule struct {
	source.Span
	Filter  Filter
	Actions Actions
}

type Getter func(string) any

func (r Rule) Match(getter Getter) bool {
	return r.Filter.Match(getter)
}

func (r Rule) String() string {
	s := "{\n"
	s += "  Filter:\n"
	for _, l := range strings.Split(r.Filter.String(), "\n") {
		s += fmt.Sprintf("    %s\n", l)
	}

	s += "  Actions:\n"
	for _, a := range r.Actions {
		for _, l := range strings.Split(a.String(), "\n") {
			s += fmt.Sprintf("    %s\n", l)
		}
	}
	s += "}"
	return s
}
