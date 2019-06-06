import os
import os.path as pth
import shutil as sh

from sys import version_info
from glob import iglob
from math import sqrt, isclose
from logging import getLogger
from datetime import datetime, timedelta
from errno import ENOENT, EEXIST
from collections import namedtuple
from pprint import pformat
from subprocess import check_output, CalledProcessError, STDOUT
from shlex import split

import gamma as gm

PY3 = version_info[0] == 3

__all__ = [
    "DataFile",
    "SLC",
    "MLI",
    "imview",
    "gnuplot",
    "colors",
    "make_colorbar",
    "string_t",
    "settings",
    "gamma_progs",
    "ScanSAR",
    "montage",
    "settings",
    "display",
    "raster",
]


ScanSAR = True

if PY3:
    string_t = str,
else:
    string_t = basestring,


settings = {
    "ras_ext": "bmp",
    "path": "/home/istvan/progs/GAMMA_SOFTWARE-20181130",
    "modules": ("DIFF", "DISP", "ISP", "LAT", "IPTA"),
    "libpaths": "/home/istvan/miniconda3/lib:"
}


os.environ["LD_LIBRARY_PATH"] = \
os.getenv("LD_LIBRARY_PATH") + settings["libpaths"]


log = getLogger("gamma.base")

gamma_cmaps = pth.join(settings["path"], "DISP", "cmaps")


gamma_commands = \
tuple(binfile for module in settings["modules"]
      for path in ("bin", "scripts")
      for binfile in iglob(pth.join(settings["path"], module, path, "*")))


def make_cmd(command):
    def cmd(*args, **kwargs):
        debug = kwargs.pop("debug", False)
        _log = kwargs.pop("log", None)
        
        Cmd = "%s %s" % (command, " ".join(_proc_arg(arg) for arg in args))
        
        log.debug('Issued command is "%s"' % Cmd)
        
        if debug:
            print(Cmd)
            return
        
        try:
            proc = check_output(split(Cmd), stderr=STDOUT)
        except CalledProcessError as e:
            print("\nNon zero returncode from command: \n'{}'\n"
                  "\nOUTPUT OF THE COMMAND: \n\n{}\nRETURNCODE was: {}"
                  .format(Cmd, e.output.decode(), e.returncode))
    
            raise e
        
        
        if _log is not None:
            with open(_log, "wb") as f:
                f.write(proc)
    
        return proc
    return cmd



# gamma_commands = ("rashgt", "ScanSAR_burst_corners")
    
gamma_progs = type("Gamma", (object,),
                   dict((pth.basename(cmd), staticmethod(make_cmd(cmd)))
                   for cmd in gamma_commands))


gp = gamma_progs


