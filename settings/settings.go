package settings

import (
	"fmt"

	"github.com/bozso/gotoolbox/enum"
	"github.com/bozso/gotoolbox/path"

	"git.st.ht/~istvan_bozso/shutil/command"

	"git.st.ht/~istvan_bozso/sert/log"
)

var (
	validExtensions = enum.NewStringSet("bmp", "ras", "tif")
	validModules    = enum.NewStringSet(
		"DIFF", "DISP", "ISP", "LAT", "IPTA")
	Extensions = validExtensions.EnumType("RasterExtension")
)

type RasterExtension string

func (r *RasterExtension) UnmarshalJSON(b []byte) (err error) {
	*r = RasterExtension(string(b))
	err = r.Validate()
	return
}

func (r RasterExtension) Validate() (err error) {
	v := string(r)

	if !validExtensions.Contains(v) {
		err = Extensions.UnknownElement(v)
	}

	return
}

type Common struct {
	RasterExtension RasterExtension `json:"raster_extension"`
	GammaDirectory  path.Dir        `json:"gamma_directory"`
	CachePath       path.Dir        `json:"cache_path"`
	Logging         log.Config      `json:"logging"`
}

type Payload struct {
	Common
	Modules  []string     `json:"modules"`
	Executor json.Payload `json:"executor"`
}

func (p Payload) ToSettings() (s Settings, err error) {
	s.Common = p.Common
	s.Executor, err = command.FromPayload(p.Executor)
	if err != nil {
		return
	}

	s.Commands, err = p.MakeCommands()
	if err != nil {
		return
	}

	s.Logger, err = p.Logging.Create()
	return
}

type Settings struct {
	Common
	Executor command.Executor
	Commands Commands
	Logger   log.Logger
}

var exeDirectories = [...]string{"bin", "scripts"}

func (p Payload) MakeCommands() (c Commands, err error) {
	gammaDir := p.GammaDirectory
	c = make(Commands)

	for _, module := range s.Modules {
		for _, dir := range exeDirectories {
			glob, err := gammaDir.Join(module, dir, "*").Glob()

			if err != nil {
				return c, err
			}

			for _, exePath := range glob {
				c[exePath.Base().String()], err = command.New(exePath)
				if err != nil {
					return c, err
				}
			}
		}
	}

	return
}
