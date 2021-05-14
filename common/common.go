package common

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/bozso/gotoolbox/errors"
	"github.com/bozso/gotoolbox/path"
)

const DefaultCachePath = "/mnt/bozso_i/cache"

type (
	Slice []string

	settings struct {
		RasExt  string
		Path    string
		Modules []string
	}
)

const (
	BufSize = 50
)

var (
	Pols = [4]string{"vv", "hh", "hv", "vh"}

	confpath = getConfigPath()

	// TODO: get settings path from environment variable
	Settings = loadSettings(confpath)
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

	f, err := path.New(s).Join("gomma.json").ToValidFile()
	Check(err)

	return
}

func loadSettings(file path.ValidFile) (ret settings) {
	if err := LoadJson(file, &ret); err != nil {
		log.Fatalf("Failed to load Gamma settings from '%s'\nError:'%s'\n!",
			file, err)
	}

	return
}

func isClose(num1, num2 float64) bool {
	return math.RoundToEven(math.Abs(num1-num2)) > 0.0
}

func (sl Slice) Contains(s string) bool {
	for _, elem := range sl {
		if s == elem {
			return true
		}
	}
	return false
}

type SavePather interface {
	SavePath(ext string) path.Like
}

func SaveJson(val SavePather) (err error) {
	return SaveJsonTo(val.SavePath("json"), val)
}

func SaveJsonTo(pth path.Like, val interface{}) (err error) {
	f, err := os.Create(pth.String())
	if err != nil {
		return
	}
	defer f.Close()

	out, err := json.MarshalIndent(val, "", "    ")
	if err != nil {
		return errors.WrapFmt(err, "failed to json encode struct: %v", val)
	}

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
	p          path.Pather
	toRetreive string
	err        error
}

func (e FileParseError) ToRetreive(s string) error {
	e.toRetreive = s
	return e
}

func (e FileParseError) Error() string {
	str := fmt.Sprintf("failed to parse file '%s'", e.p.AsPath())

	if tr := e.toRetreive; len(tr) > 0 {
		str = fmt.Sprintf("%s to retreive %s", str, tr)
	}

	return fmt.Sprintf("%s\n%s", e.err, str)
}

func (e FileParseError) Unwrap() error {
	return e.err
}

type Pather interface {
	Path() path.ValidFile
}
