package stream

import (
    "fmt"
    "os"
    "io"
    "bufio"

    "github.com/bozso/gamma/utils"
)

type In struct {
    Stream
    io.ReadCloser
}

func Open(p string) (i *In, err error) {
    i.ReadCloser, err = os.Open(p)
    i.name = p
    
    if err != nil {
        err = i.OpenFail(err)
    }
    
    return
}

func (i *In) SetCli(c *utils.Cli, name, usage string) {
    const defDesc = "By default it reads from standard input."
    
    c.Var(i, name, fmt.Sprintf("%s %s", usage, defDesc))
}

func (i *In) Set(s string) (err error) {
    var r io.ReadCloser
    
    if len(s) == 0 {
        i.ReadCloser, i.name = os.Stdin, "stdin"
    } else {
        i, err = Open(s)
        if err != nil { return }
    }
    
    i.ReadCloser = r
    return
}

func (i In) Scanner() (s *bufio.Scanner) {
    s = bufio.NewScanner(i.ReadCloser)
    return
}

func (i In) Reader() (r *bufio.Reader) {
    r = bufio.NewReader(i.ReadCloser)
    return
}

func (i In) ReadFail(err error) error {
    return i.Fail("read from", err)
}
