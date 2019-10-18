package gamma

import (
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
        Par, sep string
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
    return
}

func StringToVal(v ref.Value, kind ref.Kind, in string) error {
    switch kind {
    case ref.Int:
        set, err := conv.Atoi(in)
        if err != nil {
            return Handle(err, "failed to parse '%s' into an int", in)
        }
        
        v.SetInt(int64(set))
    case ref.Float32:
        set, err := conv.ParseFloat(in, 32)
        
        if err != nil {
            return Handle(err, "failed to parse '%s' into a float32", in)
        }
        
        v.SetFloat(set)
    case ref.Float64:
        set, err := conv.ParseFloat(in, 64)
        
        if err != nil {
            return Handle(err, "failed to parse '%s' into a float64", in)
        }
        
        v.SetFloat(set)
    case ref.Bool:
        v.SetBool(true)
    case ref.String:
        v.SetString(in)
    }
    return nil
}

func (h *Args) ParseStruct(s interface{}) error {
    vv := ref.ValueOf(s)
    kind := vv.Kind()
    
    if kind != ref.Ptr {
        return Handle(nil, "expected a pointer to struct not '%v'", kind)
    }
    
    v := vv.Elem()
    t := v.Type()
    
    
    for ii := 0; ii < v.NumField(); ii++ {
        field := t.Field(ii)
        tag, sfield := field.Tag, v.Field(ii)
        kind := field.Type.Kind()
        
        if tag == "" {
            continue
        }
        
        pos := tag.Get("pos")
        npos := len(pos)

        if npos > 0 {
            idx, err := conv.Atoi(pos)
            
            if err != nil {
                return Handle(err, "failed to parse '%s' into an int", pos)
            }
            
            if idx > npos {
                return Handle(nil, 
                    "index %d is out of the bounds of positional arguments",
                    idx)
            }
            
            if err = StringToVal(sfield, kind, h.pos[idx]); err != nil {
                return Handle(err, "failed to set struct field")
            }
        }
        
        name := tag.Get("name")
        
        if name == "" || name == "-" {
            name = field.Name
        }
        
        val, ok := h.opt[name]
        
        if !ok {
            val = tag.Get("default")
        }
        
        if err := StringToVal(sfield, kind, val); err != nil {
            return Handle(err, "failed to set struct field")
        }
    }
    
    return nil
}


const cmdErr = `Command '%v' failed!
Output of command is: %v
%w`

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
            return "", fmt.Errorf(cmdErr, cmd, result, err)
        }

        return result, nil
    }
}

func NewReader(path string) (ret FileReader, err error) {
    ret.File, err = os.Open(path)

    if err != nil {
        err = Handle(err, "Could not open file '%s'!", path)
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
        err = Handle(err, "failed to open file '%s'", path)
        return
    }

    defer f.Close()

    contents, err := io.ReadAll(f)
    if err != nil {
        err = Handle(err, "failed to read file '%s'!", path)
        return
    }

    return contents, nil
}

func FromString(params, sep string) Params {
    return Params{Par: "", sep: sep, contents: str.Split(params, "\n")}
}

func (self *Params) Param(name string) (ret string, err error) {
    if self.contents == nil {
        var file *os.File
        file, err = os.Open(self.Par)

        if err != nil {
            err = Handle(err, "failed to open file '%s'", self.Par)
            return
        }

        defer file.Close()
        scanner := bio.NewScanner(file)

        for scanner.Scan() {
            line := scanner.Text()
            if str.Contains(line, name) {
                return str.Trim(str.Split(line, self.sep)[1], " "), nil
            }
        }
    } else {
        for _, line := range self.contents {
            if str.Contains(line, name) {
                return str.Trim(str.Split(line, self.sep)[1], " "), nil
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
        err = Handle(err, "failed to convert string '%s' to int", data)
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
        err = Handle(err, "failed to convert string '%s' to float", data)
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
