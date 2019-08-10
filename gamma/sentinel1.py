import shutil as sh

from os import path as pth
from datetime import datetime, timedelta
from zipfile import ZipFile
from logging import getLogger
from re import match
from functools import partial, reduce
import operator as op

import gamma as gm

from utils import *

log = getLogger("gamma.sentinel1")

gp = gm.gp


__all__ = (
    "S1Zip",
    "S1IW",
    "S1SLC",
    "deramp_master",
    "deramp_slave",
    "S1_coreg"
)



class S1Zip(object):
    if hasattr(gm, "ScanSAR_burst_corners"):
        cmd = "ScanSAR_burst_corners"
    else:
        # fallback
        cmd = "SLC_burst_corners"
    
    burst_fun = getattr(gp, cmd)
    
    extract_regex = {
        "file": '{mission}-iw{iw}-slc-{pol}-.*-{obj.abs_orb}-'
                '{obj.DTID}-.*',
        "tiff":  "measurement/{tpl}.tiff",
        "annot": "annotation/{tpl}.xml",
        "calib": "annotation/{tpl}.xml",
        "noise": "annotation/calibration/noise-{tpl}.xml"
    }
    
    
    __slots__ = ("path", "mission", "datestr", "date", "burst_nums", "mode",
                 "prod_type", "resolution", "level", "prod_class", "pol",
                 "abs_orb", "DTID", "UID", "zip_base", "safe_join")
    
    
    __save__ = ("path", "burst_nums", "mission")
    
    
    def __init__(self, zippath, extra_info=False):
        zip_base = pth.basename(zippath)
        
        self.path, self.zip_base = zippath, zip_base
        
        self.mission = zip_base[:3]
        self.datestr = zip_base[17:48]
        self.date = gm.Date(datetime.strptime(zip_base[17:32], "%Y%m%dT%H%M%S"),
                            datetime.strptime(zip_base[33:48], "%Y%m%dT%H%M%S"))
        
        self.mode = zip_base[4:6]
        self.prod_type = zip_base[7:10]
        self.resolution = zip_base[10]
        self.level = int(zip_base[12])
        self.prod_class = zip_base[13]
        self.pol = zip_base[14:16]
        self.abs_orb = zip_base[49:55]
        self.DTID = zip_base[56:62]
        self.UID = zip_base[63:67]
        self.safe_join = partial(pth.join,
                                 zip_base.replace(".zip", ".SAFE"))
    
    
    @classmethod
    def from_json(cls, line):
        ret = cls(line["zip_path"])
        ret.burst_nums = line["burst_nums"]
        
        return ret
    
    
    def __str__(self):
        line = "%s;" % self.path
        
        return line
    
    
    def date2str(self, fmt="%Y%m%d"):
        return self.date.center.strftime(fmt)

    
    def burst_info(self, extracted, **kwargs):
        annots = (self.extract_templates(("annot",), **kwargs)
                      .map(gm.matcher, match_list=extracted)
                      .chain())
        
        print(annots.collect())
        
        exit()
        
        par, TOPS_par = tmp_file(), tmp_file()

        gp.par_S1_SLC(None, annot, None, None, par, None, TOPS_par)
        
        return S1Zip.burst_fun(par, TOPS_par).decode()
    
    
    def extract_templates(self, names, pol="vv", iw=".*"):
        tpl = self.extract_regex["file"]\
              .format(obj=self, mission=self.mission.lower(), pol=pol, iw=iw)
        
        return (Seq(self.extract_regex.get(name) for name in names)
                    .map(partial(str.format, tpl=tpl))
                    .map(self.safe_join))
    
    
    def test_zip(self):
        with ZipFile(self.zipfile, "r") as slc_zip:
    
            testzip = slc_zip.testzip()
    
            if testzip:
                log.error('Bad zipfile detected. First bad file is '
                          '"%s" in zipfile "%s".'
                          % (testzip, zipfile))
                return False
        return True

    
    
    def burst_corners(self, iw_num, pol, remove_temps=False):
        return tuple(float(elem) for elem in line.split()[:8] for line in
                     self.burst_info(iw_num, pol, remove_temps).split("\n")
                     if line.startswith("Burst:"))
        
    
    def burst_num(self, iw_num, pol, remove_temps=False):
        return tuple(float(line.split()[-1]) for line in
                     self.burst_info(iw_num, pol, remove_temps).split("\n")
                     if line.startswith("Burst:"))
    

    def get_burst_nums(self, pol, remove_tmps=False):
        return tuple(self.burst_num(ii, pol, remove_tmps) for ii in range(1,4))
    
    
    def select_bursts(self, pol, ref_burst_nums):
        log.info("Selecting bursts of %s." % self.zipfile)
        
        self.burst_nums = \
        tuple(burst_selection_helper(ref, slc) for ref, slc in
              zip(ref_burst_nums, self.get_burst_nums(pol)))
        
        return self.burst_nums




