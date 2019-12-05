package gamma

import (
    //"errors"
    "fmt"
    "log"
    "os"
    "os/exec"
    "bufio"
    "io/ioutil"
    "strconv"
    "strings"
    "reflect"
)

type (
    CmdFun     func(args ...interface{}) (string, error)
    Joiner     func(args ...string) string

    FileReader struct {
        *bufio.Scanner
        *os.File
    }

    Params struct {
        Par string `json:"paramfile" name:"par"`
        Sep string `json:"separator" name:"sep" default:":"`
        contents []string
    }

    Tmp struct {
        files []string
    }
)

type Args struct {
    opt map[string]string
    pos []string
    npos int
}

func NewArgs(args []string) (ret Args) {
    ret.opt = make(map[string]string)
    
    for _, arg := range args {
        if strings.Contains(arg, "=") {
            split := strings.Split(arg, "=")
            
            ret.opt[split[0]] = split[1]
        } else {
            ret.pos = append(ret.pos, arg)
        }
    }
    
    ret.npos = len(ret.pos)
    
    return
}

const (
    ParseIntErr Werror = "failed to parse '%s' into an integer"
    ParseFloatErr Werror = "failed to parse '%s' into an float"
)

func StringToVal(v reflect.Value, kind reflect.Kind, in string) error {
    switch kind {
    case reflect.Int:
        set, err := strconv.Atoi(in)
        if err != nil {
            return ParseIntErr.Wrap(err, in)
        }
        
        v.SetInt(int64(set))
    case reflect.Float32:
        set, err := strconv.ParseFloat(in, 32)
        
        if err != nil {
            return ParseFloatErr.Wrap(err, in)
        }
        
        v.SetFloat(set)
    case reflect.Float64:
        set, err := strconv.ParseFloat(in, 64)
        
        if err != nil {
            return ParseFloatErr.Wrap(err, in)
        }
        
        v.SetFloat(set)
    case reflect.Bool:
        v.SetBool(true)
    case reflect.String:
        v.SetString(in)
    }
    return nil
}

const (
    ParseFieldErr Werror = "parsing of struct field '%s' failed"
    SetFieldErr Werror = "failed to set struct field '%s'"
    ParseStructErr Werror = "failed to parse struct %s"
)

func (h Args) ParseStruct(s interface{}) error {
    vptr := reflect.ValueOf(s)
    kind := vptr.Kind()
    
    if kind != reflect.Ptr {
        return fmt.Errorf("expected a pointer to struct not '%v'", kind)
    }
    
    v := vptr.Elem()
    
    if err := h.parseStruct(v); err != nil {
        return ParseStructErr.Wrap(err, v.Type().Name())
    }
    return nil
}

func (h Args) parseStruct(v reflect.Value) error {
    t := v.Type()

    for ii := 0; ii < v.NumField(); ii++ {
        sField := t.Field(ii)
        sValue := v.Field(ii)
        kind := sField.Type.Kind()
        
        //fmt.Printf("Parsing field[%d]: %s\n", ii, sField.Name)
        
        if kind == reflect.Struct {
            if err := h.parseStruct(sValue); err != nil {
                return ParseFieldErr.Wrap(err, sField.Name)
            }
            continue
        }
        
        tag := sField.Tag
        
        if tag == "" {
            continue
        }
        
        pos := tag.Get("pos")
        npos := len(pos)
        
        if npos > 0 {
            idx, err := strconv.Atoi(pos)
            
            if err != nil {
                return ParseIntErr.Wrap(err, pos)
            }
            
            if idx >= h.npos {
                return Handle(nil, 
                    "index %d is out of the bounds of positional arguments",
                    idx)
            }
            
            if err = StringToVal(sValue, kind, h.pos[idx]); err != nil {
                return SetFieldErr.Wrap(err, sField.Name)
            }
            continue
        }
        
        name := tag.Get("name")
        
        if name == "" || name == "-" {
            name = sField.Name
        }
        
        if kind == reflect.Bool {
            val := false
            for _, pos := range h.pos {
                if pos == name {
                    val = true
                    break
                }
            }
            sValue.SetBool(val)
            continue
        }
        
        val, ok := h.opt[name]
        
        if !ok {
            val = tag.Get("default")
        }
        
        if err := StringToVal(sValue, kind, val); err != nil {
            return SetFieldErr.Wrap(err, sField.Name)
        }
    }
    return nil
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
    ExeErr Werror = `Command '%v' failed!
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
            return "", ExeErr.Wrap(err, cmd, result)
        }

        return result, nil
    }
}

const (
    FileOpenErr Werror = "failed to open file '%s'"
    FileReadErr Werror = "failed to read file '%s'"
)

func NewReader(path string) (ret FileReader, err error) {
    ret.File, err = os.Open(path)

    if err != nil {
        err = FileOpenErr.Wrap(err, path)
        return
    }

    ret.Scanner = bufio.NewScanner(ret.File)

    return ret, nil
}

func ReadFile(path string) (ret []byte, err error) {
    f, err := os.Open(path)
    if err != nil {
        err = FileOpenErr.Wrap(err, path)
        return
    }

    defer f.Close()

    contents, err := ioutil.ReadAll(f)
    if err != nil {
        err = FileReadErr.Wrap(err, path)
        return
    }

    return contents, nil
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
        var file *os.File
        file, err = os.Open(self.Par)

        if err != nil {
            err = FileOpenErr.Wrap(err, self.Par)
            return
        }

        defer file.Close()
        scanner := bufio.NewScanner(file)

        for scanner.Scan() {
            line := scanner.Text()
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

func (self Params) Int(name string, idx int) (ret int, err error) {
    data, err := self.Param(name)
    
    if err != nil {
        return ret, err
    }
    
    data = strings.Split(data, " ")[idx]
    
    ret, err = strconv.Atoi(data)

    if err != nil {
        err = ParseIntErr.Wrap(err, data)
        return
    }

    return ret, nil
}

func (self Params) Float(name string, idx int) (ret float64, err error) {
    data, err := self.Param(name)
    
    if err != nil {
        return 0.0, err
    }
    
    data = strings.Split(data, " ")[idx]
    
    ret, err = strconv.ParseFloat(data, 64)

    if err != nil {
        err = ParseFloatErr.Wrap(err, data)
        return
    }

    return ret, nil
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

const (
    DirCreateErr Werror = "failed to create directory '%s'"
    FileExistErr Werror = "failed to determine wether '%s' exist"
    FileWriteErr Werror = "failed to write to file '%s'"
    FileCreateErr Werror = "failed to create file '%s'"
    MoveErr Werror = "failed to move '%s' to '%s'"
    EmptyStringErr Werror = "expected %s to be a non empty string"
)
