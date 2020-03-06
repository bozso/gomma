package stream

import (
    "fmt"
    "os"
    "io"
    "bufio"

    "github.com/bozso/gamma/utils"
)

type Out struct {
    Stream
    io.WriteCloser
}

func Create(p string) (o *Out, err error) {
    o.WriteCloser, err = os.Create(p)
    o.name = p
    
    if err != nil {
        err = o.CreateFail(err)
    }
    
    return
}


func (o *Out) SetCli(c *utils.Cli, name, usage string) {
    const defDesc = "By default it writes to standard output."
    
    c.Var(o, name, fmt.Sprintf("%s %s", usage, defDesc))
}

func (o *Out) Set(s string) (err error) {
    if len(s) == 0 {
        o.WriteCloser = os.Stdout
    } else {
        o, err = Create(s)
        
        if err != nil { return }
    }
    
    return
}

func (o Out) BufWriter() (w *bufio.Writer) {
    w = bufio.NewWriter(o.WriteCloser)
    return
}

func (o Out) WriteFail(p string, err error) error {
    return o.Fail("write to", err)
}