_convert = make_cmd("convert")
_montage = make_cmd("montage")
gnuplot = make_cmd("gnuplot")
imview = make_cmd("eog")

     
class DataFile(gm.Files):
    __slots__ = ("dat", "par", "datpar", "tab", "keep")

    data_types = {
        "FCOMPLEX": 0,
        "SCOMPLEX": 1,
        "FLOAT": 0,
        "SHORT_INT": 1,
        "DOUBLE": 2
    }

    
    def __init__(self, **kwargs):
        self.keep = None
        datfile   = kwargs.get("datfile")
        parfile   = kwargs.get("parfile")
        
        if datfile is None:
            datfile = get_tmp(kwargs.get("tmpdir", tmpdir))
        
        if parfile is None:
            parfile = datfile + ".par"
        
        self.datpar = "%s %s" % (datfile, parfile)
        
        self.dat, self.par, self.tab, self.keep = \
        datfile, parfile, kwargs.get("tabfile", None), \
        bool(kwargs.get("keep", True))

    
    def rm(self):
        Files.rm(self, "dat", "par")
    
    
    def save(self, datfile, parfile=None):
        if parfile is None:
            parfile = datfile + ".par"
        
        self.mv("dat", datfile)
        self.mv("par", parfile)
        
        self.dat, self.par = datfile, parfile
        
        self.keep = True


    def __del__(self):
        keep = self.keep
        
        if keep is not None and not keep:
            self.rm()

    
    def __str__(self):
        return self.datpar


    def __bool__(self):
        return self.exist("dat") and self.exist("par")


    def __getitem__(self, key):
        return self.get("par", key)


    def rng(self):
        return self.getint("par", "range_samples")

    def azi(self):
        return self.getint("par", "azimuth_lines")
    
    def img_fmt(self):
        return self["image_format"]


    def stat(self, **kwargs):
        return Files.stat(self, "dat", self.rng(), **kwargs)
    

    def report(self, **kwargs):
        stat = self.stat(**kwargs)
        
        print("Mean\t+-\tstd\n%1.4g\t+-\t%1.2g" % 
              (stat.getfloat("mean"), stat.getfloat("stdev")))

    
    def date(self, start_stop=False):
        date = \
        datetime.strptime(" ".join(self["date"].split()[:3]), "%Y %m %d")
        
        if start_stop:
            start = timedelta(seconds=self.getfloat("par", "start_time"))
            cent  = timedelta(seconds=self.getfloat("par", "center_time"))
            stop  = timedelta(seconds=self.getfloat("par", "end_time"))
            
            return Date(date + start, date + stop, date + cent)
        else:
            return date
    
    
    def datestr(self, fmt="%Y%m%d"):
        if any(elem in fmt for elem in ("%H", "%M", "%S")):
            date = self.date(start_stop=True)
            return date.center.strftime(fmt)
        else:
            return self.date().strftime(fmt)
    
        
    @classmethod
    def from_line(cls, line):
        if line.startswith("#") or not line.strip():
            pass
        
        split = line.split()
        return cls(datfile=split[0].strip(), parfile=split[1].strip(),
                   keep=True)
    
    
    @staticmethod
    def parse_split(split):
        s = split.strip()
        
        if s == "None":
            return None
        else:
            return s
    
    
    def avg_fact(self, fact=750):
        avg_rng = int(float(self.rng()) / fact)
    
        if avg_rng < 1:
            avg_rng = 1
    
        avg_azi = int(float(self.azi()) / fact)
    
        if avg_azi < 1:
            avg_azi = 1
    
        return avg_rng, avg_azi
        
    
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


class SLC(DataFile):
    def multi_look(self, MLI, **kwargs):
        args = parse_ml_args(**kwargs)
        gp.multi_look(self.datpar, MLI.datpar, args["rng_looks"],
                      args["azi_looks"], args["start"], args["nlines"],
                      args["scale"], args["exp"])
    
    
    def plot_cmd(self):
        return "SLC"

    
    def copy(self, other, conv=None, scale=None, rng_off=0,
             rng_num=0, azi_off=0, azi_num=0, swap=0, nheader=0):

        gp.SLC_copy(self, other, conv, scale, rng_off, rng_num,
                    azi_off, azi_num, swap, nheader)
    

class MLI(DataFile):
    def plot_cmd(self):
        return "pwr"
    
    def rdc_trans(self, dem_rdc, other, lookup):
        gp.rdc_trans(self.par, dem_rdc, other.par, lookup)
    


