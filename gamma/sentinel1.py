import gamma as gm

from logging import getLogger

log = getLogger("gamma.sentinel1")

gp = gm.gamma_progs


__all__ = ("S1Zip", "S1IW", "S1SLC", "deramp_master", "deramp_slave", "coreg")


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
    
    __slots__ = ("zipfile", "mission", "date", "burst_nums", "mode",
                 "prod_type", "resolution", "level", "prod_class", "pol",
                 "abs_orb", "DTID", "UID")
    
    def __init__(self, zipfile, extra_info=False):
        zip_base = pth.basename(zipfile)
        
        self.zipfile = zipfile
        self.mission = zip_base[:3]
        self.date = Date(datetime.strptime(zip_base[17:32], "%Y%m%dT%H%M%S"),
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


    def extract_annot(self, iw, pol, out_path="."):
        regx = self.r_annot_tpl.format(iw=iw, pol=pol)
        
        with ZipFile(self.zipfile, "r") as slc_zip:
            ret = pr.extract_file(slc_zip, regx, out_path)
    
        return ret
    
    
    def extract_IW(self, pol, IW, annot=None):
        iw_num = IW.num
        log.info("Extracting IW%d of %s" % (iw_num, pth.basename(self.zipfile)))
        
        r_tiff = self.r_tiff_tpl.format(iw=iw_num, pol=pol)
        r_calib = self.r_calib_tpl.format(iw=iw_num, pol=pol)
        r_noise = self.r_noise_tpl.format(iw=iw_num, pol=pol)
    
        with ZipFile(self.zipfile, "r") as slc_zip:
            tiff  = pr.extract_file(slc_zip, r_tiff, ".")
            calib = pr.extract_file(slc_zip, r_calib, ".")
            noise = pr.extract_file(slc_zip, r_noise, ".")
    
            if annot is None:
                r_annot = self.r_annot_tpl.format(iw=iw_num, pol=pol)
                annot = pr.extract_file(slc_zip, r_annot, ".")
        
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
        
        return \
        tuple(pr.burst_selection_helper(ref, slc) for ref, slc in
              zip(ref_burst_nums, self.get_burst_nums(pol)))


class S1IW(gm.DataFile):
    __slots__ = ("TOPS_par", "num")
    
    def __init__(self, num, TOPS_parfile=None, **kwargs):
        
        DataFile.__init__(self, **kwargs)


        if TOPS_parfile is None:
            TOPS_parfile = self.dat + ".TOPS_par"
        
        self.TOPS_par, self.num = TOPS_parfile, num

    
    def save(self, datfile, parfile=None, TOPS_parfile=None):
        DataFile.save(self, datfile, parfile)
        
        self.mv("TOPS_par", TOPS_parfile)
        
        self.TOPS_par = TOPS_parfile
    
        
    def rm(self):
        Files.rm(self, "dat", "par", "TOPS_par")
    

    def __bool__(self):
        return self.exist("dat", "par", "TOPS_par")


    def __str__(self):
        return "%s %s %s" % (self.dat, self.par, self.TOPS_par)


    def __getitem__(self, key):
        ret = self.get("par", key)
        
        if ret is None:
            ret = self.get("TOPS_par", key)
            
            if ret is None:
                raise ValueError('Keyword "%s" not found in parameter files.'
                                 % key)
            else:
                return ret
        else:
            return ret
    
    
    def getfloat(self, key, idx=0):
        return float(self[key].split()[idx])

    def getint(self, key, idx=0):
        return int(self[key].split()[idx])

    
    @classmethod
    def from_tabline(cls, line):
        split = [elem.strip() for elem in line.split()]
        
        return cls(0, datfile=split[0], parfile=split[1],
                   TOPS_parfile=split[2], keep=True)
    
    
    @classmethod
    def from_template(cls, pol, date, num, **kwargs):
        tpl = kwargs.get("tpl", "{date}_iw{iw}.{pol}.slc")
        fmt = kwargs.get("fmt", "%Y%m%dT%H%M%S")
        
        return cls(num,
                   datfile=tpl.format(date=date.strftime(fmt), iw=num, pol=pol),
                   keep=True)


    def lines_offset(self):
        fl = (self.getfloat("burst_start_time_2")
              - self.getfloat("burst_start_time_1")) \
              / self.getfloat("azimuth_line_time")
        
        return Offset(fl, int(0.5 + fl))



class S1SLC(object):
    __slots__ = ("IWs", "tab", "date", "slc")
    
    def __init__(self, IWs, tabfile, date=None):
        self.IWs, self.tab, self.date, self.slc = IWs, tabfile, date, None
        
        with open(tabfile, "w") as f:
            f.write("%s\n" % str(self))


    def rm(self):
        for IW in self.IWs:
            IW.rm()
    
    
    @classmethod
    def from_SLC(cls, other, extra):
        
        tabfile = other.tab + extra
        
        IWs = tuple(
                S1IW(ii, datfile=iw.dat + extra, keep=True)
                if iw is not None else None
                for ii, iw in enumerate(other.IWs)
        )
        
        return cls(IWs, tabfile, other.date)
    
    
    @classmethod
    def from_tabfile(cls, tabfile, date=None):
        
        with open(tabfile, "r") as f:
            IWs = tuple(S1IW.from_tabline(line) for line in f)
        
        return cls(IWs, tabfile, date)    
    
    
    @classmethod
    def from_template(cls, pol, date, burst_num, **kwargs):
        tpl_tab = kwargs.get("tpl_tab", "{date}.{pol}.SLC_tab")
        fmt = kwargs.pop("fmt", "%Y%m%dT%H%M%S")
        
        IWs = tuple(
                S1IW.from_template(pol, date.mean, ii + 1, fmt=fmt, **kwargs)
                if iw is not None else None
                for ii, iw in enumerate(burst_num)
        )
        
        return cls(IWs, tpl_tab.format(date=date.date2str(fmt), pol=pol),
                   date=date)


    def num_IWs(self):
        return sum(1 for iw in self.IWs if iw is not None)
    
    
    def cp(self, other):
        for iw1, iw2 in zip(self.IWs, other.IWs):
            if iw1 is not None and iw2 is not None:
                sh.copy(iw1.dat, iw2.dat) 
                sh.copy(iw1.par, iw2.par) 
                sh.copy(iw1.TOPS_par, iw2.TOPS_par) 
        
    
    def mosaic(self, rng_looks=1, azi_looks=1, debug=False, **kwargs):
        slc = SLC(**kwargs)
        
        gp.SLC_mosaic_S1_TOPS(self.tab, slc.datpar, rng_looks, azi_looks,
                              debug=debug)
        
        return slc
        

    def multi_look(self, MLI, rng_looks=1, azi_looks=1, wflg=0):
        gp.multi_S1_TOPS(self.tab, MLI.datpar, rng_looks, azi_looks, wflg)

    
    def __bool__(self):
        return all(bool(IW) for IW in self.IWs if IW is not None)

    
    def __str__(self):
        return "\n".join(str(IW) for IW in self.IWs if IW is not None)


def coreg(master, SLC, RSLC, hgt=0.1, rng_looks=10, azi_looks=2,
          poly1=None, poly2=None, cc_thresh=0.8, frac_thresh=0.01,
          ph_std_thresh=0.8, clean=True, use_inter=False, RSLC3=None,
          diff_dir="."):
    
    mslc = master["S1SLC"]
    
    cleaning = 1 if clean else 0
    flag1 = 1 if use_inter else 0
    
    SLC1_tab, SLC1_ID = mslc.tab, mslc.date.date2str()
    SLC2_tab, SLC2_ID = SLC.tab, SLC.date.date2str()
    
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
    
    ifg = IFG(ID + ".diff", ID + ".off", ID + ".diff_par",
              ID + ".coreg_quality")
    
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