@gm.extend(gm.DataFile, "TOPS_par")
class S1IW:
    tpl = gm.settings["templates"]["IW"]
    
    def __init__(self, num, TOPS_parfile=None, **kwargs):
        
        gm.DataFile.__init__(self, **kwargs)


        if TOPS_parfile is None:
            TOPS_parfile = self.dat + ".TOPS_par"
        
        self.TOPS_par, self.num = gm.Parfile(parfile=TOPS_parfile), num

    
    def save(self, datfile, parfile=None, TOPS_parfile=None):
        DataFile.save(self, datfile, parfile)
        
        mv(self.TOPS_par, TOPS_parfile)
        
        self.TOPS_par = gm.Parfile(TOPS_parfile)
    
    
    def rm(self):
        rm(self, "dat", "par", "TOPS_par")
    

    def __bool__(self):
        return Files.exist(self, "dat", "par", "TOPS_par")


    def __str__(self):
        return "%s %s %s" % (self.dat, self.par, self.TOPS_par.par)


    def __getitem__(self, key):
        ret = gm.Parfile.__getitem__(self, key)
        
        if ret is None:
            ret = self.TOPS_par[key]
            
            if ret is None:
                raise ValueError('Keyword "%s" not found in parameter files.'
                                 % key)
        
        return ret
    
    
    @classmethod
    def from_tabline(cls, line):
        split = [elem.strip() for elem in line.split()]
        
        return cls(0, datfile=split[0], parfile=split[1],
                   TOPS_parfile=split[2])
    
    
    @classmethod
    def from_template(cls, pol, date, num, tpl=None, **kwargs):
        if tpl is None:
            tpl = cls.tpl
        
        return cls(num, datfile=tpl.format(date=date, iw=num, pol=pol),
                   **kwargs)


    def lines_offset(self):
        fl = (self.float("burst_start_time_2")
              - self.float("burst_start_time_1")) \
              / self.float("azimuth_line_time")
        
        return Offset(fl, int(0.5 + fl))


not_none = partial(filter, lambda x: x is not None)