# ************************
# * Auxilliary functions *
# ************************

    
def interfero(date1, date2, master_date, output_dir=".", range_looks=4,
              azimuth_looks=1):
    
    dates = "{}_{}".format(date2, date1)
    
    coreg_path = pth.join(output_dir, "coreg_out")
    
    sbas_dir = pth.join(output_dir, "SMALL_BASELINES", dates)

    create_dir(sbas_dir)

    SLC2_par = pth.join(coreg_path, date2 + ".slc.par")
    RSLC1 = pth.join(coreg_path, date1 + ".rslc")

    RSLC2 = pth.join(coreg_path, date2 + ".rslc")
    RMLI1 = pth.join(coreg_path, date1 + ".rmli")

    RSLC1_par = RSLC1 + ".par"
    RSLC2_par = RSLC2 + ".par"

    symlink(RSLC1, pth.join(sbas_dir, date1 + ".rslc"))
    symlink(RSLC2, pth.join(sbas_dir, date2 + ".rslc"))

    symlink(RSLC1_par, pth.join(sbas_dir, date1 + ".rslc.par"))
    symlink(RSLC2_par, pth.join(sbas_dir, date2 + ".rslc.par"))

    off = search_off(date1, output_dir)
    
    hgt = pth.join(output_dir, "geo", master_date + "dem.rdc")
    
    sim_unw = pth.join(output_dir, "sim_unw")
    
    diff = pth.join(output_dir, "SMALL_BASELINES", dates, dates + ".diff")

    REF_SLC_par = pth.join(coreg_path, master_date + ".rslc.par")
    REF_MLI_par = pth.join(coreg_path, master_date + ".rmli.par")
    
    REF_MLI_width = get_rng(REF_MLI_par)

    gp.phase_sim_orb(RSLC1_par, SLC2_par, off, hgt, sim_unw, REF_SLC_par,
                     None, None, 1, 1)

    gp.SLC_diff_intf(RSLC1, RSLC2, RSLC1_par, RSLC2_par, off,
                     sim_unw, diff, range_looks, azimuth_looks,
                     1, 0, 0.2, 1, 1)

    avg_rng, avg_azi = avg_factor(REF_MLI_par)

    gp.rasmph_pwr24(diff, RMLI1, REF_MLI_width, 1, 1, 0, avg_rng, avg_azi,
                    1.0, 0.35, 1, diff + ".bmp")

    base = pth.join(output_dir, "SMALL_BASELINES", dates, dates + ".base")
    
    gp.base_orbit(RSLC1_par, RSLC2_par, base)

#phase_sim_orb $RSLC3_ID.rslc.par $SLC_par $off $hgt $q.sim_unw $REF_SLC
#SLC_diff_intf $RSLC3_ID.rslc $RSLC $RSLC3_ID.rslc.par $RSLC_par $off $q.sim_unw $q.diff $RLK $AZLK 1 0 0.2 1 1
#rasmph_pwr24 $q.diff RMLI3 $REF_MLI_width 1 1 0 1 1 1. .35 1 $q.diff.$ras


def parse_ml_args(**kwargs):
    return {
        "rng_looks"    : int(kwargs.get("rng_looks", 1)),
        "azi_looks"    : int(kwargs.get("azi_looks", 1)),
        "start"        : int(kwargs.get("start", 0)),
        "nlines"       : kwargs.get("nlines", None),
        "scale"        : float(kwargs.get("scale", 1.0)),
        "exp"          : float(kwargs.get("exp", 1.0))
    }


