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
}

type Payload struct {
    Common
    Modules []string               `json:"modules"`
    Executor command.ExecutorSetup `json:"executor"`
}

type Settings struct {
    Common
    executor        command.Executor
    Commands        Commands
}

func (s *Settings) Update(p Payload) (er error) {

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

func (s Settings) MakeCommands(create command.Creator) (c Commands, err error) {
    gammaDir := s.GammaDirectory
    c = make(Commands)

    for _, module := range s.Modules {
        for _, dir := range exeDirectories {
            glob, err := gammaDir.Join(module, dir, "*").Glob()

            if err != nil {
                return c, err
            }

            for _, exePath := range glob {
                com, err := create.Create(exePath.String())
                if err != nil {
                    return c, err
                }

                c[exePath.Base().String()] = com
            }
        }
    }

    return
}
