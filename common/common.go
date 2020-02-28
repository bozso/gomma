package common

import (
    "fmt"
    "log"
    "os"
    "math"
    "path"
    "strings"
    "encoding/json"
    "path/filepath"
    
    "github.com/bozso/gamma/utils"
    "github.com/bozso/gamma/utils/io"
    "github.com/bozso/gamma/utils/command"
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
    
    // TODO: get settings path from environment variable
    Settings = loadSettings("/home/istvan/progs/gamma/bin/settings.json")
    commands = makeCommands()
)

func Must(name string) (c Command) {
    return commands.Must(name)
}

func Select(name1, name2 string) (c Command) {
    return commands.Select(name1, name2)
}

func loadSettings(path string) (ret settings) {
    if err := LoadJson(path, &ret); err != nil {
        log.Fatalf("Failed to load Gamma settings from '%s'\nError:'%s'\n!",
            path, err)
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
                utils.Fatal(err, "Glob '%s' failed! %s", _path, err)
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


func NoExt(p string) string {
    return strings.TrimSuffix(p, path.Ext(p))
}

func isclose(num1, num2 float64) bool {
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


func SaveJson(path string, val interface{}) (err error) {
    out, err := json.MarshalIndent(val, "", "    ")
    if err != nil {
        return utils.WrapFmt(err, "failed to json encode struct: %v", val)
    }

    f, err := os.Create(path)
    if err != nil {
        return io.CreateFail(path, err)
    }
    defer f.Close()

    if _, err = f.Write(out); err != nil {
        return io.WriteFail(path, err)
    }

    return nil
}

type Validator interface {
    Validate() error
}

func LoadJson(path string, val interface{}) (err error) {
    d, err := io.ReadFile(path)
    if err != nil {
        return utils.WrapFmt(err, "failed to read file '%s'", path)
    }
    
    if err := json.Unmarshal(d, &val); err != nil {
        return utils.WrapFmt(err, "failed to parse json data %s'", d)
    }
    
    v, ok := val.(Validator)
    
    if !ok {
        return nil
    }
    
    err = v.Validate()
    
    return
}
