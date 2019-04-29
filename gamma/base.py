import os
import os.path as pth
import shutil as sh

from sys import version_info
from shutil import copyfileobj
from glob import iglob
from math import sqrt, isclose
from argparse import ArgumentParser
from zipfile import ZipFile
from logging import getLogger
from datetime import datetime, timedelta
from errno import ENOENT, EEXIST
from collections import namedtuple
from itertools import tee
from tempfile import _get_default_tempdir, _get_candidate_names
from pprint import pformat
from subprocess import check_output, CalledProcessError, STDOUT
from shlex import split



PY3 = version_info[0] == 3

__all__ = ("DataFile", "SLC", "MLI", "gp", "imview", "gnuplot", "Temps",
           "Argp", "mkdir", "ln", "rm", "mv", "colors", "Files", "HGT",
           "make_colorbar", "Base", "IFG", "cat", "tmpdir", "string_t",
           "settings", "all_same", "gamma_progs", "ScanSAR", "montage",
           "get_tmp")


ScanSAR = True

if PY3:
    string_t = str,
else:
    string_t = basestring,


os.environ["LD_LIBRARY_PATH"] = \
os.getenv("LD_LIBRARY_PATH") + "/home/istvan/miniconda3/lib:"

tmpdir = _get_default_tempdir()

settings = {
    "ras_ext": "bmp",
    "path": "/home/istvan/progs/GAMMA_SOFTWARE-20181130",
    "modules": ("DIFF", "DISP", "ISP", "LAT", "IPTA")
}

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


gamma_progs = type("Gamma", (object,),
                   dict((pth.basename(cmd), staticmethod(make_cmd(cmd)))
                   for cmd in gamma_commands))


gp = gamma_progs


_convert = make_cmd("convert")
_montage = make_cmd("montage")
gnuplot = make_cmd("gnuplot")
imview   = make_cmd("eog")


def all_same(iterable):
    return len(set(tee(iterable,1))) == 1


def make_object(name, inherit=(object,), **kwargs):
    return type(name, inherit, **kwargs)



class Params(object):
    def __init__(self, dictionary):
        self.params = dictionary
    
    @classmethod
    def from_file(cls, path, sep=":"):
        with open(path, "r") as f:
            return cls({
                line.split(sep)[0].strip() : " ".join(line.split(sep)[1:]).strip()
                for line in f if line
            })
    
    def __str__(self):
        return pformat(self.params)
    
    def __getitem__(self, key):
        return self.params[key]

    def getfloat(self, key, idx=0):
        return float(self[key].split()[idx])

    def getint(self, key, idx=0):
        return int(self[key].split()[idx])


class Temps(object):
    def __init__(self, *args):
        self.tmp_paths = list(*args)

    
    def add(self, *args):
        self.tmp_paths.extend(args)

    
    def __del__(self):
        for path in self.tmp_paths:
            log.debug('Removed file: "%s"' % path)
            rm(path)


tmp = Temps()


def get_tmp(path=tmpdir):
    global tmp
    
    path = pth.join(path, next(_get_candidate_names()))
    
    tmp.add(path)
    
    return path



