package gamma;

import (
    //"fmt";
    fp "path/filepath";
    str "strings";
    "time";
    "os";
    zip "archive/zip";
    set "github.com/deckarep/golang-set"
    conv "strconv";
);


const (
    useVersion = "20181130";
    BufSize = 50;
)


var versions = map[string]string {
    "20181130": "/home/istvan/progs/GAMMA_SOFTWARE-20181130",
};


var (
    pols = []interface{}{"vv", "hh", "hv", "vh"};
    Pols = set.NewSetFromSlice(pols);
    DataTypes = map[string]int {
        "FCOMPLEX": 0,
        "SCOMPLEX": 1,
        "FLOAT": 0,
        "SHORT_INT": 1,
        "DOUBLE": 2,
    };
);



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
func (self *ParamFile) Param(name string) string {
    return "implement";
}

func toInt(par string, idx int) int {
    ret, err := conv.Atoi(str.Split(par, " ")[idx]);
    Check(err, "Could not convert string %s to int!", par);
    return ret;
}

func toFloat(par string, idx int) float64 {
    ret, err := conv.ParseFloat(str.Split(par, " ")[idx], 64);
    Check(err, "Could not convert string %s to float64!", par);
    return ret;
}

func (self *ParamFile) IntPar(name string) int {
    return toInt(self.Param(name), 0);
}

func (self *ParamFile) FloatPar(name string) float64 {
    return toFloat(self.Param(name), 0);
}

type dataFile struct {
    dat string;
    ParamFile;
    date;
}

func (self *dataFile) Rng() int {
    return self.IntPar("range_samples");
}


func (self *dataFile) Azi() int {
    return self.IntPar("azimuth_samples");
}


func (self *dataFile) imgFormat() string {
    return self.Param("image_format");
}


func (self dataFile) Date() time.Time {
    return self.center;
}


type Extract struct {
    file *zip.ReadCloser;
    fileList []string;
};


func (self *Extract) Close() {
    self.file.Close();
}


type Extracted struct {
    fileSet set.Set;
    root string;
}

func NewExtract(path string, templates []string, root string) Extract {
    file, err := zip.OpenReader(path);
    defer file.Close()
    
    Check(err, "Could not open zipfile: \"%s\"", path);
    
    list := make([]string, BufSize);
    
    for _, file := range file.File {
        name := file.Name;
        if _, err := os.Stat(fp.Join(root, name)); err == nil {
            // TODO: Check if matches template.
            if os.IsNotExist(err) {
                list = append(list, name);
            } else {
                Check(err, "Stat failed on file : \"%s\"", file);
            }
        }
    }
    
    return Extract{file, list};
}

//func (self Extract) Filter(extracted []string)
    

type point struct {
    x, y float64;
};

type rect struct {
    max, min point;
};


func pointInRect(p point, r rect) bool {
    return (p.x < r.max.x && p.x > r.min.x &&
            p.y < r.max.y && p.y > r.min.y);
}


func First() string {
    return "First";
}

type disArgs struct {
    flip, debug bool;
    rng, azi int;
    imgFormat, datfile string;
};

type rasArgs struct {
    disArgs;
    ext string;
    avgFact int;
};

func ParseDisArgs(d dataFile, args disArgs) disArgs {
    
    if len(args.datfile) == 0 {
        args.datfile = d.dat;
    }
    
    if args.rng == 0 {
        args.rng = d.Rng();
    }
    
    if args.azi == 0 {
        args.azi = d.Azi();
    }
    
    // parts = pth.basename(datfile).split(".")
    if len(args.imgFormat) == 0 {
        args.imgFormat = d.imgFormat();
    }
    
    // args.flip = -1 if flip else 1
    
    /*
    if cmd is None:
        try:
            ext = [ext for ext in parts if ext in extensions][0]
        except IndexError:
            raise ValueError("Unrecognized extension of file %s. Available "
                             "extensions: %s" % (datfile, pr.extensions))

        cmd = [cmd for cmd, exts in plot_cmd_files.items()
               if ext in exts][0]
    */
    
    return args;
}

/*

def date2str(obj, fmt="%Y%m%d"):
    date = obj.date
    
    if isinstance(date, Date):
        date = date.center
    
    return date.strftime(fmt)


@staticmethod
def parse_dis_args(gp_file, **kwargs):
    datfile = kwargs.get("datfile", None)
    cmd  = kwargs.get("mode", None)
    flip = bool(kwargs.get("flip", False))
    rng = kwargs.get("rng", None)
    azi = kwargs.get("azi", None)
    img_fmt = kwargs.get("image_format", None)
    debug = bool(kwargs.get("debug", False))    
    
    if datfile is None:
        datfile = gp_file.dat

    parts = pth.basename(datfile).split(".")
    

    if rng is None:
        rng = gp_file.rng()

    if azi is None:
        azi = gp_file.azi()
    
    if img_fmt is None:
        img_fmt = gp_file.img_fmt()
    
    
    flip = -1 if flip else 1
    

    if cmd is None:
        try:
            ext = [ext for ext in parts if ext in extensions][0]
        except IndexError:
            raise ValueError("Unrecognized extension of file %s. Available "
                             "extensions: %s" % (datfile, pr.extensions))

        cmd = [cmd for cmd, exts in plot_cmd_files.items()
               if ext in exts][0]
    
    
    return {
        "cmd"      : cmd,
        "datfile"  : datfile,
        "rng"      : rng,
        "azi"      : azi,
        "img_fmt"  : DataFile.data_types[img_fmt.upper()],
        "start"    : kwargs.get("start", None),
        "nlines"   : kwargs.get("nlines", None),
        "scale"    : kwargs.get("scale", None),
        "exp"      : kwargs.get("exp", None),
        "LR"       : int(flip),
        "debug"    : debug
    }


@staticmethod
def parse_ras_args(gp_file, **kwargs):
    args = DataFile.parse_dis_args(gp_file, **kwargs)
    
    raster = kwargs.get("raster", None)
    avg_fact = kwargs.get("avg_fact", 750)
    
    if raster is None:
        raster = "%s.%s" % (args["datfile"], settings["ras_ext"])

    if avg_fact == "noavg":
        avg_rng, avg_azi = None, None
    else:
        avg_rng, avg_azi = gp_file.avg_fact(avg_fact)

    
    args.update({
        "raster" : raster,
        "arng": avg_rng,
        "aazi": avg_azi,
        "hdrsz": int(kwargs.get("hdrsz", 0))
        })

    return args


def raster(self, **kwargs):
    args = DataFile.parse_ras_args(self, **kwargs)
    
    cmd = args["cmd"]
    ras = getattr(gp, "ras" + cmd)
    
    if cmd == "SLC":
        ras\
        (args["datfile"], args["rng"], args["start"], args["nlines"],
         args["arng"], args["aazi"], args["scale"], args["exp"], args["LR"],
         args["img_fmt"], args["hdrsz"], args["raster"],
         debug=args["debug"])
    else:
        sec = kwargs.pop("sec", None)
        
        if sec is None:
            ras\
            (args["datfile"], args["rng"], args["start"], args["nlines"],
             args["arng"], args["aazi"], args["scale"], args["exp"], args["LR"],
             args["raster"], args["img_fmt"], args["hdrsz"],
             debug=args["debug"])
        else:
            ras\
            (args["datfile"], sec, args["rng"], args["start"], args["nlines"],
             args["arng"], args["aazi"], args["scale"], args["exp"], args["LR"],
             args["raster"], args["img_fmt"], args["hdrsz"],
             debug=args["debug"])
    
    self.ras = args["raster"]


*/