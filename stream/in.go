package stream

import (
	"io"
	"os"
	"strings"
)

type In struct {
	name string
	r    io.ReadCloser
	Meta Meta
}

// Read implements io.Reader.
func (i *In) Read(b []byte) (n int, err error) {
	return i.r.Read(b)
}

// Close implements io.Closer.
func (i *In) Close() (err error) {
	return i.r.Close()
}

func setStdin(in *In) {
	in.name = names.in
	in.r = os.Stdin
	in.Meta = StdinMeta()
}

func (i *In) Set(s string) (err error) {
	switch strings.ToLower(s) {
	case names.in, "":
		setStdin(i)
	case names.out:
		err = MismatchedStreamType{
			Expected: InStream,
			Got:      StdoutMeta(),
		}
	default:
		r, err := open(s)
		if err != nil {
			return err
		}
		i.name = s
		i.r = r
		i.Meta = InMeta(PathMode)
	}
	return
}