class Files(object):
    def __init__(self, **kwargs):
        for key, value in kwargs.items():
            setattr(self, key, value)
    
    def stat(self, attrib, rng=None, roff=0, loff=0, nr=None, nl=None):
        obj = getattr(self, attrib)
        
        if rng is None:
            rng = obj.rng()
        
        if isinstance(obj, string_t):
            gp.image_stat(obj, rng, roff, loff, nr, nl, "tmp")
        else:
            gp.image_stat(obj.dat, rng, roff, loff, nr, nl, "tmp")
        
        pars = Params.from_file("tmp")
        rm("tmp")
        
        return pars
    

    def exist(self, *attribs):
        return all(pth.isfile(getattr(self, attrib)) for attrib in attribs)
    
    
    def mv(self, attrib, dst):
        Files._mv(getattr(self, attrib), dst)
    
    
    def move(self, attribs, dst):
        for attrib in attribs:
            attr = getattr(self, attrib)
            Files._mv(attr, dst)
            
            newpath = pth.join(pth.abspath(dst), pth.basename(attr))
            
            setattr(self, attrib, newpath)
    
    
    def rm(self, *attribs):
        for attrib in attribs:
            rm(getattr(self, attrib))
    
    def ln(self, attrib, other):
        ln(getattr(self, attrib), other)
    
    def cp(self, attrib, other):
        sh.copy(getattr(self, attrib), other)
    
    def get(self, attrib, key):
        return Files.get_par(key, getattr(self, attrib))
    
    def getfloat(self, attrib, key, idx=0):
        return Files._getfloat(key, getattr(self, attrib), idx)

    def getint(self, attrib, key, idx=0):
        return Files._getint(key, getattr(self, attrib), idx)
    
    def set(self, attrib, key, **kwargs):
        return Files.set_par(key, getattr(self, attrib), **kwargs)
    
    
    def empty(self, attrib):
        return Files.is_empty(getattr(self, attrib))

    
    @staticmethod
    def _mv(src, dst):
        if pth.isfile(dst):
            dst_ = dst
        elif pth.isdir(dst):
            dst_ = pth.join(dst, pth.basename(src))
            
        rm(dst_)
        sh.move(src, dst_)
        
        log.debug('File "%s" moved to "%s".' % (src, dst_))


    @staticmethod
    def is_empty(path):
        return pth.getsize(path) == 0
    
    
    @staticmethod
    def get_par(key, data):
        value = None

        if pth.isfile(data):
            with open(data, "r") as f:
                for line in f:
                    if key in line:
                        value = line
                        break

        elif isinstance(data, bytes):
            for line in data.decode().split("\n"):
                if key in line:
                    value = line
                    break

        elif isinstance(data, str):
            for line in data.split("\n"):
                if key in line:
                    value = line
                    break
                
        else:
            for line in data:
                if key in line:
                    value = line
                    break
        
        if value is not None:
            return " ".join(value.split(":")[1:]).strip() 
        else:
            return None

        
    @staticmethod
    def _getfloat(key, data, idx=0):
        return float(Files.get_par(key, data).split()[idx])

    @staticmethod
    def _getint(key, data, idx=0):
        return int(Files.get_par(key, data).split()[idx])


    @staticmethod
    def set_par(key, infile, new=""):
        if Files.is_empty(infile):
            with open(infile, "w") as f:
                f.write("%s: %s\n" % (key, new))
            
            return
        
        
        with open(infile, "r+") as f:
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

    
def Multi(**kwargs):
    return type("Multi", (object,), kwargs)


class Base(Files):
    def __init__(self, base, **kwargs):
        self.keep = kwargs.pop("keep", True)
        
        for key, value in kwargs.items():
            setattr(self, key, "%s%s" % (base, value))
        
        self.base = base
    
    def rm(self):
        for elem in dir(self):
            Files.rm(self, elem)

    def __del__(self):
        if not self.keep:
            self.rm()


class Date(object):
    __slots__ = ("start", "stop", "center")
    
    def __init__(self, start_date, stop_date, center=None):
        self.start = start_date
        self.stop = stop_date
        
        if center is None:
            center = (start_date - stop_date) / 2.0
            center = stop_date + mean
        
        self.center = center
    
    
    def date2str(self, fmt="%Y%m%d"):
        return self.center.strftime(fmt)
    
    def __eq__(self, other):
        return (self.start == other.start and self.stop == other.stop and
                self.mean == other.mean)
    
    def __str__(self):
        return self.date2str()

    def __repr__(self):
        return "<Date start: %s stop: %s mean: %s>"\
                % (self.start, self.stop, self.mean)


class DataFile(Files):
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
        
        if keep is not None and not self.keep:
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
        date = datetime.strptime(self["date"], "%Y %m %d")
        
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
        
        
        flipe = -1 if flip else 1
        

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
            raster = "%s.%s" % (args["datfile"], ras_ext)
    
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
    def __init__(self, **kwargs):
        DataFile.__init__(self, **kwargs)
    
    
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
    def __init__(self, **kwargs):
        DataFile.__init__(self, **kwargs)

    def plot_cmd(self):
        return "pwr"
    
    def rdc_trans(self, dem_rdc, other, lookup):
        gp.rdc_trans(self.par, dem_rdc, other.par, lookup)
    


