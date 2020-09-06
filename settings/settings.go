package settings

import (
    "github.com/bozso/gotoolbox/path"

    "github.com/bozso/gomma/command"
)

type Settings struct {
    RasterExtension string   `json:"raster_extension"`
    GammaDirectory  path.Dir `json:"gamma_directory"`
    Modules         []string `json:"modules"`
    CachePath       path.Dir `json:"cache_path"`
}

func (s *Settings) SetCachePath(cachePath string) (err error) {
    s.CachePath, err = path.New(cachePath).Mkdir()
    return
}

func (s *Settings) Default() (err error) {
    return s.SetCachePath(".")
}

var exeDirectories = [2]string{"bin", "scripts"}

func (s Settings) MakeCommands() (c command.Commands, err error) {
    gammaDir := s.GammaDirectory
    c = make(command.Commands)

    for _, module := range s.Modules {
        for _, dir := range exeDirectories {
            glob, err := gammaDir.Join(module, dir, "*").Glob()

            if err != nil {
                return c, err
            }

            for _, exePath := range glob {
                c[exePath.Base().String()] = command.New(exePath.String())
            }
        }
    }

    return
}
