package settings

import (
    "fmt"
    
    "github.com/bozso/gotoolbox/path"
    "github.com/bozso/gotoolbox/enum"

    "github.com/bozso/gomma/command"
)

var (
    validExtensions = enum.NewStringSet("bmp", "ras", "tif")
    validModules = enum.NewStringSet(
        "DIFF", "DISP", "ISP", "LAT", "IPTA")
)

type Setup struct {
    RasterExtension string   `json:"raster_extension"`
    GammaDirectory  string `json:"gamma_directory"`
    Modules         []string `json:"modules"`
    CachePath       string `json:"cache_path"`    
}

type Settings struct {
    RasterExtension string   `json:"raster_extension"`
    GammaDirectory  path.Dir `json:"gamma_directory"`
    Modules         []string `json:"modules"`
    CachePath       path.Dir `json:"cache_path"`
}

func (s Setup) New() (st Settings, err error) {
    if err = st.SetRasterExtension(s.RasterExtension); err != nil {
        return
    }

    if err = st.SetGammaDirectory(s.GammaDirectory); err != nil {
        return
    }

    if err = st.SetModules(s.Modules); err != nil {
        return
    }

    err = st.SetCachePath(s.CachePath)
    return
}


func (s *Settings) SetModules(modules []string) (err error) {
    s.Modules = modules
    return
}

func (s *Settings) SetRasterExtension(ext string) (err error) {
    if !validExtensions.Contains(ext) {
        err = fmt.Errorf("expected either '%s', got '%s'",
            validExtensions, ext)
        return
    }
    
    s.RasterExtension = ext
    return
}

func (s *Settings) SetGammaDirectory(gammaPath string) (err error) {
    s.GammaDirectory, err = path.New(gammaPath).Mkdir()
    return    
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
