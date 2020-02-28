package utils

import (
    "fmt"
    "log"
    "strconv"
    "strings"
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
            end      = "\033[0m"
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

type SplitParser struct {
    split []string
    len int
}

func NewSplitParser(s, sep string) (sp SplitParser, err error) {
    sp.split = strings.Split(s, sep)
    sp.len = len(sp.split)
    
    if sp.len == 0 {
        err = fmt.Errorf(
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

    s, err := sp.Idx(idx)
    
    if err != nil {
        return
    }
    
    i, err = strconv.Atoi(s)
    return
}

func (sp SplitParser) Float(idx int) (f float64, err error) {
    s, err := sp.Idx(idx)
    if err != nil { return }
    
    f, err = strconv.ParseFloat(s, 64)
    
    return
}
