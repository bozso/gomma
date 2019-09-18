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
)

type (
	CmdFun     func(args ...interface{}) (string, error)
	handlerFun func(err error, format string, args ...interface{}) error
	Joiner     func(args ...string) string

	Params struct {
		par, sep string
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

const cmdErr = `Command '%v' failed!
Output of command is: %v
%w`

var tmp = Tmp{}

func Fatal(err error, format string, args ...interface{}) {
	if err != nil {
		str := fmt.Sprintf(format, args...)
		log.Fatalf("Error: %s\nError: %s", str, err)
	}
}

func Empty(str string) bool {
    return len(str) == 0
}

func Handler(name string) handlerFun {
	name = fmt.Sprintf("In %s", name)

	return func(err error, format string, args ...interface{}) error {
		str := fmt.Sprintf(format, args...)

		if err == nil {
			return fmt.Errorf("%s: %s\n", name, str)
		} else {
			return fmt.Errorf("%s: %s\nError: %w", name, str, err)
		}
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

		out, err := exec.Command(cmd, arg...).CombinedOutput()
		result := string(out)

		if err != nil {
			return "", fmt.Errorf(cmdErr, cmd, result, err)
		}

		return result, nil
	}
}

func NewPath(args ...string) path {
	return path{fp.Join(args...), args}
}

func (self *path) Join(args ...string) path {
	newpath := append(self.parts, args...)
	return path{fp.Join(newpath...), newpath}
}

func (self *path) Glob() ([]string, error) {
	ret, err := fp.Glob(self.path)

	if err != nil {
		return ret,
			fmt.Errorf("In path.Glob: Could not get Glob of: '%s'",
				self.path)
	}
	return ret, nil
}

func (self *path) Info() (os.FileInfo, error) {
	ret, err := os.Stat(self.path)

	if err != nil {
		return ret,
			fmt.Errorf("In path.Info: Could not get FileInfo of: '%s'",
				self.path)
	}
	return ret, nil
}

func Exist(path string) (ret bool, err error) {
    _, err = os.Stat(path)
    
    if err != nil {
        if os.IsNotExist(err) {
            return false, nil
        }
        return false, err
    }
    return true, nil
}

func ReadFile(path string) (ret []byte, err error) {
	handle := Handler("ReadFile")

	f, err := os.Open(path)
	if err != nil {
		err = handle(err, "Could not open file: '%v'!", path)
        return
	}

	defer f.Close()

	contents, err := io.ReadAll(f)
	if err != nil {
		 err = handle(err, "Could not read file: '%v'!", path)
         return
	}

	return contents, nil
}

func FromString(params, sep string) Params {
    return Params{par:"", sep:sep, contents: str.Split(params, "\n")}
}

func (self *Params) Par(name string) (ret string, err error) {
    handle := Handler("Params.Par")
    
    if self.contents == nil {
        var file *os.File
        file, err = os.Open(self.par)
        
        if err != nil {
            err = handle(err, "Could not open file: '%s'!", self.par)
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
    
    
	err = fmt.Errorf("In Par: Could not find parameter '%s' in %v",
		name, self.par)
    return
}

func toInt(par string, idx int) (int, error) {
	ret, err := conv.Atoi(str.Split(par, " ")[idx])

	if err != nil {
		return 0,
			fmt.Errorf("In toInt: Could not convert string %s to float64!\nError: %w",
				par, err)
	}

	return ret, nil
}

func toFloat(par string, idx int) (float64, error) {
	ret, err := conv.ParseFloat(str.Split(par, " ")[idx], 64)

	if err != nil {
		return 0.0,
			fmt.Errorf("Could not convert string %s to float64!\nError: %w",
				par, err)
	}

	return ret, nil
}

func (self *Params) Int(name string) (int, error) {
	data, err := self.Par(name)

	if err != nil {
		return 0, err
	}

	return toInt(data, 0)
}

func (self *Params) Float(name string) (float64, error) {
	data, err := self.Par(name)

	if err != nil {
		return 0.0, err
	}

	return toFloat(data, 0)
}

func TmpFile() (ret string, err error) {
	file, err := io.TempFile("", "*")

	if err != nil {
		err = fmt.Errorf(
            "In TmpFile: Failed to create a temporary file!\nError: %w", err)
        return
	}

	defer file.Close()

	name := file.Name()

	tmp.files = append(tmp.files, name)

	return name, nil
}

func TmpFileExt(ext string) (string, error) {
	file, err := io.TempFile("", "*." + ext)

	if err != nil {
		return "", fmt.Errorf(
            "In TmpFileExt: Failed to create a temporary file!\nError: %w",
            err)
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
			log.Printf("Failed to remove temporary file '%s': %w\n", file, err)
		}
	}
}
