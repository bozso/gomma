package gamma;

import (
    //"fmt";
    fp "path/filepath";
    //str "strings";
    "time";
    zip "archive/zip";
);

type setting map[string]string;

const useVersion = "20181130";


var versions = map[string]string {
    "20181130": "/home/istvan/progs/GAMMA_SOFTWARE-20181130",
};



type templates struct {
    IW string;
    Tab string;
};


type settings struct {
    RasExt string;
    path string;
    modules []string;
    Templates templates;
};


var Settings = settings{
    RasExt: "bmp",
    path: versions[useVersion],
    modules: []string{"DIFF", "DISP", "ISP", "LAT", "IPTA"},
    Templates: templates{
        IW: "{{date}}_iw{{iw}}.{{pol}}.slc",
        Tab: "{{date}}.{{pol}}.SLC_tab",
    },
};


func makeGamma() map[string]CmdFun {
    Path := Settings.path;
    
    result := make(map[string]CmdFun);
    
    for _, module := range Settings.modules {
        _path := fp.Join(Path, module, "bin", "*")
        
        glob, err := fp.Glob(_path)
        
        Check(err, "Glob in %s failed!", _path);
        
        for _, path := range glob {
            result[fp.Base(path)] = MakeCmd(path);
        }
        
        _path = fp.Join(Path, module, "scripts", "*")
        
        glob, err = fp.Glob(_path)
        
        Check(err, "Glob in %s failed!", _path);
        
        for _, path := range glob {
            result[fp.Base(path)] = MakeCmd(path);
        }
    }
    
    return result;
}

var (
    Gamma = makeGamma();
    Imv = MakeCmd("eog");
);



type Date struct {
    start, stop, center time.Time;
};


const (
    DateShort = "20060102";
    DateLong = "20060102T150405";
);


func ParseDate(fmt string, str string) time.Time {
    ret, err := time.Parse(fmt, str)
    Check(err, "Failed to parse date: %s", str);
    return ret;
}


type DataFile interface {
    Rng() int;
    Azi() int;
    Intpar() int;
    Floatpar() float32;
    Param() string;
};


type ParamFile struct {
    par string;
};


// TODO: implement
func (self ParamFile) Param(name string) string {
    return "implement";
}

// TODO: implement
func toInt(str string, idx int) int {
    var ret int;
    return ret;
}

// TODO: implement
func toFloat(str string, idx int) float32 {
    var ret float32;
    return ret;
}

func (self ParamFile) Intpar(name string) int {
    return toInt(self.Param(name), 0);
}

func (self ParamFile) Floatpar(name string) float32 {
    return toFloat(self.Param(name), 0);
}

type dataFile struct {
    dat string;
    ParamFile;
    Date;
}

func (self dataFile) Rng() int {
    return self.Intpar("range_samples");
}


func (self dataFile) Azi() int {
    return self.Intpar("azimuth_samples");
}


type Extract struct {
    File *zip.ReadCloser;
    FileList []string;
};


func NewExtract(path string, templates []string) Extract {
    file, err := zip.OpenReader(path);
    defer file.Close()
    
    Check(err, "Could not open zipfile: \"%s\"", path);
    
    list := make([]string, 10);
    
    
    for ii, file := range file.File {
        // TODO: select if matches template
        list = append(list, file);
    }
    
    return Extract{file, list}
}

//func (self Extract) Filter(extracted []string)
    



func First() string {
    return "First";
}