package gamma;

import (
    //"fmt";
    fp "path/filepath";
    str "strings";
    "time";
    zip "archive/zip";
    conv "strconv";
);


type setting map[string]string;

const (
    useVersion = "20181130";
    BufSize = 50;
)


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
        for _, dir := range [2]string{"bin", "scripts"} {
            _path := fp.Join(Path, module, dir, "*")
            glob, err := fp.Glob(_path)
            
            Check(err, "Glob in %s failed!", _path);
            
            for _, path := range glob {
                result[fp.Base(path)] = MakeCmd(path);
            }
        }
        
        
        /*
        _path = fp.Join(Path, module, "scripts", "*")
        
        glob, err = fp.Glob(_path)
        
        Check(err, "Glob in %s failed!", _path);
        
        for _, path := range glob {
            result[fp.Base(path)] = MakeCmd(path);
        }
        */
    }
    
    return result;
}



var (
    Gamma = makeGamma();
    Imv = MakeCmd("eog");
);


type date struct {
    start, stop, center time.Time;
};

type Date interface {
    Date() time.Time;
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
    IntPar() int;
    FloatPar() float32;
    Param() string;
};


type ParamFile struct {
    par string;
};


// TODO: implement
func (self ParamFile) Param(name string) string {
    return "implement";
}

func toInt(par string, idx int) int {
    ret, err := conv.Atoi(str.Split(par, " ")[idx]);
    Check(err, "Could not convert string to int!");
    return ret;
}

func toFloat(par string, idx int) float64 {
    ret, err := conv.ParseFloat(str.Split(par, " ")[idx], 64);
    Check(err, "Could not convert string to float64!");
    return ret;
}

func (self ParamFile) IntPar(name string) int {
    return toInt(self.Param(name), 0);
}

func (self ParamFile) FloatPar(name string) float64 {
    return toFloat(self.Param(name), 0);
}

type dataFile struct {
    dat string;
    ParamFile;
    date;
}

func (self dataFile) Rng() int {
    return self.IntPar("range_samples");
}


func (self dataFile) Azi() int {
    return self.IntPar("azimuth_samples");
}


func (self dataFile) Date() time.Time {
    return self.center;
}


type Extract struct {
    file *zip.ReadCloser;
    fileList []string;
};


func NewExtract(path string, templates []string) Extract {
    file, err := zip.OpenReader(path);
    defer file.Close()
    
    Check(err, "Could not open zipfile: \"%s\"", path);
    
    list := make([]string, BufSize);
    
    
    for _, file := range file.File {
        // TODO: select if matches template
        list = append(list, file.Name);
    }
    
    return Extract{file, list}
}

func (self Extract) Close() {
    self.file.Close();
}


//func (self Extract) Filter(extracted []string)
    



func First() string {
    return "First";
}