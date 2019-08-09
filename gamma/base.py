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
from json import JSONEncoder
from keyword import iskeyword
from collections import OrderedDict
from functools import partial
from re import match

import gamma as gm


from utils import *


__all__ = (
    "save",
    "Struct",
    "extend",
    "Extract",
    "extract",
    "Date",
    "Parfile",
    "DataFile",
    "SLC",
    "MLI",
    "imview",
    "str_t",
    "settings",
    "gp",
    "ScanSAR",
    "settings",
    "display",
    "raster",
    "make_cmd"
)


ScanSAR = True



versions = {
    "20181130": "/home/istvan/progs/GAMMA_SOFTWARE-20181130"
}


settings = {
    "ras_ext": "bmp",
    "path": versions["20181130"],
    "modules": {"DIFF", "DISP", "ISP", "LAT", "IPTA"},
    "libpaths": "/home/istvan/miniconda3/lib:",
    "templates": {
        "IW": "{date}_iw{iw}.{pol}.slc",
        "date": {
            "short": "%Y%m%d",
            "long": "%Y%m%dT%H%M%S"
        },
        "tab": "{date}.{pol}.SLC_tab"
    },
    "cache_default_path" : "/mnt/bozso_i/cache"
}



os.environ["LD_LIBRARY_PATH"] += settings["libpaths"]

log = getLogger("gamma.base")


gamma_cmaps = pth.join(settings["path"], "DISP", "cmaps")


gamma_commands = {
    binfile for module in settings["modules"]
    for path in {"bin", "scripts"}
    for binfile in iglob(pth.join(settings["path"], module, path, "*"))
}


def make_cmd(command):
    def cmd(*args, **kwargs):
        _log = kwargs.pop("log", None)
        
        Cmd = "%s %s" % (command, " ".join(_proc_arg(arg) for arg in args))
        
        
        if gm.is_debug():
            log.debug('Issued command is "%s"' % Cmd)
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


if 0:
    gamma_commands = {"rashgt", "ScanSAR_burst_corners"}


gp = type("Gamma", (object,), 
          {pth.basename(cmd): staticmethod(make_cmd(cmd))
           for cmd in gamma_commands})


imview = make_cmd("eog")


def save(obj):
    return obj.__save__

def extend(base, *field_names):
    def inner(cls):
        attribs = {"__slots__": base.__slots__ + field_names}
        attribs.update({key: val for key, val in cls.__dict__.items()
                        if key not in {"__dict__", "__weakref__"}})
        
        return type(cls.__name__, (base,), attribs)
    
    return inner

class PlainBase(object):
    def init(self, *args, **kwargs):
        for field in self.__slots__:
            setattr(self, field, kwargs.get(field))
    
    def __eq__(self, other):
        if not isinstance(other, self.__class__):
            return False
        return all(i == j for i, j in zip(self, other))

    def __iter__(self):
        for name in self.__slots__:
            yield getattr(self, name)

    def __repr__(self):
        txt = ", ".join("%s: %s" % (val, getattr(self, val))
                        for val in self.__slots__)
        return "%s(%s)"% (type(self).__name__, txt)

    def to_dict(self):
        return OrderedDict(zip(self.__slots__, self))


def Struct(*field_names):
    def inner(cls):
        attribs = {"__slots__": field_names}
        attribs.update({key: val for key, val in cls.__dict__.items()
                        if key not in {"__dict__", "__weakref__"}})
        
        return type(cls.__name__, (PlainBase,), attribs)
    
    return inner


def check_name(name):
    if not all(c.isalnum() or c == '_' for c in name):
        raise ValueError('Type names and field names can only contain alphanumeric characters and underscores: %r' % name)
    if iskeyword(name):
        raise ValueError('Type names and field names cannot be a keyword: %r' % name)
    if name[0].isdigit():
        raise ValueError('Type names and field names cannot start with a number: %r' % name)


Extract = new_type("Extract", "zipfile, files")


def extract(ext, outpath):
    extractor = partial(ext.zipfile.extract, path=outpath)
    
    return ext.files.map(extractor)


make_match = partial(partial, match)


def search_extract(zipfile, *tpl):
    namelist = zipfile.namelist()
    
    return gm.Extract(zipfile=zipfile,
                      files=Seq(*tpl).map(make_match)
                                     .map(lambda x: filter(x, namelist))
                                     .chain())


@Struct("start", "stop", "center")
class Date:
    def __init__(self, start_date, stop_date, center=None):
        self.start = start_date
        self.stop = stop_date
        
        if center is None:
            center = (start_date - stop_date) / 2.0
            center = stop_date + center
    
        self.center = center
    
    
    def date2str(self, fmt="%Y%m%d"):
        return self.center.strftime(fmt)
    
    # def __str__(self):
    #     return self.date2str()


