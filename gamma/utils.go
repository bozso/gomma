package gamma

import (
    //"errors"
    "fmt"
    "log"
    "os"
    "os/exec"
    bio "bufio"
    io "io/ioutil"
    fp "path/filepath"
    conv "strconv"
    str "strings"
    ref "reflect"
)

type (
    CmdFun     func(args ...interface{}) (string, error)
    Joiner     func(args ...string) string

    FileReader struct {
        *bio.Scanner
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

    path struct {
        path  string
        parts []string
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
        if str.Contains(arg, "=") {
            split := str.Split(arg, "=")
            
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

func StringToVal(v ref.Value, kind ref.Kind, in string) error {
    switch kind {
    case ref.Int:
        set, err := conv.Atoi(in)
        if err != nil {
            return ParseIntErr.Wrap(err, in)
        }
        
        v.SetInt(int64(set))
    case ref.Float32:
        set, err := conv.ParseFloat(in, 32)
        
        if err != nil {
            return ParseFloatErr.Wrap(err, in)
        }
        
        v.SetFloat(set)
    case ref.Float64:
        set, err := conv.ParseFloat(in, 64)
        
        if err != nil {
            return ParseFloatErr.Wrap(err, in)
        }
        
        v.SetFloat(set)
    case ref.Bool:
        v.SetBool(true)
    case ref.String:
        v.SetString(in)
    }
    return nil
}

const (
    ParseFieldErr Werror = "parsing of struct field '%s' failed"
    SetFieldErr Werror = "failed to set struct field '%s'"
    ParseStructErr Werror = "failed to parse struct %#v"
)

func (h Args) ParseStruct(s interface{}) error {
    vptr := ref.ValueOf(s)
    kind := vptr.Kind()
    
    if kind != ref.Ptr {
        return fmt.Errorf("expected a pointer to struct not '%v'", kind)
    }
    
    v := vptr.Elem()
    
    if err := h.parseStruct(v); err != nil {
        return ParseStructErr.Wrap(err, v)
    }
    return nil
}

func (h Args) parseStruct(v ref.Value) error {
    t := v.Type()

    for ii := 0; ii < v.NumField(); ii++ {
        sField := t.Field(ii)
        sValue := v.Field(ii)
        kind := sField.Type.Kind()
        
        //fmt.Printf("Parsing field[%d]: %s\n", ii, sField.Name)
        
        if kind == ref.Struct {
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
            idx, err := conv.Atoi(pos)
            
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
        
        
        if kind == ref.Bool {
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
    val := ref.ValueOf(dict)
    kind := val.Kind()
    
    if kind != ref.Map {
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
    FileReadErr Werror = "failed to open file '%s'"
)

func NewReader(path string) (ret FileReader, err error) {
    ret.File, err = os.Open(path)

    if err != nil {
        err = FileOpenErr.Wrap(err, path)
        //err = Handle(err, "Could not open file '%s'!", path)
        return
    }

    ret.Scanner = bio.NewScanner(ret.File)

    return ret, nil
}

func NewPath(args ...string) path {
    return path{fp.Join(args...), args}
}

func ReadFile(path string) (ret []byte, err error) {
    f, err := os.Open(path)
    if err != nil {
        err = FileOpenErr.Wrap(err, path)
        return
    }

    defer f.Close()

    contents, err := io.ReadAll(f)
    if err != nil {
        err = FileReadErr.Wrap(err, path)
        return
    }

    return contents, nil
}

func FromString(params, sep string) Params {
    return Params{Par: "", Sep: sep, contents: str.Split(params, "\n")}
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
        scanner := bio.NewScanner(file)

        for scanner.Scan() {
            line := scanner.Text()
            if str.Contains(line, name) {
                return str.Trim(str.Split(line, self.Sep)[1], " "), nil
            }
        }
    } else {
        for _, line := range self.contents {
            if str.Contains(line, name) {
                return str.Trim(str.Split(line, self.Sep)[1], " "), nil
            }
        }
    }

    err = Handle(nil, "failed to find parameter '%s' in '%s'", name, self.Par)
    return
}

/*
func toInt(par string, idx int) (ret int, err error) {
    ret, err = conv.Atoi(str.Split(par, " ")[idx])

    if err != nil {
        err = Handle(err, "failed to convert string '%s' to int", par)
        return
    }

    return ret, nil
}

func toFloat(par string, idx int) (ret float64, err error) {
    ret, err = conv.ParseFloat(str.Split(par, " ")[idx], 64)

    if err != nil {
        err = Handle(err, "failed to convert string '%s' to float64", par)
        return
    }

    return ret, nil
}
*/

func (self Params) Int(name string, idx int) (ret int, err error) {
    data, err := self.Param(name)
    
    if err != nil {
        return ret, err
    }
    
    data = str.Split(data, " ")[idx]
    
    ret, err = conv.Atoi(data)

    if err != nil {
        err = ParseIntErr.Wrap(err, data)
        return
    }

    return ret, nil
    
    /*
    data, err := self.Param(name)

    if err != nil {
        return 0, err
    }

    return toInt(data, 0)
    */
}

func (self Params) Float(name string, idx int) (ret float64, err error) {
    data, err := self.Param(name)
    
    if err != nil {
        return 0.0, err
    }
    
    data = str.Split(data, " ")[idx]
    
    ret, err = conv.ParseFloat(data, 64)

    if err != nil {
        err = ParseFloatErr.Wrap(err, data)
        return
    }

    return ret, nil
    
    /*
    data, err := self.Param(name)

    if err != nil {
        return 0.0, err
    }

    return toFloat(data, 0)
    */
}

func TmpFile() (ret string, err error) {
    file, err := io.TempFile("", "*")

    if err != nil {
        err = Handle(err, "failed to create a temporary file")
        return
    }

    defer file.Close()

    name := file.Name()

    tmp.files = append(tmp.files, name)

    return name, nil
}

func TmpFileExt(ext string) (ret string, err error) {
    file, err := io.TempFile("", "*." + ext)

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


/*
func (e *werror) Wrap(err error, format string, args ...interface{}) {
    str := fmt.Sprintf(format, args...)

    if err == nil {
        return fmt.Errorf("%s: %w", str, e)
    } else {
        return fmt.Errorf("%s: %w: %w", str, err, e)
    }
}
*/

const (
    DirCreateErr Werror = "failed to create directory '%s'"
    FileExistErr Werror = "failed to determine wether '%s' exist"
    FileWriteErr Werror = "failed to write to file '%s'"
    FileCreateErr Werror = "failed to create file '%s'"
    MoveErr Werror = "failed to move '%s' to '%s'"
    EmptyStringErr Werror = "expected %s to be a non empty string"
)
