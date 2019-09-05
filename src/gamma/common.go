package gamma;

import (
    "fmt";
    fp "path/filepath";
    "time";
    "os";
    zip "archive/zip";
    set "github.com/deckarep/golang-set"
);


const (
    useVersion = "20181130";
    BufSize = 50;
    DateShort = "20060102";
    DateLong = "20060102T150405";
)


type pol int;

const (
    VV pol = iota;
    HH;
    HV;
    VH;
);


var (
    versions = map[string]string {
        "20181130": "/home/istvan/progs/GAMMA_SOFTWARE-20181130",
    };
    
    pols = []interface{}{"vv", "hh", "hv", "vh"};
    Pols = set.NewSetFromSlice(pols);
    
    DataTypes = map[string]int {
        "FCOMPLEX": 0,
        "SCOMPLEX": 1,
        "FLOAT": 0,
        "SHORT_INT": 1,
        "DOUBLE": 2,
    };
    Gamma = makeGamma();
    Imv = MakeCmd("eog");
);


type(
    templates struct {
        IW string;
        Tab string;
    };

    settings struct {
        RasExt string;
        path string;
        modules []string;
        Templates templates;
    };
    
    date struct {
        start, stop, center time.Time;
    };

    Date interface {
        Date() time.Time;
    };
    
    dataFile struct {
        dat string;
        Params;
        date;
    };
    
    
    DataFile interface {
        Rng() int;
        Azi() int;
        IntPar() int;
        FloatPar() float32;
        Param() string;
    };


    ParamFile struct {
        par string;
    };
    
    
    Extract struct {
        file *zip.ReadCloser;
        fileList []string;
    };
    
    
    Extracted struct {
        fileSet set.Set;
        root string;
    };
    
    
    point struct {
        x, y float64;
    };

    
    rect struct {
        max, min point;
    };
    
    
    disArgs struct {
        flip, debug bool;
        rng, azi int;
        imgFormat, datfile string;
    };

    
    rasArgs struct {
        disArgs;
        ext string;
        avgFact int;
    };
);


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
            glob, err := fp.Glob(_path);
            
            if err != nil {
                Fatal(err, "makeGamma: Glob '%s' failed! %w", _path, err);
            }
            
            for _, path := range glob {
                result[fp.Base(path)] = MakeCmd(path);
            }
        }
    }
    
    return result;
}


func ParseDate(format string, str string) (time.Time, error) {
    
    ret, err := time.Parse(format, str);
    
    if err != nil {
        return time.Time{}, 
               fmt.Errorf("In ParseDate: Failed to parse date: %s!\nError: %w", 
                           str, err)
    }
    
    return ret, nil;
}


func (self *dataFile) Rng() (int, error) {
    return self.Int("range_samples");
}


func (self *dataFile) Azi() (int, error) {
    return self.Int("azimuth_samples");
}


func (self *dataFile) imgFormat() (string, error) {
    return self.Par("image_format");
}


func (self dataFile) Date() time.Time {
    return self.center;
}


func NewExtract(path string, templates []string, root string) (*Extract, error) {
    handle := Handler("NewExtract");
    
    file, err := zip.OpenReader(path);
    
    if err != nil {
        return nil, handle(err, "Could not open zipfile: '%s'!", path);
    }
    
    defer file.Close()
    
    list := make([]string, BufSize);
    
    for _, file := range file.File {
        name := file.Name;
        if _, err := os.Stat(fp.Join(root, name)); err == nil {
            // TODO: Check if matches templates.
            if os.IsNotExist(err) {
                list = append(list, name);
            } else {
                return nil, handle(err, "Stat failed on file : '%s'!", file);
            }
        }
    }
    
    return &Extract{file, list}, nil;
}


func (self *Extract) Close() {
    self.file.Close();
}


func (self *point) inRect(r *rect) bool {
    return (self.x < r.max.x && self.x > r.min.x &&
            self.y < r.max.y && self.y > r.min.y);
}


func First() string {
    return "First";
}


func ParseDisArgs(d dataFile, args disArgs) (*disArgs, error) {
    var err error;
    handle := Handler("ParseDisArgs");
    
    if len(args.datfile) == 0 {
        args.datfile = d.dat;
    }
    
    
    if args.rng == 0 {
        if args.rng, err = d.Rng(); err != nil {
            return nil, handle(err, "Could not get range_samples!");
        }
    }
    
    
    if args.azi == 0 {
        if args.azi, err = d.Azi(); err != nil {
            return nil, handle(err, "Could not get azimuth_lines!");
        }
    }
    
    
    // parts = pth.basename(datfile).split(".")
    if len(args.imgFormat) == 0 {
        if args.imgFormat, err = d.imgFormat(); err != nil {
            return nil, handle(err, "Could not get image_format!");
        }
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
    
    return &args, nil;
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