class Parfile(object):
    __slots__ = ("par", "cache")
    __save__ = __slots__
    
    
    def __init__(self, **kwargs):
        self.cache = {}
        self.par = kwargs.get("parfile")
    
    def __str__(self):
        return self.par
    
    
    def __getitem__(self, item):
        if item in self.cache:
            value = self.cache[item]
        else:
            value = gm.get_par(item, self.par)
            self.cache[item] = value
            
        return value
    
    
    def __contains__(self, item):
        return self[item] is not None
    
    def __str__(self):
        return self.par
    
    def float(self, key, idx=0):
        return float(self[key].split()[idx])
    
    def int(self, key, idx=0):
        return int(self[key].split()[idx])
    
    
    def set_par(key, new):
        if gm.Files.is_empty(self.par) or key not in self:
            with open(self.path, "a") as f:
                f.write("%s: %s\n" % (key, new))
        
        elif key in self:
            with open(self.par, "r+") as f:
                lines = (line for line in f)
        
                lines = (
                            "%s: %s" % (key, new)
                            if key in line
                            else line
                            for line in lines
                        )
            
                f.seek(0)
                f.truncate()
            
                f.write("%s\n" % "\n".join(lines))


@Struct("dat", "par", "datpar", "tab", "files")
class DataFile:
    __save__ = {"dat", "par", "tab"}
    
    data_types = {
        "FCOMPLEX": 0,
        "SCOMPLEX": 1,
        "FLOAT": 0,
        "SHORT_INT": 1,
        "DOUBLE": 2
    }
    
    pols = frozenset({"vv", "hh", "hv", "vh"})

    
    def __init__(self, **kwargs):
        datfile   = kwargs.get("datfile")
        parfile   = kwargs.get("parfile")
        
        if datfile is None:
            datfile = tmp_file(**kwargs)
        
        if parfile is None:
            parfile = datfile + ".par"
        
        self.files = (datfile, parfile)
        self.datpar = " ".join(self.files)
        
        self.dat, self.par, self.tab, = \
        datfile, parfile, kwargs.get("tabfile", None)
    
    
    @classmethod
    def from_json(cls, line):
        return cls(datfile=line["dat"], parfile=line["par"],
                   tabfile=line["tab"])
    
    
    def rm(self):
        rm(*self.files)
    
    
    def save(self, datfile, parfile=None):
        if parfile is None:
            parfile = datfile + ".par"
        
        mv(self.dat, datfile)
        mv(self.par, parfile)
        
        self.dat, self.par = datfile, parfile

    
    def __str__(self):
        return self.datpar

    def __bool__(self):
        return  all(map(isfile, self.files))

    def rng(self):
        return self.int("range_samples")

    def azi(self):
        return self.int("azimuth_lines")
    
    def img_fmt(self):
        return self["image_format"]
    
    
    def pol(self):
        dat = self.dat
        for pol in self.pols:
            if pol in dat or pol.upper() in dat:
                return pol
    

    def stat(self, **kwargs):
        return Files.stat(self, "dat", self.rng(), **kwargs)
    

    def report(self, **kwargs):
        stat = self.stat(**kwargs)
        
        print("Mean\t+-\tstd\n%1.4g\t+-\t%1.2g" % 
              (stat.float("mean"), stat.float("stdev")))

    
    def date(self, start_stop=False):
        date = \
        datetime.strptime(" ".join(self["date"].split()[:3]), "%Y %m %d")
        
        if start_stop:
            start = timedelta(seconds=self.float("start_time"))
            cent  = timedelta(seconds=self.float("center_time"))
            stop  = timedelta(seconds=self.float("end_time"))
            
            return gm.Date(start=date + start, stop=date + stop,
                           center=date + cent)
        else:
            return date
    
    
    def datestr(self, fmt="%Y%m%d"):
        # TODO: use regexp for matching?
        if any(map(lambda x: x in fmt, {"%H", "%M", "%S"})):
            date = self.date(start_stop=True)
            return date.center.strftime(fmt)
        else:
            return self.date().strftime(fmt)
    
        
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
    plot_cmd = "SLC"
    
    def multi_look(self, **kwargs):
        mli = MLI(**kwargs)
        args = parse_ml_args(**kwargs)
        
        gp.multi_look(self.datpar, mli.datpar, args["rng_looks"],
                      args["azi_looks"], args["start"], args["nlines"],
                      args["scale"], args["exp"])
        
        return mli
        
    
    def copy(self, other, conv=None, scale=None, rng_off=0,
             rng_num=0, azi_off=0, azi_num=0, swap=0, nheader=0):

        gp.SLC_copy(self, other, conv, scale, rng_off, rng_num,
                    azi_off, azi_num, swap, nheader)
    

class MLI(DataFile):
    plot_cmd = "pwr"
    
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
