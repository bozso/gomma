package utils

import (
    "fmt"
    "log"
    "os/exec"
    "strconv"
    "strings"
)

const merr = ModuleName("gamma.utils")

type (
    CmdFun     func(args ...interface{}) (string, error)
    Joiner     func(args ...string) string
)

type ColorCode int

const (
    Info ColorCode = iota
    Notice
    Warning
    Error
    Debug
    Bold
)

func Color(s string, color ColorCode) string {
    const (
            info     = "\033[1;34m%s\033[0m"
            notice   = "\033[1;36m%s\033[0m"
            warning  = "\033[1;33m%s\033[0m"
            error_   = "\033[1;31m%s\033[0m"
            debug    = "\033[0;36m%s\033[0m"
            bold     = "\033[1;0m%s\033[0m"
    )
    
    var format string
    
    switch color {
    case Info:
    format = info
    case Notice:
    format = notice
    case Warning:
    format = warning
    case Error:
    format = error_
    case Debug:
    format = debug
    case Bold:
    format = bold
    }
    
    return fmt.Sprintf(format, s)
}

func Empty(s string) bool {
    return len(s) == 0
}


func Fatal(err error, format string, args ...interface{}) {
    if err != nil {
        str := fmt.Sprintf(format, args...)
        log.Fatalf("Error: %s; %s", str, err)
    }
}


func MakeCmd(cmd string) CmdFun {
    return func(args ...interface{}) (string, error) {
        arg := make([]string, len(args))

        for ii, elem := range args {
            if elem != nil {
                arg[ii] = fmt.Sprint(elem)
            } else {
                arg[ii] = "-"
            }
        }

        // fmt.Printf("%s %s\n", cmd, str.Join(arg, " "))
        // os.Exit(0)

        out, err := exec.Command(cmd, arg...).CombinedOutput()
        result := string(out)

        if err != nil {
            return "", ExeErr.Wrap(err, cmd, strings.Join(arg, " "), result)
        }

        return result, nil
    }
}

type SplitParser struct {
    split []string
    len int
}

func NewSplitParser(s, sep string) (sp SplitParser, err error) {
    sp.split = strings.Split(s, sep)
    sp.len = len(sp.split)
    
    if sp.len == 0 {
        err = merr.Make("NewSplitParser").Fmt(
            "string '%s' could no be split into " +
            "multiple parts with separator '%s'", s, sep)
    }
    
    return
}

func (sp SplitParser) Len() int {
    return sp.len
}

func (sp SplitParser) Idx(idx int) (s string, err error) {
    if length := sp.len; idx >= length {
        err = OutOfBoundError{idx: idx, length: length}
        return
    }
    
    return sp.split[idx], nil
}

func (sp SplitParser) Int(idx int) (i int, err error) {
    ferr := merr.Make("SplitParser.Int")
    
    s, err := sp.Idx(idx)
    
    if err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    i, err = strconv.Atoi(s)
    
    if err != nil {
        err = ferr.Wrap(err)
    }
    
    return
}

func (sp SplitParser) Float(idx int) (f float64, err error) {
    s, err := sp.Idx(idx)
    if err != nil { return }
    
    f, err = strconv.ParseFloat(s, 64)
    
    return
}


type ( 
    ModuleName string
    FnName string
    
    OpError struct {
        module ModuleName
        fn     FnName
        ctx    string
        err    error
    }
)

func (m ModuleName) Make(fn FnName) OpError {
    return OpError{module: m, fn:fn}
}

func (e OpError) Error() (s string) {
    s = fmt.Sprintf("\n  %s/%s", e.module, e.fn)
    //s = fmt.Sprintf("\n  %s", Color(s, Error))
    
    if ctx := e.ctx; len(ctx) > 0 {
        s = fmt.Sprintf("%s: %s", s, ctx)
    } 
    
    if e.err != nil {
        s = fmt.Sprintf("%s: %s", s, e.err)
    }
    
    return
}

func (e OpError) Unwrap() error {
    return e.err
}

func (e OpError) Wrap(err error) error {
    e.err = err
    return e
}

func (e OpError) WrapFmt(err error, msg string, args ...interface{}) error {
    e.err = err
    e.ctx = fmt.Sprintf(msg, args...)
    
    return e
}

func (e OpError) Fmt(msg string, args ...interface{}) error {
    e.err = fmt.Errorf(msg, args...)
    
    return e
}