class DEM(DataFile):
    __slots__ = ("lookup", "lookup_old")
    
    _geo2rdc = {
        "dist": 0,
        "nearest_neigh": 1,
        "sqr_dist": 2,
        "const": 3,
        "gauss": 4
    }
    
    _rdc2geo = {
        "nearest_neigh": 0,
    }
    

    def __init__(self, datfile, parfile=None, lookup=None, lookup_old=None):
        self.dat = datfile
        
        if parfile is None:
            parfile = pth.splitext(datfile) + ".dem_par"
        
        self.par, self.lookup, self.lookup_old = parfile, lookup, lookup_old
    
    
    def geo2rdc(self, infile, outfile, width, nlines=0, interp="dist",
                dtype=0):

        _interp = DEM._geo2rdc[interp]

        gp.geocode(self.lookup, infile, self["width"], outfile, width, nlines,
                   _interp, dtype)


    # TODO: interpolation modes
    def rdc2geo(self, infile, outfile, width, nlines=0, interp=1, dtype=0,
                flip_in=False, flip_out=False, order=5, func=None):
        
        lr_in = -1 if flip_in else 1
        lr_out = -1 if flip_out else 1
        
        gp.geocode_back(infile, self["width"], self.lookup, outfile, width,
                        nlines, interp, dtype, lr_in, lr_out, order)
    

    def rng(self):
        return self.getint("par", "width")

    def azi(self):
        return self.getint("par", "lines")
    
    
    def raster(self, obj, **kwargs):
        if obj in ("lookup", "lookup_old"):
            kwargs.setdefault("image_format", "FCOMPLEX")
            kwargs.setdefault("datfile", getattr(self, obj))
            kwargs.setdefault("mode", "mph")
            super(DEM, self).raster(**kwargs)


class Geocode(Files):
    _items = ("sim_sar", "zenith", "orient", "inc", "pix", "psi", "ls_map",
              "diff_par", "offs", "offsets", "ccp", "coffs", "coffsets")

    def __init__(self, path, mli, sigpa0=None, gamma0=None, **kwargs):
        self.par = mli.par
        
        if sigpa0 is None:
            sigpa0 = pth.join(path, "sigpa0")

        if gamma0 is None:
            gamma0 = pth.join(path, "gamma0")
        
        self.sigma0, self.gamma0 = sigpa0, gamma0
        
        
        elems = (
            (item, pth.join(path, kwargs[item]))
            if item in kwargs else
            (item, None)
            for item in self._items
        )

        self.__dict__.update(dict(elems))
    
    
    def rng(self):
        return self.getint("par", "range_samples")

    def azi(self):
        return self.getint("par", "azimuth_lines")

    
    def raster(self, obj, **kwargs):
        kwargs.setdefault("mode", "pwr")

        raster(getattr(self, obj), **kwargs)


class HGT(DataFile):
    rashgt = getattr(gp, "rashgt")
    
    def __init__(self, hgt, mli):
        self.keep = None
        self.dat = hgt
        self.mli = mli
        self.par = mli.par
        
        self.keep = True
    
    
    def __str__(self):
        return self.dat
    
    def raster(self, start_hgt=None, start_pwr=None, m_per_cycle=None,
               **kwargs):
        args = DataFile.parse_ras_args(self, **kwargs)
        
        HGT.rashgt(args["datfile"], self.mli.dat, args["rng"],
                   start_hgt, start_pwr, args["nlines"], args["arng"],
                   args["aazi"], m_per_cycle, args["scale"], args["exp"],
                   args["LR"], args["raster"], args["debug"])
    

