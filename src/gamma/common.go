package gamma;

import (
    // "fmt";
    fp "path/filepath";
    "time";
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
        
        Check(err);
        
        for _, path := range glob {
            result[fp.Base(path)] = MakeCmd(path);
        }
        
        _path = fp.Join(Path, module, "scripts", "*")
        
        glob, err = fp.Glob(_path)
        
        Check(err);
        
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
    Check(err);
    return ret;
}





func First() string {
    return "First";
}