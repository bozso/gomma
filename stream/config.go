package stream

/*
import (
	"fmt"
	"os"
	"strings"
)

type Mode int

const (
	Stdout Mode = iota
	Stdin
	Path
)

func (m *Mode) Set(s string) (err error) {
	err = nil
	switch strings.ToLower(s) {
	case names.out:
		*m = Stdout
	case names.in:
		*m = Stdin
	default:
		*m = Path
	}
	return
}

func (m Mode) String() (s string) {
	switch m {
	case Stdout:
		s = names.out
	case Stdin:
		s = names.in
	default:
		s = "path"
	}
	return
}

type Config struct {
	Mode    Mode
	Logfile string `json:"logfile"`
}

func (c *Config) Set(s string) (err error) {
	err = nil
	if err = c.Mode.Set(s); err != nil {
		return
	}
	c.Logfile = s
	return
}

func (c *Config) UnmarshalJSON(b []byte) (err error) {
	err = c.Set(TrimJSON(b))
	return
}

func (c Config) Name() (s string) {
	return
}

func (c Config) NotMode(m Mode, expected string) (err error) {
	err = nil
	if m == c.Mode {
		err = fmt.Errorf("%s is not allowed for %s", m, expected)
	}
	return
}

func (c Config) ToInStream() (in In, err error) {
	switch m := c.Mode; m {
	case Stdin:
		in.name = m.String()
		in.r = os.Stdin
	case Stdout:
		err = fmt.Errorf("%s cannot be used as in stream", m.String())
	default:
		f, err := os.Open(c.Logfile)
		if err != nil {
			err = fmt.Errorf("while opening '%s': %w", c.Logfile, err)
		}
		in.name = c.Logfile
		in.r = f
	}

	return
}

func (c Config) ToOutStream() (out Out, err error) {
	switch m := c.Mode; m {
	case Stdin:
		err = fmt.Errorf("%s cannot be used as out stream", m.String())
	case Stdout:
		out.name = m.String()
		out.w = os.Stdout
	default:
		f, err := os.Create(c.Logfile)
		if err != nil {
			err = fmt.Errorf("while creating '%s': %w", c.Logfile, err)
		}
		out.name = c.Logfile
		out.w = f
	}

	return
}
*/
