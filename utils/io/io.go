package io

import (
    "strings"
    "fmt"
    "bufio"
    "os"
    "io"
    "io/ioutil"
    "path/filepath"
    
    "github.com/bozso/gamma/utils"
    "github.com/bozso/gamma/utils/path"
)

type Reader struct {
    *bufio.Scanner
    *os.File
}

func NewReader(path string) (f Reader, err error) {
    if f.File, err = os.Open(path); err != nil {
        err = OpenFail(path, err)
        return
    }

    f.Scanner = bufio.NewScanner(f.File)

    return f, nil
}

func (f *Reader) SetCli(c *utils.Cli, name, usage string) {
    const defDesc = "By default it reads from standard input."
    
    c.Var(f, name, fmt.Sprintf("%s %s", usage, defDesc))
}

func (f Reader) String() string {
    return ""
}

func (f *Reader) Set(s string) (err error) {
    var r io.Reader
    
    if len(s) == 0 {
        r = os.Stdin
    } else {
        *f, err = NewReader(s)
        if err != nil {
            return
        }
        r = f.File
    }
    
    f.Scanner = bufio.NewScanner(r)
    return nil
}

type Writer struct {
    *bufio.Writer
    *os.File
    err error
}

func NewWriter(wr io.Writer) (w Writer) {
    w.Writer = bufio.NewWriter(wr)
    return
}

func (w *Writer) SetCli(c *utils.Cli, name, usage string) {
    const defDesc = "By default it writes to standard output."
    
    c.Var(w, name, fmt.Sprintf("%s %s", usage, defDesc))
}

func (w Writer) String() string {
    return ""
}

func (w *Writer) Set(s string) (err error) {
    if len(s) == 0 {
        w.Writer = bufio.NewWriter(os.Stdout)
    } else {
        if w.File, err = os.Create(s); err != nil {
            return CreateFail(s, err)
        }
        w.Writer = bufio.NewWriter(w.File)
    }
    
    
    return nil    
}

func (w *Writer) Wrap() error {
    if w.err == nil {
        return nil
    }

    if w.File != nil { 
        return fmt.Errorf("error while writing to file '%s': %w",
            w.File.Name(), w.err)
    } else {
        return fmt.Errorf("error while writing: %w", w.err)
    }
}

func (w *Writer) Close() {
    if w.File != nil {
        w.File.Close()
    }
    
    w.Writer.Flush()
}

func WriterFromPath(name string) (w Writer) {
    w.File, w.err = os.Create(name)
    w.Writer = bufio.NewWriter(w.File)
    
    return
}

func (w *Writer) Write(b []byte) (n int) {
    if w.err != nil {
        return 0
    }
    
    n, w.err = w.Writer.Write(b)
    return
}

func (w *Writer) WriteString(s string) (n int) {
    if w.err != nil {
        return 0
    }
    
    n, w.err = w.Writer.WriteString(s)
    return
}

func (w *Writer) WriteFmt(tpl string, args ...interface{}) int {
    return w.WriteString(fmt.Sprintf(tpl, args...))
}


func ReadFile(p string) (b []byte, err error) {
    var f *os.File
    if f, err = os.Open(p); err != nil {
        err = OpenFail(p, err)
        return
    }
    defer f.Close()
    
    b, err = ioutil.ReadAll(f)
    return
}

type FileError struct {
    path, op string
    err error
}

func (e FileError) Error() string {
    return fmt.Sprintf("failed to %s file '%s'", e.op, e.path)
}

func (e FileError) Unwrap() error {
    return e.err
}

func OpenFail(p string, err error) FileError {
    return FileError{p, "open", err}
}

func CreateFail(p string, err error) FileError {
    return FileError{p, "create", err}
}

func ReadFail(p string, err error) FileError {
    return FileError{p, "read from", err}
}

func WriteFail(p string, err error) FileError {
    return FileError{p, "write to", err}
}


type Path struct {
    s string
}

func (p Path) String() string {
    return p.s
}

func (p Path) Abs() (pp Path, err error) {
    pp.s, err = filepath.Abs(p.s)
    return
}

func (p Path) Len() int {
    return len(p.s)
}

type File struct {
    Path
}

func (v *File) Set(s string) (err error) {
    b := false
    
    if len(s) == 0 {
        return utils.EmptyStringError{}
    }
    
    b, err = path.Exist(s)
    
    if err != nil { return }
    
    if !b {
        return fmt.Errorf("path '%s' does not exist", s)
    }
    
    v.s = s
    return nil
}

func (f File) Reader() (r io.Reader, err error) {
    r, err = NewReader(f.String())
    return
}

type Files []*File

func (f Files) String() string {
    if f != nil {
        // TODO: list something sensible
        return ""
    }
    
    return ""
}

func (f Files) Set(s string) (err error) {
    split := strings.Split(s, ",")
    
    f = make(Files, len(split))
    
    for ii, fpath := range f {
        if err = fpath.Set(split[ii]); err != nil {
            return
        }
    }
    return nil
}
