import gamma as gm

from logging import getLogger

log = getLogger("gamma.interferometry")

gp = gm.gamma_progs


__all__ = [
    "IFG",
    "base_plot"
]


class IFG(gm.DataFile):
    
    __slots__ = {"diff_par", "qual", "filt", "cc", "dt", "slc1", "slc2",
                 "sim_unw"}
    
    __save__ = __slots__
    
    
    _cc_weights = {
        "constant": 0,
        "gaussian": 1,
    }
    
    
    _off_algorithm = {
        "int_cc": 1,
        "fringe_vis": 2
    }
    

    def __init__(self, datfile, parfile=None, sim_unw=None, diff_par=None,
                 quality=None):
        self.dat = datfile
        
        base = pth.splitext(datfile)[0]
        
        if parfile is None:
            parfile = "%s.off" % base

        if sim_unw is None:
            sim_unw = "%s.sim_unw" % base
        

        self.par, self.qual, self.diff_par, self.filt, self.cc, \
        self.slc1, self.slc2, self.sim_unw = parfile, quality, diff_par, \
        None, None, None, None, sim_unw


    @classmethod
    def from_json(cls, line):
        return cls(line["dat"], line["par"], line["sim_unw"], 
                   line["diff_par"], line["qual"])

    
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
    

    def coherence(self, slope_win=5, weight_type="gaussian", corr_thresh=0.4,
                  box_lims=(3.0,9.0)):
        wgt_flag = IFG._cc_weights[weight_type]
        
        #log.info("CALCULATING COHERENCE AND CREATING QUICKLOOK IMAGES.")
        #log.info('Weight type is "%s"'.format(weight_type))
        
        width = ifg.rng()
        
        log.info("Estimating phase slope.", end=" ")
        gp.phase_slope(ifg.dat, slope, width, slope_win, corr_thresh)

        log.info("Calculating coherence.", end=" ")
        gp.cc_ad(ifg.dat, mli1, mli2, slope, None, ifg.cc, width, box_lims[0],
                 box_lims[1], wgt_flag)

    
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



def base_plot(midx, RSLCs, bperp_lims=(0.0, 150.0),
              delta_T_lims=(0.0, 15.0), SLC_tab="SLC_tab",
              bperp="bperp", itab="itab"):

    with open(SLC_tab, "w") as f:
        f.write("%s\n" % "\n".join(str(rslc) for rslc in RSLCs))
    
    mslc_par = RSLCs[midx].par
    
    gp.base_calc(SLC_tab, mslc_par, bperp, itab, 1, 1, bperp_lims[0],
                 bperp_lims[1], delta_T_lims[0], delta_T_lims[1])
    
    gp.base_plot(SLC_tab, mslc_par, itab, bperp, 1)