def parse_dis_args(datfile, **kwargs):
    datfile = str(datfile)
    cmd  = kwargs.get("mode", None)
    flip = bool(kwargs.get("flip", False))
    rng = kwargs.get("rng", None)
    azi = kwargs.get("azi", None)
    parfile = kwargs.get("parfile", None)
    img_fmt = kwargs.get("image_format", None)
    debug = bool(kwargs.get("debug", False))
    
    parts = pth.basename(datfile).split(".")
    
    try:
        ext = [ext for ext in parts if ext in extensions][0]
    except IndexError:
        raise ValueError("Unrecognized extension of file %s. Available "
                         "extensions: %s" % (datfile, pr.extensions))
    
    
    if ext in ("sbi", "sm", "diff", "cc"):
        parfile_ext = ".diff_par"
        diff_file = True
    else:
        parfile_ext = ".par"
        diff_file = False
    
    if parfile is None:
        parfile = "%s%s" % (datfile, parfile_ext)
    
    if not pth.isfile(parfile):
        noparfile = True
    else:
        noparfile = False

    
    if flip:
        flip = -1
    else:
        flip = 1
    
    rng_none, azi_none, image_none = rng is None, azi is None, img_fmt is None

    if rng_none and noparfile:
        raise ValueError('Either "rng" or "parfile" (%s either missing or '
                         'not valid path) has to be given' % parfile)

    if image_none and noparfile:
        raise ValueError('Either "image_format" or "parfile" (%s either '
                         'missing or not valid path) has to be given'
                         % parfile)

    if rng_none:
        if diff_file:
            rng = Files.get_par("interferogram_width", parfile)
        else:
            rng = Files.get_par("range_samples", parfile)

    if azi_none:
        if diff_file:
            azi = Files.get_par("interferogram_azimuth_lines", parfile)
        else:
            azi = Files.get_par("azimuth_lines", parfile)
    

    if image_none:
        if diff_file:
            if ext == "cc":
                img_fmt = "FLOAT"
            else:
                img_fmt = "FCOMPLEX"
        else:
            img_fmt = Files.get_par("image_format", parfile)
    
    
    if cmd is None:
        cmd = [cmd for cmd, exts in plot_cmd_files.items()
               if ext in exts][0]
    
    
    return {
        "cmd"      : cmd,
        "parfile"  : parfile,
        "rng"      : rng,
        "azi"      : azi,
        "img_fmt"  : DataFile.data_types[img_fmt],
        "start"    : kwargs.get("start", None),
        "nlines"   : kwargs.get("nlines", None),
        "scale"    : kwargs.get("scale", None),
        "exp"      : kwargs.get("exp", None),
        "LR"       : int(flip),
        "debug"    : debug
    }


def parse_ras_args(datfile, **kwargs):
    args = parse_dis_args(datfile, **kwargs)
    
    raster = kwargs.get("raster", None)
    azi = kwargs.get("azi", None)
    avg_fact = kwargs.get("avg_fact", 750)
    parfile = args["parfile"]
    
    if raster is None:
        raster = "%s.%s" % (datfile, settings["ras_ext"])

    if avg_fact == "noavg":
        avg_rng, avg_azi = None, None
    else:
        avg_rng, avg_azi = DataFile.avg_factor(args["rng"], args["azi"], avg_fact)

    
    args.update({
        "raster" : raster,
        "arng": avg_rng,
        "aazi": avg_azi,
        "hdrsz": int(kwargs.get("hdrsz", 0))
        })
    
    return args


def display(datfile, **kwargs):
    args = parse_dis_args(datfile, **kwargs)

    gp.__dict__["dis" + args["cmd"]]\
    (datfile, args["rng"], args["start"], args["nlines"], args["scale"],
     args["exp"], args["img_fmt"], debug=args["debug"])


def raster(datfile, **kwargs):
    args = parse_ras_args(datfile, **kwargs)
    
    ras, cmd = getattr(gp, "ras" + args["cmd"]), args["cmd"]
    
    if cmd == "SLC":
        ras\
        (datfile, args["rng"], args["start"], args["nlines"],
         args["arng"], args["aazi"], args["scale"], args["exp"], args["LR"],
         args["img_fmt"], args["hdrsz"], args["raster"],
         debug=args["debug"])
    else:
        sec = kwargs.pop("sec", None)
        
        if sec is None:
            ras\
            (datfile, args["rng"], args["start"], args["nlines"],
             args["arng"], args["aazi"], args["scale"], args["exp"], args["LR"],
             args["raster"], args["img_fmt"], args["hdrsz"],
             debug=args["debug"])
        else:
            ras\
            (datfile, sec, args["rng"], args["start"], args["nlines"],
             args["arng"], args["aazi"], args["scale"], args["exp"], args["LR"],
             args["raster"], args["img_fmt"], args["hdrsz"],
             debug=args["debug"])
            

def palette_line(line):
    return " ".join(str(float(elem) / 255.0) for elem in line.split())