class IFG(DataFile):
    __slots__ = ("diff_par", "qual", "filt", "cc", "dt", "slc1", "slc2",
                 "sim_unw")
    
    _off_algorithm = {
        "int_cc": 1,
        "fringe_vis": 2
    }
    

    def __init__(self, datfile, parfile=None, sim_unw=None, diff_par=None,
                 quality=None):
        
        self.keep = True
        self.dat = datfile
        
        base = pth.splitext(datfile)[0]
        
        if parfile is None:
            parfile = "%s.off" % base

        if sim_unw is None:
            sim_unw = "%s.sim_unw" % base
        

        self.par, self.qual, self.diff_par, self.filt, self.cc, \
        self.slc1, self.slc2, self.sim_unw = parfile, quality, diff_par, \
        None, None, None, None, sim_unw


    def rm(self):
        Files.rm(self, "dat", "par", "sim_unw")
    
    def __str__(self):
        return "%s %s %s %s %s" % (self.dat, self.par, self.sim_unw,
                                   self.diff_par, self.qual)
    
    def __repr__(self):
        return "<IFG datfile: %s, parfile: %s, sim_unw: %s, diff_par: %s, "\
               "quality_file: %s>" % (self.dat, self.par, self.sim_unw,
                                      self.diff_par, self.qual)
    
    
    @classmethod
    def from_SLC(cls, slc1, slc2, base, algorithm="int_cc",
                 rng_looks=1, azi_looks=1, interact=False, hgt=None,
                 slc_ref=None):

        _int = 1 if interact else 0
        
        off = "%s.off" % base
        sim_unw = "%s.sim_unw" % base
        
        gp.create_offset(slc1.par, slc2.par, off,
                         IFG._off_algorithm[algorithm], rng_looks, azi_looks,
                         _int)
        
        slc_ref_par = None if slc_ref is None else slc_ref.par
        
        gp.phase_sim_orb(slc1.par, slc2.par, off, hgt, sim_unw, slc_ref_par,
                         None, None, 1)
        
        gp.SLC_diff_intf(slc1.dat, slc2.dat, slc1.par, slc2.par, off,
                         sim_unw, diff, rng_looks, azi_looks, 0, 0)
        
        ret = cls(diff, parfile=off)
        ret.slc1, ret.slc2, ret.dt = slc1, slc2, slc2.date.mean - slc1.date.mean
        
        return ret
    
    
    @classmethod
    def from_line(cls, line):
        split = line.split()
        
        datfile  = DataFile.parse_split(split[0])
        parfile  = DataFile.parse_split(split[1])
        sim_unw  = DataFile.parse_split(split[2])
        diff_par = DataFile.parse_split(split[3])
        qual     = DataFile.parse_split(split[4])
        
        return cls(datfile, parfile, sim_unw, diff_par, qual)

        
    def rng(self):
        return self.getint("par", "interferogram_width")

    def azi(self):
        return self.getint("par", "interferogram_azimuth_lines")
    
    def img_fmt(self):
        return "FCOMPLEX"
    

    def check_quality(self):
        
        qual = self.qual
        
        with open(qual, "r") as f:
            offs = sum(float(line.split()[1]) for line in f
                       if line.startswith("azimuth_pixel_offset"))
        
        log.info("Sum of azimuth offsets in %s is %f pixel."
                 % (qual, offs))
        
        if isclose(offs, 0.0):
            return True
        
        return False

    
    def adf(self, filt=None, cc=None, alpha=0.5, fftwin=32, ccwin=7,
            step=None, loff=0, nlines=0, wfrac=0.7):

        if step is None:
            step = fftwin / 8
        
        if filt is None:
            filt = self.datfile + ".filt"

        if cc is None:
            cc = self.datfile + ".cc"

        self.filt, self.cc = filt, cc
        
        rng = self["interferogram_width"]
        
        gp.adf(self.dat, self.filt, self.cc, rng, alpha, fftwin, ccwin,
               step, loff, nlines, wfrac)
    

    def raster(self, start_cpx=1, start_pwr=1, start_cc=1, cc_min=0.2,
               **kwargs):
        mli = kwargs.pop("mli")
        
        args = DataFile.parse_ras_args(self, **kwargs)
        
        if self.cc is None:
            gp.rasmph_pwr24(args["datfile"], mli.dat, args["rng"],
                            start_cpx, start_pwr, args["nlines"],
                            args["arng"], args["aazi"], args["scale"],
                            args["exp"], args["LR"], args["raster"])
        else:
            gp.rasmph_pwr24(args["datfile"], mli.dat, args["rng"],
                            start_cpx, start_pwr, args["nlines"],
                            args["arng"], args["aazi"], args["scale"],
                            args["exp"], args["LR"], args["raster"],
                            self.cc, start_cc, cc_min)

    
    def rascc(self):
        pass
    


# ************************
# * Auxilliary functions *
# ************************

def search_pair(slc1, SLCs, used_SLCs):

    for slc2 in SLCs:
        if  slc1.date.mean.date() == slc2.date.mean.date() \
        and slc1.date.mean != slc2.date.mean and slc2 not in used_SLCs:
            return slc2

    return None