@gm.Struct("IWs", "tab", "slc")
class S1SLC:
    __save__ = {"tab",}
    
    
    tab_tpl = gm.settings["templates"]["tab"]
    
    
    def __init__(self, IWs, tabfile):
        self.IWs, self.tab, self.slc = IWs, tabfile, None
        
        with open(tabfile, "w") as f:
            f.write("%s\n" % str(self))
    
    
    def __bool__(self):
        return all(map(bool, not_none(self.IWs)))

    
    def __str__(self):
        return "\n".join(map(str, not_none(self.IWs)))
    
    
    @classmethod
    def from_json(cls, line):
        return cls.from_tabfile(line["tab"])
    
    
    @classmethod
    def from_SLC(cls, other, extra):
        
        tabfile = other.tab + extra
        
        
        
        IWs = tuple(
                S1IW(ii, datfile=iw.dat + extra)
                for ii, iw in enumerate(not_none(other.IWs))
        )
        
        return cls(IWs, tabfile)

    
    @classmethod
    def from_tabfile(cls, tabfile):
        
        with open(tabfile, "r") as f:
            IWs = tuple(map(S1IW.from_tabline, f))
        
        return cls(IWs, tabfile)    
    
    
    @classmethod
    def from_template(cls, date, burst_num, pol, fmt="short", dirpath=".",
                      ext=None, **kwargs):
        tpl_tab = pth.join(dirpath, cls.tab_tpl)
        
        if fmt is not None:
            date = date.date2str(gm.settings["templates"]["date"][fmt])
        
        
        tpl = pth.join(dirpath, S1IW.tpl)
        
        if ext is not None:
            tpl = "%s.%s" % (tpl, ext)
            tpl_tab = "%s.%s" % (tpl_tab, ext)
        
        
        IWs = tuple(
                S1IW.from_template(pol, date, ii + 1, tpl=tpl, **kwargs)
                for ii, iw in enumerate(not_none(burst_num))
        )
        
        return cls(IWs, tpl_tab.format(date=date, pol=pol))
    
    
    def date(self, *args, **kwargs):
        return self.IWs[0].date(*args, **kwargs)
    
    
    def datestr(self, *args, **kwargs):
        return self.IWs[0].datestr(*args, **kwargs)
    
    
    def pol(self):
        return self.IWs[0].pol()
    
    
    def rm(self):
        for IW in self.IWs:
            IW.rm()
    
    
    def make_other(self, fmt="short", **kwargs):
        date = self.date(start_stop=True)
        burst_num = self.IWs
        pol = self.pol()
        
        return S1SLC.from_template(date, burst_num, pol, fmt=fmt, **kwargs)
        
    
    def num_IWs(self):
        return sum(1 for iw in self.IWs if iw is not None)
    
    
    def cp(self, other):
        for iw1, iw2 in zip(self.IWs, other.IWs):
            if iw1 is not None and iw2 is not None:
                sh.copy(iw1.dat, iw2.dat) 
                sh.copy(iw1.par, iw2.par) 
                sh.copy(str(iw1.TOPS_par), str(iw2.TOPS_par)) 
        
    
    def mosaic(self, rng_looks=1, azi_looks=1, debug=False, **kwargs):
        slc = gm.SLC(**kwargs)
        
        gp.SLC_mosaic_S1_TOPS(self.tab, slc.datpar, rng_looks, azi_looks,
                              debug=debug)
        
        return slc
        

    def multi_look(self, rng_looks=1, azi_looks=1, wflg=0, **kwargs):
        mli = gm.MLI(**kwargs)
        
        gp.multi_S1_TOPS(self.tab, mli.datpar, rng_looks, azi_looks, wflg)
        
        return mli

    
    def deramp(self, **kwargs):
        kwargs.setdefault("ext", "deramp")

        master, rng_looks, azi_looks, cleaning = \
        kwargs.get("master"), kwargs.get("rng_looks", 10), \
        kwargs.get("azi_looks", 2), kwargs.get("cleaning", False)
        
        
        if master is True:
            gp.S1_deramp_TOPS_reference(self.tab)
            
            return self.make_other(**kwargs)
        
        elif isinstance(master, S1SLC):
            cleaning = 1 if cleaning else 0
            
            gp.S1_deramp_TOPS_slave(self.tab, self.datestr(), master.tab,
                                    rng_looks, azi_look, cleaning)
            
            return self.make_other(**kargs)
        
        else:
            raise ValueError('"master" should either be a boolean or the '
                             'master S1SLC object!')
    
    
