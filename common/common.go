package common

import (
    "os"
    "fmt"
    "log"
    "math"
    "encoding/json"
    "path/filepath"
    
    "github.com/bozso/gotoolbox/command"
    "github.com/bozso/gotoolbox/errors"
    "github.com/bozso/gotoolbox/path"
)

const DefaultCachePath = "/mnt/bozso_i/cache"

type (
    Slice []string
    Commands map[string]Command

    settings struct {
        RasExt    string
        Path      string
        Modules   []string
    }
)

const (
    BufSize    = 50
)

var (
    Pols = [4]string{"vv", "hh", "hv", "vh"}
    
    confpath = getConfigPath()
    
    // TODO: get settings path from environment variable
    Settings = loadSettings(confpath)
    commands = makeCommands()
)

func Check(err error) {
    if err != nil {
        log.Fatalf("%s\n", err)
    }
}

func getConfigPath() (f path.ValidFile) {
    s, ok := os.LookupEnv("GOMMA_CONFIG")
    
    if !ok {
        var err error
        s, err = os.UserConfigDir()
        Check(err)
        return
    }
    
    f, err := path.New(s).Join("gomma.json").ToFile().ToValid()
    Check(err)
    
    return
}

func Must(name string) (c Command) {
    return commands.Must(name)
}

func Select(name1, name2 string) (c Command) {
    return commands.Select(name1, name2)
}

func loadSettings(file path.ValidFile) (ret settings) {
    if err := LoadJson(file, &ret); err != nil {
        log.Fatalf("Failed to load Gamma settings from '%s'\nError:'%s'\n!",
            file, err)
    }
    
    return
}

type Command struct {
    command.Command
}

func (c Command) Call(args ...interface{}) (s string, err error) {
    arg := make([]string, len(args))

    for ii, elem := range args {
        if elem == nil {
            arg[ii] = "-"
        } else {
            arg[ii] = fmt.Sprint(elem)
        }
    }

    return c.Command.CallWithArgs(arg...)
}

func makeCommands() Commands {
    Path := Settings.Path
    result := make(Commands)

    for _, module := range Settings.Modules {
        for _, dir := range [2]string{"bin", "scripts"} {

            _path := filepath.Join(Path, module, dir, "*")
            glob, err := filepath.Glob(_path)

            if err != nil {
                log.Fatal(err, "Glob '%s' failed! %s", _path, err)
            }

            for _, path := range glob {
                result[filepath.Base(path)] = Command{
                        Command:command.New(path),
                    }
            }
        }
    }

    return result
}

func (cs Commands) Select(name1, name2 string) (c Command) {
    c, ok := cs[name1]
    
    if ok {
        return
    }
    
    c, ok = cs[name2]
    
    if !ok {
        log.Fatalf("either '%s' or '%s' must be an available executable",
            name1, name2)
    }
    
    return
}

func (cs Commands) Must(name string) (c Command) {
    c, ok := cs[name]
    
    if !ok {
        log.Fatalf("failed to find Gamma executable '%s'", name)
    }
    
    return
}

func isClose(num1, num2 float64) bool {
    return math.RoundToEven(math.Abs(num1 - num2)) > 0.0
}

func (sl Slice) Contains(s string) bool {
    for _, elem := range sl {
        if s == elem {
            return true
        }
    }
    return false
}


func SaveJson(pth path.File, val interface{}) (err error) {
    out, err := json.MarshalIndent(val, "", "    ")
    if err != nil {
        return errors.WrapFmt(err, "failed to json encode struct: %v", val)
    }
    
    f, err := pth.Create()
    if err != nil { return }
    defer f.Close()

    if _, err = f.Write(out); err != nil {
        return
    }

    return nil
}

type Validator interface {
    Validate() error
}

func LoadJson(f path.ValidFile, val interface{}) (err error) {
    d, err := f.ReadAll()
    if err != nil {
        return errors.WrapFmt(err, "failed to read file '%s'", f)
    }
    
    if err := json.Unmarshal(d, &val); err != nil {
        return errors.WrapFmt(err, "failed to parse json data %s'", d)
    }
    
    v, ok := val.(Validator)
    
    if !ok {
        return nil
    }
    
    err = v.Validate()
    
    return
}

func ParseFail(p path.Pather, err error) FileParseError {
    return FileParseError{p, "", err}
}

type FileParseError struct {
    p path.Pather
    toRetreive string
    err error
}

func (e FileParseError) ToRetreive(s string) error {
    e.toRetreive = s
    return e
}

func (e FileParseError) Error() string {
    str := fmt.Sprintf("failed to parse file '%s'", e.p.GetPath())
    
    if tr := e.toRetreive; len(tr) > 0 {
        str = fmt.Sprintf("%s to retreive %s", str, tr)
    }
    
    return errors.Wrap(e.err, str)
}

func (e FileParseError) Unwrap() error {
    return e.err
}
