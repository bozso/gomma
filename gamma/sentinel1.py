from os import path as pth
from datetime import datetime, timedelta
from zipfile import ZipFile
import shutil as sh


import gamma as gm
from gamma.private import extract_file, burst_selection_helper

from logging import getLogger

log = getLogger("gamma.sentinel1")

gp = gm.gamma_progs


__all__ = ("S1Zip", "S1IW", "S1SLC", "deramp_master", "deramp_slave", "coreg")


def check_paths(path):
    if len(path) != 1:
        raise Exception("More than one or none file(s) found in the zip that "
                        "corresponds to the regexp. Paths: {}".format(path))
    else:
        return path[0]


class S1Zip(object):
    if gm.ScanSAR:
        cmd = "ScanSAR_burst_corners"
    else:
        cmd = "SLC_burst_corners"
    
    burst_fun = getattr(gp, cmd)

    r_tiff_tpl = ".*.SAFE/measurement/s1.*-iw{iw}-slc-{pol}.*.tiff"
    r_annot_tpl = ".*.SAFE/annotation/s1.*-iw{iw}-slc-{pol}.*.xml"
    r_calib_tpl = ".*.SAFE/annotation/calibration/calibration"\
                  "-s1.*-iw{iw}-slc-{pol}.*.xml"
    r_noise_tpl = ".*.SAFE/annotation/calibration/noise-s1.*-"\
                  "iw{iw}-slc-{pol}.*.xml"
    
    __slots__ = {"zipfile", "mission", "date", "burst_nums", "mode",
                 "prod_type", "resolution", "level", "prod_class", "pol",
                 "abs_orb", "DTID", "UID"}
    
    __save__ = {"zipfile", "burst_nums", "date", "mission"}
    
    
    def __init__(self, zipfile, extra_info=False):
        zip_base = pth.basename(zipfile)
        
        self.zipfile = zipfile
        self.mission = zip_base[:3]
        self.date = gm.Date(datetime.strptime(zip_base[17:32], "%Y%m%dT%H%M%S"),
                            datetime.strptime(zip_base[33:48], "%Y%m%dT%H%M%S"))
        
        self.burst_nums = None
        
        
        if extra_info:
            self.mode = zip_base[4:6]
            self.prod_type = zip_base[7:10]
            self.resolution = zip_base[10]
            self.level = int(zip_base[12])
            self.prod_class = zip_base[13]
            self.pol = zip_base[14:16]
            self.abs_orb = int(zip_base[49:55])
            self.DTID = zip_base[56:62]
            self.UID = zip_base[63:67]
    
    
    @classmethod
    def from_json(cls, line):
        ret = cls(line["zipfile"])
        ret.burst_nums = line["burst_nums"]
        
        return ret
    
    
    def __str__(self):
        line = "%s;" % self.zipfile
        
        if self.burst_nums is not None:
            line += ",".join(str(elem) for elem in self.burst_nums)
    
    
    def datestr(self, fmt="%Y%m%d"):
        return self.date.center.strftime(fmt)

    
    def extract_annot(self, iw, pol, out_path="."):
        regx = self.r_annot_tpl.format(iw=iw, pol=pol)
        
        with ZipFile(self.zipfile, "r") as slc_zip:
            ret = extract_file(slc_zip, regx, out_path)
    
        return ret
    
    
    def extract_IW(self, pol, IW, annot=None):
        iw_num = IW.num
        log.info("Extracting IW%d of %s" % (iw_num, pth.basename(self.zipfile)))
        
        r_tiff = self.r_tiff_tpl.format(iw=iw_num, pol=pol)
        r_calib = self.r_calib_tpl.format(iw=iw_num, pol=pol)
        r_noise = self.r_noise_tpl.format(iw=iw_num, pol=pol)
    
        with ZipFile(self.zipfile, "r") as slc_zip:
            tiff  = extract_file(slc_zip, r_tiff, ".")
            calib = extract_file(slc_zip, r_calib, ".")
            noise = extract_file(slc_zip, r_noise, ".")
    
            if annot is None:
                r_annot = self.r_annot_tpl.format(iw=iw_num, pol=pol)
                annot = extract_file(slc_zip, r_annot, ".")
        
        tiff, annot, calib, noise = check_paths(tiff), check_paths(annot), \
                                    check_paths(calib), check_paths(noise)
        
        gp.par_S1_SLC(tiff, annot, calib, noise, IW.par, IW.dat,
                      IW.TOPS_par)
        
        return IW
        
        
    def test_zip(self):
        with ZipFile(self.zipfile, "r") as slc_zip:
    
            testzip = slc_zip.testzip()
    
            if testzip:
                log.error("Bad zipfile detected. First bad file is "
                          "\"%s\" in zipfile \"%s\"."
                          % (testzip, zipfile))
                return False
        return True

    
    def burst_info(self, iw_num, pol, remove_temps=False):
        annot = self.extract_annot(iw_num, pol)[0]

        gp.par_S1_SLC(None, annot, None, None, "tmp_par", None, "tmp_TOPS_par")
        
        out = S1Zip.burst_fun("tmp_par", "tmp_TOPS_par").decode()
        
        
        if remove_temps:
            rm("tmp_par", "tmp_TOPS_par")
        
        return out

    
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