def S1_coreg(master, SLC, RSLC, hgt=0.1, rng_looks=10, azi_looks=2,
             poly1=None, poly2=None, cc_thresh=0.8, frac_thresh=0.01,
             ph_std_thresh=0.8, clean=True, use_inter=False, RSLC3=None,
             diff_dir="."):
    
    mslc = master["S1SLC"]
    
    cleaning = 1 if clean else 0
    flag1 = 1 if use_inter else 0
    
    SLC1_tab, SLC1_ID = mslc.tab, mslc.datestr()
    SLC2_tab, SLC2_ID = SLC.tab, SLC.datestr()
    
    if 1:
        if RSLC3 is None:
            log.info("Coregistering: %s." % SLC2_tab)
            out = gp.S1_coreg_TOPS(SLC1_tab, SLC1_ID, SLC2_tab, SLC2_ID,
                                   RSLC.tab, hgt, rng_looks, azi_looks,
                                   poly1, poly2, cc_thresh, frac_thresh,
                                   ph_std_thresh, cleaning, flag1)
        else:
            RSLC3_tab, RSLC3_ID = RSLC3.tab, RSLC3.date.date2str()
            log.info("Coregistering: %s. Reference: %s" % (SLC2_tab, RSLC3_tab))

            out = gp.S1_coreg_TOPS(SLC1_tab, SLC1_ID, SLC2_tab, SLC2_ID,
                                   RSLC.tab, hgt, rng_looks, azi_looks,
                                   poly1, poly2, cc_thresh, frac_thresh,
                                   ph_std_thresh, cleaning, flag1,
                                   RSLC3_tab, RSLC3_ID)
        
    
    ID = "%s_%s" % (SLC1_ID, SLC2_ID)
    
    ifg = gm.IFG(ID + ".diff", parfile=ID + ".off", diff_par=ID + ".diff_par",
                 quality=ID + ".coreg_quality")
    
    with open("coreg.output", "wb") as f:
        f.write(out)

    if ifg.check_quality():
        raise RuntimeError("Coregistration of %s failed!" % SLC2_ID)
    
    ifg.move(("dat", "par", "diff_par", "qual"), diff_dir)
    ifg.raster(mli=master["MLI"])


def deramp_master(mslc, slcd, rng_looks=4, azi_looks=1):
    mslcd = gb.S1SLC.from_SLC(mslc, ".deramp")
    
    gp.S1_deramp_TOPS_reference(mslc.tab)
    mslcd.mosaic(datfile=slcd, rng_looks=rng_looks, azi_looks=azi_looks)
    
    # RSLC = self.meta["RSLC"]
    RSLC = self.meta["RLSC"]
    
    deramped = tuple(gb.S1SLC.from_SLC(slc, ".deramp") for slc in RSLC)
    
    for rslc, dslc in zip(RSLC, deramped):
        date = rslc.date.date2str()
        
        gp.S1_deramp_TOPS_slave(rslc.tab, date, mslc.tab, rng_looks,
                                azi_looks, 0)
        
        _slc = pth.join(deramp_dir, "%s.slc.deramp" % date)
        
        dslc.mosaic(datfile=_slc, rng_looks=rng_looks, azi_looks=azi_looks)
    
    return mslcd


def deramp_slave(mslc, rslc, rslcd, rng_looks=4, azi_looks=1):
    date = rslc.date.date2str()

    deramped = gb.S1SLC.from_SLC(rslc, ".deramp")
    
    gp.S1_deramp_TOPS_slave(rslc.tab, date, mslc.tab, rng_looks,
                            azi_looks, 0)
    
    deramped.mosaic(datfile=rslcd, rng_looks=rng_looks, azi_looks=azi_looks)
    
    return deramped


def diff_burst(burst1, burst2):
    
    diff_sqrt = sqrt((burst1 - burst2) * (burst1 - burst2))
    
    return int(burst1 - burst2 + 1.0
               + ((burst1 - burst2) / (0.001 + diff_sqrt)) * 0.5)


def burst_selection_helper(ref_burst, slc_burst):
    if ref_burst is not None:
        iw_start_burst = slc_burst[0]
    
        diff = [diff_burst(ref_burst[0], iw_start_burst),
                diff_burst(ref_burst[-1], iw_start_burst)]
        
        total_slc_bursts = len(slc_burst)
    
        if diff[1] < 1 or diff[0] > total_slc_bursts:
            return None
    
        if diff[0] <= 0:
            diff[0] = 1
    
        if diff[1] > total_slc_bursts:
            diff[1] = total_slc_bursts
    
        return tuple((diff[0], diff[1]))
    else:
        return None


def check_paths(path):
    if len(path) != 1:
        raise Exception("More than one or none file(s) found in the zip that "
                        "corresponds to the regexp. Paths: {}".format(path))
    else:
        return path[0]