def make_palette(cmap):
    ret = "defined (%s)"
    
    with open(cmap) as f:
        return ret % ",".join("%d %s" % (ii, palette_line(line))
                              for ii, line in enumerate(f))


def make_colorbar(inras, outras, cmap, title="", ratio=1, start=0.0,
                  stop=255.0):
    
    tmp = Files.get_tmp()
    cbar, script = tmp + ".png", tmp + ".prt"
    cmap = pth.join(gamma_cmaps, "%s" % cmap)
    
    palette = "set palette %s" % make_palette(cmap)
    
    
    with open(script, "w") as f:
        f.write(cbar_tpl.format(out=cbar, xmin=start, xmax=stop,
                                ratio=ratio, title=title, palette=palette))
    
    _gnuplot(script)
    _convert(inras, cbar, "+append", outras)
    
    rm(cbar, script)
    


# specialize function
# from functools import partial

def multi_run_wrapper(args):
   return add(*args)

def add(args):
    (x, y) = args
    return x + y

# execution
    from multiprocessing import Pool
    pool = Pool(4)
    
    with Pool(4) as pool:
        results = pool.map(add ,[(1,2),(2,3),(3,4)])
    
    print(results)


def parse_opt(key, value):
    if value is True:
        return "-%s" % (key)
    else:
        return "-%s %s" % (key, value)


def montage(out, *args, **kwargs):
    size = kwargs.pop("size")
    debug = bool(kwargs.get("debug", False))
    
    
    if size is not None:
        kwargs["resize"] = "x".join(str(pixel) if pixel is not None else ""
                                    for pixel in size)
    
    options = " ".join(parse_opt(key, value)
                       for key, value in kwargs.items())
    
    
    
    files = " ".join(str(arg) for arg in args)
    
    _montage(files, files, options, out, debug=debug)
    

    
class RGB(object):
    def __init__(self, r=0, g=0, b=0):
        self.rgb = (r, g, b)
    
    def __str__(self):
        return "%d, %d, %d" % (self.rgb[0], self.rgb[1], self.rgb[2])





colors = {
    "black"     : RGB(  0,    0,    0),
    "white"     : RGB(255,  255,  255),
    "red"       : RGB(255,    0,    0),
    "lime"      : RGB(  0,  255,    0),
    "blue"      : RGB(  0,    0,  255),
    "yellow"    : RGB(255,  255,    0),
    "aqua"      : RGB(  0,  255,  255),
    "magenta"   : RGB(255,    0,  255),
    "silver"    : RGB(192,  192,  192),
    "gray"      : RGB(128,  128,  128),
    "maroon"    : RGB(128,    0,    0),
    "olive"     : RGB(128,  128,    0),
    "olive"     : RGB(128,  128,    0),
    "green"     : RGB(  0,  128,    0),
    "purple"    : RGB(128,    0,  128),
    "teal"      : RGB(  0,  128,  128),
    "navy"      : RGB(  0,    0,  128)
}


def _proc_arg(arg):
    if arg is not None:
        return str(arg)
    else:
        return "-"

        
plot_cmd_files = {
    "pwr": ("pix_sigma0", "pix_gamma0", "sbi_pwr", "cc", "rmli", "mli"),
    "SLC": ("slc", "rslc"),
    "mph": ("sbi", "sm", "diff", "lookup", "lt"),
    "hgt": ("hgt", "rdc")
}


extensions = " ".join(" ".join(items) for items in plot_cmd_files.values())


cbar_tpl = \
"""\
set terminal pngcairo size 200,800
set output "{out}"
set pm3d map

g(x,y) = y

set yrange [{xmin}:{xmax}]
# set ytics 0.2
set ytics scale 1.5 nomirror
# set mytics 2
set size ratio 1e{ratio}

{palette}

unset colorbox; unset key; set tics out; unset xtics
set title "{title}"
splot g(x,y)
"""