class S1IW(gm.DataFile):
    __slots__ = {"TOPS_par", "num"}
    
    tpl = gm.settings["templates"]["IW"]
    
    
    def __init__(self, num, TOPS_parfile=None, **kwargs):
        
        gm.DataFile.__init__(self, **kwargs)


        if TOPS_parfile is None:
            TOPS_parfile = self.dat + ".TOPS_par"
        
        self.TOPS_par, self.num = gm.Parfile(parfile=TOPS_parfile), num

    
    def save(self, datfile, parfile=None, TOPS_parfile=None):
        DataFile.save(self, datfile, parfile)
        
        self.mv("TOPS_par", TOPS_parfile)
        
        self.TOPS_par = gm.Parfile(TOPS_parfile)
    
    
    def rm(self):
        Files.rm(self, "dat", "par", "TOPS_par")
    

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
            else:
                return ret
        else:
            return ret
    
    
    # def date(self, start_stop=False):
    #     date = \
    #     datetime.strptime(" ".join(self["date"].split()[:3]), "%Y %m %d")
    #     
    #     if start_stop:
    #         start = timedelta(seconds=self.getfloat("start_time"))
    #         cent  = timedelta(seconds=self.getfloat("center_time"))
    #         stop  = timedelta(seconds=self.getfloat("end_time"))
    #         
    #         return gm.Date(date + start, date + stop, date + cent)
    #     else:
    #         return date
    

    @classmethod
    def from_tabline(cls, line):
        split = [elem.strip() for elem in line.split()]
        
        return cls(0, datfile=split[0], parfile=split[1],
                   TOPS_parfile=split[2], keep=True)
    
    
    @classmethod
    def from_template(cls, pol, date, num, tpl=None, **kwargs):
        if tpl is None:
            tpl = cls.tpl
        
        return cls(num, datfile=tpl.format(date=date, iw=num, pol=pol),
                   keep=True, **kwargs)


    def lines_offset(self):
        fl = (self.getfloat("burst_start_time_2")
              - self.getfloat("burst_start_time_1")) \
              / self.getfloat("azimuth_line_time")
        
        return Offset(fl, int(0.5 + fl))



class S1SLC(object):
    __slots__ = {"IWs", "tab", "slc"}
    __save__ = {"tab",}
    
    tab_tpl = gm.settings["templates"]["tab"]
    
    
    def __init__(self, IWs, tabfile):
        self.IWs, self.tab, self.slc = IWs, tabfile, None
        
        with open(tabfile, "w") as f:
            f.write("%s\n" % str(self))
    
    
    def __bool__(self):
        return all(bool(IW) for IW in self.IWs if IW is not None)

    
    def __str__(self):
        return "\n".join(str(IW) for IW in self.IWs if IW is not None)
    
    @classmethod
    def from_json(cls, line):
        return cls.from_tabfile(line["tab"])
    
    
    @classmethod
    def from_SLC(cls, other, extra):
        
        tabfile = other.tab + extra
        
        IWs = tuple(
                S1IW(ii, datfile=iw.dat + extra, keep=True)
                if iw is not None else None
                for ii, iw in enumerate(other.IWs)
        )
        
        return cls(IWs, tabfile)

    
    @classmethod
    def from_tabfile(cls, tabfile):
        
        with open(tabfile, "r") as f:
            IWs = tuple(S1IW.from_tabline(line) for line in f)
        
        return cls(IWs, tabfile)    
    
    
    @classmethod
    def from_template(cls, date, burst_num, pol, fmt="short", dirpath=".",
                      **kwargs):
        tpl_tab = pth.join(dirpath, cls.tab_tpl)
        
        if fmt is not None:
            date = date.date2str(gm.settings["templates"]["date"][fmt])
        
        
        tpl = pth.join(dirpath, S1IW.tpl)
        
        IWs = tuple(
                S1IW.from_template(pol, date, ii + 1, tpl=tpl, **kwargs)
                if iw is not None else None
                for ii, iw in enumerate(burst_num)
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
        self.slc = gm.SLC(**kwargs)
        
        gp.SLC_mosaic_S1_TOPS(self.tab, self.slc.datpar, rng_looks, azi_looks,
                              debug=debug)
        
        return self.slc
        

    def multi_look(self, MLI, rng_looks=1, azi_looks=1, wflg=0):
        gp.multi_S1_TOPS(self.tab, MLI.datpar, rng_looks, azi_looks, wflg)


def coreg(master, SLC, RSLC, hgt=0.1, rng_looks=10, azi_looks=2,
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