def check_paths(path):
    if len(path) != 1:
        raise Exception("More than one or none file(s) found in the zip that "
                        "corresponds to the regexp. Paths: {}".format(path))
    else:
        return path[0]

    
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
        ext = [ext for ext in parts if ext in pr.extensions][0]
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
        cmd = [cmd for cmd, exts in pr.plot_cmd_files.items()
               if ext in exts][0]
    
    
    return {
        "cmd"      : cmd,
        "parfile"  : parfile,
        "rng"      : rng,
        "azi"      : azi,
        "img_fmt"  : pr.data_types[img_fmt],
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
        raster = "%s.%s" % (datfile, _default_image_fmt)

    if avg_fact == "noavg":
        avg_rng, avg_azi = None, None
    else:
        avg_rng, avg_azi = pr.avg_factor(args["rng"], args["azi"], avg_fact)

    
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
    

def mkdir(path):
    try:
        os.makedirs(path)
        log.debug("Directory \"{}\" created.".format(path))
        return path
    except OSError as e:
        if e.errno != EEXIST:
            raise e
        else:
            log.debug("Directory \"{}\" already exists.".format(path))
            return path


def ln(target, link_name):
    try:
        os.symlink(target, link_name)
    except OSError as e:
        if e.errno == EEXIST:
            os.remove(link_name)
            os.symlink(target, link_name)
            log.debug("Symlink from \"%s\" to \"%s\" created"
                         % (target, link_name))
        else:
            raise e


def rm(*args):
    for arg in args:
        for path in iglob(arg):
            if not pth.isfile(path) and not pth.isdir(path):
                return
            elif pth.isdir(path):
                sh.rmtree(path)
                log.debug("Directory \"%s\" deleted," % path)
            elif pth.isfile(path):
                try:
                    os.remove(path)
                    log.debug("File \"%s\" deleted." % path)
                except OSError as e:
                    if e.errno != ENOENT:
                        raise e
            else:
                raise Exception("%s is not a file nor is a directory!" % path)


def mv(*args, **kwargs):
    dst = kwargs.pop("dst", None)
    for arg in args:
        for src in iglob(arg): 
            rm(pth.join(dst, src))
            sh.move(src, dst)
            log.debug("File \"%s\" moved to \"%s\"." % (src, dst))    


class Argp(ArgumentParser):
    def __init__(self, **kwargs):
        subcmd = bool(kwargs.pop("subcmd", False))
        
        ArgumentParser.__init__(self, **kwargs)
        
        if subcmd:
            self.subparser = self.add_subparsers(**kwargs)
        else:
            self.subparser = None
        
    
    def addargs(self, *args):
        for arg in args:
            Argp.ap_add_arg(self, arg)

    
    def subp(self, **kwargs):
        if self.subparser is None:
            self.subparser = self.add_subparsers(**kwargs)
        else:
            raise ValueError("One subparser is already initiated!")
    
    
    def subcmd(self, name, fun, *args, **kwargs):
        
        subtmp = self.subparser.add_parser(name, **kwargs)
        
        for arg in args:
            Argp.ap_add_arg(subtmp,  arg)
        
        subtmp.set_defaults(fun=fun)

    
    @staticmethod
    def narg(name, help=None, kind="opt", alt=None, type=str, choices=None,
             default=None, nargs=None):
        return (name, help, default, kind, alt, type, choices, nargs)
    
    
    @staticmethod
    def ap_add_arg(obj, arg):
        if arg[3] == "opt":
            if arg[4] is not None:
                obj.add_argument(
                    "--{}".format(arg[0]), "-{}".format(arg[4]),
                    help=arg[1],
                    type=arg[5],
                    default=arg[2],
                    nargs=arg[7],
                    choices=arg[6])
            else:
                obj.add_argument(
                    "--{}".format(arg[0]),
                    help=arg[1],
                    type=arg[5],
                    default=arg[2],
                    nargs=arg[7],
                    choices=arg[6])
        elif arg[3] == "pos":
            obj.add_argument(
                arg[0],
                help=arg[1],
                type=arg[5],
                choices=arg[6])
        
        elif arg[3] == "flag":
            obj.add_argument(
                "--{}".format(arg[0]),
                help=arg[1],
                action="store_true")


display_parser = Argp(add_help=False)

display_parser.addargs(
    Argp.narg("datfile", kind="pos", alt="d", help="Datafile."),
    Argp.narg("mode", alt="m", help="Command to use."),
    Argp.narg("flip", alt="f", kind="flag", help="Flip image left-right."),
    Argp.narg("rng", alt="r", help="Range samples."),
    Argp.narg("parfile", alt="p", help="Parameter file."),
    Argp.narg("image_format", alt="i", help="Image format."),
    Argp.narg("debug", alt="d", kind="flag", help="Debug mode.")
)


raster_parser = Argp(add_help=False, parents=[display_parser])

raster_parser.addargs(
    Argp.narg("raster", alt="R", help="Output raster file."),
    Argp.narg("azi", alt="a", help="Azimuth lines."),
    Argp.narg("avg_fact", alt="v", type=int, help="Pixel averaging factor",
              default=750)
)


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



def cat(out, *args):
    assert len(args) >= 1, "Minimum one input file is required"
    
    with open(out, 'wb') as f_out, open(args[0], 'rb') as f_in:
        copyfileobj(f_in, f_out)

    for arg in args[1:]:
        with open(out, 'ab') as f_out, open(arg, 'rb') as f_in:
            copyfileobj(f_in, f_out)
        


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
