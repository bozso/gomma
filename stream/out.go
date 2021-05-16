package stream

import (
	"io"
	"os"
	"strings"

	json "git.sr.ht/~istvan_bozso/sert/json"
)

type Out struct {
	name string
	w    io.WriteCloser
}

func setToStdout(out *Out) {
	if out == nil {
		out = new(Out)
	}
	out.name = names.out
	out.w = os.Stdout
}

func Stdout() (out *Out) {
	setToStdout(out)

	return
}

// Write implements io.Writer.
func (o *Out) Write(b []byte) (n int, err error) {
	return o.w.Write(b)
}

// Close implements io.Closer.
func (o *Out) Close() (err error) {
	return o.w.Close()
}

func (o *Out) FromFile(s string) (err error) {
	w, err := os.Create(s)
	if err != nil {
		return err
	}
	o.name = s
	o.w = w

	return nil
}

func ExpectedOut() (m MismatchedStreamType) {
	return MismatchedStreamType{
		Expected: OutStream,
		Got:      StdinMeta(),
	}
}

func (o *Out) UnmarshalJSON(b []byte) (err error) {
	return o.Set(json.Trim(b))
}

func (o *Out) Set(s string) (err error) {
	switch l := strings.ToLower(s); l {
	case names.out, "":
		setToStdout(o)
	case names.in:
		err = ExpectedOut()
	default:
		err = o.FromFile(s)
	}
	return
}
