package data

import (
	"fmt"
	"io"
	"strings"
)

type CreationMode int

const (
	CreationNone CreationMode = iota
	CreationCommand
)

func (c CreationMode) Some() (b bool) {
	return c != CreationNone
}

type CreatedBy struct {
	Mode CreationMode `json:"mode"`
	Cmd  string       `json:"command"`
}

func CreationUnknown() (c CreatedBy) {
	return CreatedBy{Mode: CreationNone, Cmd: ""}
}

func (c CreatedBy) DescribeTo(w io.Writer) (n int, err error) {
	if !c.Mode.Some() {
		return fmt.Fprintf(w, "no information available on creation")
	}

	return fmt.Fprintf(w, "created by running command: '%s'", c.Cmd)
}

func (c CreatedBy) Decsribe() (s string, err error) {
	sb := new(strings.Builder)
	_, err = c.DescribeTo(sb)
	s = sb.String()

	return
}

func (c CreatedBy) Command() (cmd string, b bool) {
	b = false
	if c.Mode.Some() {
		b = true
	}

	return c.Cmd, b
}

func (c *CreatedBy) UnmarshalJSON(b []byte) (err error) {
	if len(b) == 0 {
		c.Mode = CreationNone
	} else {
		c.Mode = CreationCommand
		c.Cmd = string(b)
	}

	return nil
}

func (c *CreatedBy) MarshalJSON() (b []byte, err error) {
	if c.Mode.Some() {
		b = []byte(c.Cmd)
	}

	return
}
