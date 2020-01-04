package gamma

import (
    "fmt"
    "log"
    "os"
    "io"
    "os/exec"
    "bufio"
    "io/ioutil"
    "strconv"
    "strings"
    "reflect"
    "flag"
    "errors"
)

type (
    CmdFun     func(args ...interface{}) (string, error)
    Joiner     func(args ...string) string

    Params struct {
        Par string `json:"paramfile" name:"par"`
        Sep string `json:"separator" name:"sep" default:":"`
        contents []string
    }

    Tmp struct {
        files []string
    }
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

const (
    ParseIntErr Werror = "failed to parse '%s' into an integer"
    ParseFloatErr Werror = "failed to parse '%s' into an float"
)

type EmptyStringError struct {
    variable string
    err      error
}

func (e EmptyStringError) Error() (s string) {
    s = "expected non empty string"
    
    if v := e.variable; len(v) > 0 {
        s = fmt.Sprintf("%s for '%s'", s, v)
    }
    
    return
}

func (e EmptyStringError) Unwrap() error {
    return e.err
}

func MapKeys(dict interface{}) (ret []string) {
    val := reflect.ValueOf(dict)
    kind := val.Kind()
    
    if kind != reflect.Map {
        log.Fatalf("expected a map not an '%s'", kind)
    }
    
    keys := val.MapKeys()
    ret = make([]string, len(keys))

    for ii, key := range keys {
        ret[ii] = key.String()
    }
    
    return
}


var tmp = Tmp{}

func Empty(s string) bool {
    return len(s) == 0
}

func Exist(s string) (ret bool, err error) {
    _, err = os.Stat(s)

    if err != nil {
        if os.IsNotExist(err) {
            return false, nil
        }
        return false, err
    }
    return true, nil
}

type UnrecognizedMode struct {
    name, got string
    err error
}

func (e UnrecognizedMode) Error() string {
    return fmt.Sprintf("unrecognized mode '%s' for %s", e.got, e.name)
}

func (e UnrecognizedMode) Unwrap() error {
    return e.err
}

type ModeError struct {
    name string
    got fmt.Stringer
    err error
}

func (e ModeError) Error() string {
    return fmt.Sprintf("unrecognized mode '%s' for %s", e.got.String(), e.name)
}

func (e ModeError) Unwrap() error {
    return e.err
}

type (
    Decodable interface {
        Decode(string) error
    }
    
    Action interface {
        SetCli(*Cli)
        Run() error
    }
    
    subcommand struct {
        action Action
        cli Cli
    }
    
    subcommands map[string]subcommand
    
    Cli struct {
        desc string
        *flag.FlagSet
        decodables map[*string]Decodable
        subcommands
    }
)

func (c subcommands) Keys() []string {
    s := make([]string, len(c))
    
    ii := 0
    for k := range c {
        s[ii] = k
        ii++
    }
    return s
}

func NewCli(name, desc string) (c Cli) {
    c.desc = desc
    c.FlagSet = flag.NewFlagSet(name, flag.ContinueOnError)
    c.decodables = make(map[*string]Decodable)
    c.subcommands = make(map[string]subcommand)
    
    return c
}

func (c *Cli) AddAction(name, desc string, act Action) {
    c.subcommands[name] = subcommand{
        action: act,
        cli: NewCli(name, desc),
    }
}

func (c *Cli) StringVar(name, usage string, p *string) {
    c.FlagSet.StringVar(p, name, *p, usage) 
} 

func (c *Cli) VarFlag(name, usage string, dec Decodable) {
    str := c.FlagSet.String(name, "", usage)
    
    c.decodables[str] = dec
}

func (c Cli) HasSubcommands() bool {
    return c.subcommands != nil && len(c.subcommands) != 0
}

func (c Cli) Usage() {
    fmt.Printf("Program: %s. Description: %s\n",
        Color(c.Name(), Bold), c.desc)
    c.PrintDefaults()
    
    if c.HasSubcommands() {
        fmt.Printf("\nAvailable subcommands: %s\n", c.subcommands.Keys())
    }
}

func (c Cli) Parse(args []string) (err error) {
    err = c.FlagSet.Parse(args)
    
    if errors.Is(err, flag.ErrHelp) {
        //c.Usage()
        os.Exit(0)
    }
    
    if err != nil {
        return
    }
    
    for key, val := range c.decodables {
        if err = val.Decode(*key); err != nil {
            return
        }
    }
    
    return nil
}

func (c Cli) Run(args []string) (err error) {
    //ferr := merr.Make("Cli.Run")
    
    if !c.HasSubcommands() {
        err = c.Parse(args)
        return
    }
    
    l := len(args)
    
    if l < 1 {
        fmt.Printf("Expected at least one parameter specifying subcommand.\n")
        c.Usage()
        return nil
    }
    
    // TODO: check if args is long enough
    mode := args[0]
    
    if mode == "-help" || mode == "-h" {
        c.Usage()
        return nil
    }
    
    subcom, ok := c.subcommands[mode]
    
    if !ok {
        return UnrecognizedMode{got:mode, name:"gamma"}
    }
    
    cli, act := &subcom.cli, subcom.action
    subcom.action.SetCli(cli)
    
    err = cli.Parse(args[1:])
    
    if err != nil {
        return
    }
    
    return act.Run()
}

func Fatal(err error, format string, args ...interface{}) {
    if err != nil {
        str := fmt.Sprintf(format, args...)
        log.Fatalf("Error: %s\nError: %s", str, err)
    }
}

func Handle(err error, format string, args ...interface{}) error {
    str := fmt.Sprintf(format, args...)

    if err == nil {
        return fmt.Errorf("%s", str)
    } else {
        return fmt.Errorf("%s: %w", str, err)
    }
}

const (
    CmdErr Werror = "execution of command '%s' failed"
    ExeErr Werror = `Command '%s %s' failed!
    Output of command is: %v`
)

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

const (
    FileOpenErr Werror = "failed to open file '%s'"
    FileReadErr Werror = "failed to read file '%s'"
)

type FileReader struct {
    *bufio.Scanner
    *os.File
}

func NewReader(path string) (f FileReader, err error) {
    if f.File, err = os.Open(path); err != nil {
        err = FileOpenErr.Wrap(err, path)
        return
    }

    f.Scanner = bufio.NewScanner(f.File)

    return f, nil
}

func (f *FileReader) Decode(s string) (err error) {
    var r io.Reader
    if s == "-" {
        r = os.Stdin
    } else {
        if f.File, err = os.Open(s); err != nil {
            err = FileOpenErr.Wrap(err, s)
            return
        }
        r = f.File
    }
    
    f.Scanner = bufio.NewScanner(r)
    return nil
}

func ReadFile(path string) (b []byte, err error) {
    var f *os.File
    if f, err = os.Open(path); err != nil {
        err = FileOpenErr.Wrap(err, path)
        return
    }
    defer f.Close()
    
    if b, err = ioutil.ReadAll(f); err != nil {
        err = FileReadErr.Wrap(err, path)
    }
    
    return
}

type ParameterError struct {
    path, par string
    Err error
}

func (p ParameterError) Error() string {
    return fmt.Sprintf("failed to retreive parameter '%s' from file '%s'",
        p.par, p.path)
}

func (p ParameterError) Unwrap() error {
    return p.Err
}

func FromString(params, sep string) Params {
    return Params{Par: "", Sep: sep, contents: strings.Split(params, "\n")}
}

func (self *Params) Param(name string) (ret string, err error) {
    if self.contents == nil {
        var reader FileReader
        if reader, err = NewReader(self.Par); err != nil {
            return
        }
        defer reader.Close()

        for reader.Scan() {
            line := reader.Text()
            
            if strings.Contains(line, name) {
                return strings.Trim(strings.Split(line, self.Sep)[1], " "), nil
            }
        }
    } else {
        for _, line := range self.contents {
            if strings.Contains(line, name) {
                return strings.Trim(strings.Split(line, self.Sep)[1], " "), nil
            }
        }
    }

    err = ParameterError{path: self.Par, par: name}
    return
}

func (self Params) Int(name string, idx int) (i int, err error) {
    var data string
    if data, err = self.Param(name); err != nil {
        return
    }
    
    data = strings.Split(data, " ")[idx]
    
    if i, err = strconv.Atoi(data); err != nil {
        err = ParseIntErr.Wrap(err, data)
    }

    return
}

func (self Params) Float(name string, idx int) (f float64, err error) {
    var data string
    if data, err = self.Param(name); err != nil {
        return
    }
    
    data = strings.Split(data, " ")[idx]

    if f, err = strconv.ParseFloat(data, 64); err != nil {
        err = ParseFloatErr.Wrap(err, data)
    }

    return
}

type SplitParser struct {
    s string
    split []string
    err error
}

func NewSplitParser(s, sep string) (sp SplitParser) {
    sp.s, sp.err = s, nil
    sp.split = strings.Split(s, sep)
    
    if len(sp.split) == 0 {
        sp.err = fmt.Errorf("could no be split into " +
            "multiple parts with separator '%s'", sep)
    }
    return
}

func (sp SplitParser) Wrap() error {
    if sp.err != nil {
        sp.err = fmt.Errorf("failed to parse string '%s': %w",
            sp.s, sp.err) 
    }
    return sp.err
}

func (sp *SplitParser) Int(idx int) (i int) {
    if sp.err != nil {
        return
    }
    
    i, sp.err = strconv.Atoi(sp.split[idx])
    return
}

func (sp *SplitParser) Float(idx, prec int) (f float64) {
    if sp.err != nil {
        return
    }
    
    f, sp.err = strconv.ParseFloat(sp.split[idx], prec)
    return
}

func TmpFile(ext string) (ret string, err error) {
    var file *os.File
    if len(ext) > 0 {
        file, err = ioutil.TempFile("", "*." + ext)
    } else {
        file, err = ioutil.TempFile("", "*")
    }

    if err != nil {
        err = Handle(err, "failed to create a temporary file")
        return
    }

    defer file.Close()

    name := file.Name()

    tmp.files = append(tmp.files, name)

    return name, nil
}

func RemoveTmp() {
    log.Printf("Removing temporary files...\n")
    for _, file := range tmp.files {
        if err := os.Remove(file); err != nil {
            log.Printf("Failed to remove temporary file '%s': %s\n", file, err)
        }
    }
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

func NewWriterFile(name string) (w Writer) {
    w.File, w.err = os.Create(name)
    w.Writer = bufio.NewWriter(w.File)
    
    return
}

func (w *Writer) Write(b []byte) (n int) {
    if w.err != nil {
        return
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

func Wrap(err1 error, err2 error) error {
    return fmt.Errorf("%w: %w", err1, err2)
}

type Werror string
type CWerror string

func (e Werror) Wrap(err error, args ...interface{}) error {
    str := fmt.Sprintf(string(e), args...)
    return fmt.Errorf("%s: %w", str, err)
}

func (e Werror) Make(args ...interface{}) error {
    return fmt.Errorf(string(e), args...)
}

func (e CWerror) Wrap(err error) error {
    return fmt.Errorf("%s: %w", string(e), err)
}

func (e CWerror) Make() error {
    return fmt.Errorf(string(e))
}

type FileOpenError struct {
    path string
    err error
}

func (e FileOpenError) Error() string {
    return fmt.Sprintf("failed to open file '%s'", e.path)
}

func (e FileOpenError) Unwrap() error {
    return e.err
}

func Mkdir(name string) (err error) {
    if err = os.MkdirAll(name, os.ModePerm); err != nil {
        err = fmt.Errorf("failed to create directory '%s': %w", name,
            err)
    }
    return
}

const (
    DirCreateErr Werror = "failed to create directory '%s'"
    FileExistErr Werror = "failed to determine wether '%s' exist"
    FileWriteErr Werror = "failed to write to file '%s'"
    FileCreateErr Werror = "failed to create file '%s'"
    MoveErr Werror = "failed to move '%s' to '%s'"
    EmptyStringErr Werror = "expected %s to be a non empty string"
)

type ( 
    ModuleName string
    FnName string
    
    OpError struct {
        module ModuleName
        fn     FnName
        ctx    string
        err    error
    }
    
    opErrorFactory struct {
        module ModuleName
    }
)


func NewModuleErr(mod ModuleName) opErrorFactory {
    return opErrorFactory{mod}
}

func (o opErrorFactory) Make(fn FnName) OpError {
    return OpError{module: o.module, fn:fn}
}

func (e OpError) Error() (s string) {
    s = fmt.Sprintf("%s/%s", e.module, e.fn)
    s = fmt.Sprintf("\n  %s", Color(s, Error))
    
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
