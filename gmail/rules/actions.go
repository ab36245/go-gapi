package rules

import (
	"fmt"
	"strings"
)

type Actions []Action

func (as Actions) String() string {
	s := "[\n"
	for _, a := range as {
		for _, l := range strings.Split(a.String(), "\n") {
			s += fmt.Sprintf("  %s\n", l)
		}
	}
	s += "]"
	return s
}

type Action interface {
	String() string
}

type AddAction struct {
	Label string
}

func (a AddAction) String() string {
	return fmt.Sprintf("add %s", a.Label)
}

type DeleteAction struct {
}

func (a DeleteAction) String() string {
	return "delete"
}

type MoveToAction struct {
	Label string
}

func (a MoveToAction) String() string {
	return fmt.Sprintf("moveto %s", a.Label)
}

type RemoveAction struct {
	Label string
}

func (a RemoveAction) String() string {
	return fmt.Sprintf("remove %s", a.Label)
}
