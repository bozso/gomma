package stream

import (
    "fmt"
    "os"
    "io"
    "strings"

    "github.com/bozso/gotoolbox/path"
)

func TrimJSON(b []byte) (s string) {
    s = strings.Trim(string(b), "\"")
    return strings.Trim(s, " ")
}


func open(s string) (f *os.File, err error) {
    vf, err := path.New(s).ToValidFile()
    if err != nil {
        return
    }
    f, err = vf.Open()
    return
}

var names = struct {
    out, in string
}{
    out: "stdout",
    in: "stdin",
}

type In struct {
    name string
    r io.ReadCloser
}

// Read implements io.Reader.
func (i *In) Read(b []byte) (n int, err error) {
    return i.r.Read(b)
}

// Close implements io.Closer.
func (i *In) Close() (err error) {
    return i.r.Close()
}

func (i *In) Set(s string) (err error) {
    switch strings.ToLower(s) {
    case names.in, "":
        i.name = names.in
        i.r = os.Stdin
    case names.out:
        err = fmt.Errorf("stream.In cannot be set to stdout")
    default:
        r, err := open(s)
        if err != nil {
            return err
        }
        i.name = s
        i.r = r
    }
    return
}

type Out struct {
    name string
    w io.WriteCloser
}

// Write implements io.Writer.
func (o *Out) Write(b []byte) (n int, err error) {
    return o.w.Write(b)
}

// Close implements io.Closer.
func (o *Out) Close() (err error) {
    return o.w.Close()
}

func (o *Out) Set(s string) (err error) {
    switch strings.ToLower(s) {
    case names.out, "":
        o.name = names.out
        o.w = os.Stdout
    case names.in:
        err = fmt.Errorf("stream.Out cannot be set to stdin")
    default:
        w, err := os.Create(s)
        if err != nil {
            return err
        }
        o.name = s
        o.w = w
    }
    return
}
