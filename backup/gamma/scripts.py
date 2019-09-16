import os.path as pth
import math
from shutil import copy
from logging import getLogger

import gamma.base as gb
import gamma.private as gp


log = getLogger("gamma.scripts")

gm, gm_ras = gb.gm, gb._default_image_fmt


if gp.ScanSAR:
    interp_lt = getattr(gm, "SLC_interp_lt_ScanSAR")
else:
    interp_lt = getattr(gm, "SLC_interp_lt_S1_TOPS")


def ssi_int(SLC1, MLI1, hgt, lt_fine, off, SLC2, id1, id2, id,
            mode="nounwrap", outdir=".", cleaning=True):
    """
    SSI_INT: generate for an SLC pair the split-spectrum interferogram, v2.0 2-May-2018 uw
    SSI_INT <SLC1> <par1> <MLI1> <MLI_par1> <hgt> <lt_fine> <off> <SLC2> <par2> <mode> <ID1> <ID2> <ID> [dir] [cleaning]
    SLC1        (input) reference SLC data (e.g. 20070214.slc)
    par1        (input) reference SLC parameter file (e.g. 20070214.slc.par)
    MLI1        (input) reference MLI data (e.g. 20070214.mli)
    MLI_par1    (input) reference MLI parameter file (e.g. 20070214.mli.par)
    hgt         (input) terrain height in reference MLI geometry (e.g. 20070214.hgt)
    lt_fine     (input) co-registration lookup table (in reference MLI geometry (e.g. 20070214_20070228.lt_fine)
                the lookup table can be determined with rdc_trans + a polynomial refinement or
                it can be the result of a careful estimation of an offset field.
    off         (input) ISP offset parameter file used in SLC coregistration (contains offset refinement, e.g. 20070228.slc.off)
    SLC2        (input) slave SLC data (e.g. 20070228.slc, not co-registered to SLC1)
    par2        (input) slave SLC parameter file (e.g. 20070228.slc.par)
    mode        mode 1: generate complex valued split-spectrum interferogram only
                     2: generate complex valued and unwrapped split-spectrum interferogram
    ID1         ID to use for reference files generated (e.g. 20070214)
    ID2         ID to use for slave files generated (e.g. 20070228)
    ID          ID to use for files relating to reference and slave (e.g. 20070214_20070228)
    dir         directory to use for files generated (default = ./)
    cleaning    flag to remove many intermediate data sets (default=1: removing; 0: not removing)
    """
    
    # check if LAT programs are available:
    if hasattr(gm, "mask_class") is None:
        print("LAT program mask_class is available")
    else:
        raise RuntimeError("LAT program mask_class is not available")

    if hasattr(gm, "bpf_ssi") is None:
        print("script bpf_ssi is available")
    else:
        raise RuntimeError("script bpf_ssi is not available")

    ###########################################################################
    
    if pth.isfile(outdir):
        print("Indicated working directory ({}) exists.".format(outdir))
    else:
        gb.mkdir(outdir)
        
        if pth.isfile(outdir):
            print("Indicated working directory ({}) created.".format(outdir))
        else:
            raise RuntimeError("Cannot create indicated working directory ({})".format(outdir))
    
    # check scripts used are available:
    
    if mode != "nounwrap" or mode != "unwrap":
        raise OptionError('"mode" should be either "nounwrap" or unwrap"')
    
    # determine range and azimuth looks used
    
    rlks, azlks, width, nlines = \
    MLI1["range_looks"], MLI1["azimuth_looks"], MLI1.rng, MLI1.azi
    
    width_half, nlines_half = float(width) / 2.0, float(nlines) / 2.0
    
    ######################################
    
    print("Step 1: Apply band-pass filtering for split-spectrum "
          "interferometry.")
    
    low1 = DataFile(pth.join(outdir, id1 + ".slc.low")) # $dir/$id1.slc.flow
    high1 = DataFile(pth.join(outdir, id1 + ".slc.high")) # $dir/$id1.slc.fhigh

    gm.bpf_ssi(SLC1.datpar(), low1.datpar(), high1.datpar(), 0.6666)

    low2 = DataFile(pth.join(outdir, id2 + ".slc.low"))
    high2 = DataFile(pth.join(outdir, id2 + ".slc.high"))
    
    gm.bpf_ssi(SLC2.datpar(), low2.datpar(), high2.datpar(), 0.6666)
    
    print("Step 2: resample spectrally filtered slave slcs to reference.")
    
    mli2 = DataFile(pth.join(outdir, id2 + ".mli"))
    rslc2 = DataFile(pth.join(outdir, id2 + ".rslc"))
    rslc_low2 = DataFile(pth.join(outdir, id2 + ".rslc.flow"))
    rslc_high2 = DataFile(pth.join(outdir, id2 + ".rslc.fhigh"))
    
    gm.multi_look(SLC2.datpar(), mli2.datpar(), rlks, azlks)
    
    gm.SLC_interp_lt(SLC2.dat, SLC1.par, SLC2.par, lt_fine, MLI1.par, 
                     mli2.par, off, rslc2.datpar())
    
    gm.SLC_interp_lt(low2.dat, low1.par, low2.par, lt_fine,
                     MLI1.par, mli2.par, off, rslc_low2.datpar())

    gm.SLC_interp_lt(high2.dat, high1.par, high2.par, lt_fine,
                     MLI1.par, mli2.par, off, rslc_high2.datpar())
    
    #######################################################################
    
    print("Step 3: Calculate differential interferograms for full-"
          "bandwidth, low, and high sub-band.")
    
    # calculate differential interferometry for high sub-band
    off_high = pth.join(outdir, id + ".off.fhigh")
    sim_high = pth.join(outdir, id + ".sim.fhigh")
    diff_high = pth.join(outdir, id + ".diff.fhigh")
    diff_high_ras = diff_high + ".bmp"
    

    gm.create_offset(high1.par, rslc_high2.par, off_high, 1, rlks, azlks, 0)
    
    gm.phase_sim_orb(high1.par, rslc_high2.par, off_high, hgt, sim_high,
                     None, None, None, 1, 0)
    
    gm.SLC_diff_intf(high1.dat, rslc_high2.dat, high1.par,
                     rslc_high2.par, off_high, sim_high, diff_high,
                     rlks, azlks, 1, 0, 0.25)
    
    rng_avg, azi_avg = MLI1.avg_factor()
    
    gm.mph_pwr(diff_high, MLI1.dat, width, None, None, None,
               rng_avg, azi_avg, None, None, None, diff_high_ras)

    
    # calculate differential interferometry for low sub-band
    
    off_low = pth.join(outdir, id + ".off.flow")
    sim_low = pth.join(outdir, id + ".sim.flow")
    diff_low = pth.join(outdir, id + ".diff.flow")
    diff_low_ras = diff_low + ".bmp"
    
    gm.create_offset(low1.par, rslc_low2.par, off_low, 1, rlks, azlks, 0)
    
    gm.phase_sim_orb(low1.par, rslc_low2.par, off_low, hgt, sim_low,
                     None, None, None, 1, 0)
    
    gm.SLC_diff_intf(low1.dat, rslc_low2.dat, low1.par,
                     rslc_low2.par, off_low, sim_low, diff_low,
                     rlks, azlks, 1, 0, 0.25)
    
    gm.mph_pwr(diff_low, MLI1.dat, width, None, None, None,
               rng_avg, azi_avg, None, None, None, diff_low_ras)

    
    # calculate differential interferometry for full bandwidth

    off = pth.join(outdir, id + ".off")
    sim = pth.join(outdir, id + ".sim")
    diff = pth.join(outdir, id + ".diff")
    diff_ras = diff + ".bmp"
    
    # calculate differential interferometry for high sub-band
    gm.create_offset(SLC1.par, rslc2.par, off, 1, rlks, azlks, 0)
    
    gm.phase_sim_orb(SLC1.par, rslc2.par, off, hgt, sim,
                     None, None, None, 1, 0)
    
    gm.SLC_diff_intf(SLC1.dat, rslc2.dat, SLC1.par, rslc2.par, off,
                     sim, diff, rlks, azlks, 1, 0, 0.25)
    
    gm.mph_pwr(diff, MLI1.dat, width, None, None, None,
               rng_avg, azi_avg, None, None, None, diff_low_ras)

    if cleaning:
        remove(sim, sim_high, sim_low)
        
        for item in (rslc_high2, rslc_low2, high1, low1, high2, low2):
            item.remove()
        
    ########################################################################
    
    print("Step 4: Calculate, filter and unwrap double difference "
          "differential interferogram.")
    
    diff_low_phase = diff_low + ".phase"
    ddiff = pth.join(outdir, id + ".ddiff")
    ddiff_ras = ddiff + ".bmp"
    
    gm.cpx_to_real(diff_low, diff_low_phase, width, 4)

    gm.subtract_phase(diff_high, diff_low_phase, ddiff, width, 1)

    gm.rasmph_pwr(ddiff, MLI1.dat, width, None, None, None,
                  rng_avg, azi_avg, None, None, None, ddiff_ras)
    
    if mode == "unwrap":
        if not hasattr(gm, "mask_class"):
            raise RuntimeError("LAT program mask_class is not available.\n"
                               "LAT programs are required for mode "
                               "\"unwrap\".\n"
                               "If you don't have access the LAT you "
                               "can do the filtering and unwrapping of "
                               "the split-spectrum double-difference "
                               "interferogram outside SSI_INT.")
        else:
            print("LAT program mask_class is available")
            
            base = ddif + "_sm" 
            ddiff_sm = tuple(None, base + "1", base + "2", base + "3")
            
            sm4 = Multi(ddiff + ".sm4", phase=".phase", cc=".cc",
                        shifted=".shifted", shifted_phase=".shifted.phase",
                        cc_mask=".cc_mask.bmp")
    
            
            sm4_detrend = Multi(ddiff + ".sm4.detrend", phase="phase",
                                cc=".cc", mask=".masked",
                                mask_tmp=".masked.tmp")
            
            ddiffm = Multi(ddiff, phase_shift=".phase_shift",
                           phase_trend=".phase.trend",
                           phase_detrend=".phase.detrended", cc=".cc",
                           five=".5", sm_cc=".sm.cc",
                           cc_mask=".cc_mask.bmp", cc5=".cc.5",
                           detrend_interp=".detrended.interp.phase",
                           detrend_fspf=".detrended.phase.fspf",
                           phase_fspf=".phase.fspf",
                           phase_fspf_4pi=".phase.fspf.4PI.bmp",
                           phase_fspf_half_pi=".phase.fspf.half_PI.bmp")
            
            
            ddiff5 = Multi(ddiff + "5", phase_interp=".phase.interp",
                           phase_interp_fspf=".phase.interp.fspf",
                           phase_interp_outlier=".phase.interp.outlier",
                           phase_tmp=".phase.tmp", phase_tmp1=".phase.tmp1",
                           cc=".cc", mask=".mask.bmp")


            MLI5 = MLI1.dat + "5"

            off1 = pth.join(outdir, id + ".off1")
            off5 = pth.join(outdir, id + ".off5")
            diff_par = pth.join(outdir, id + ".ddiff.phase.diff_par")
            image_stat = pth.join(outdir, "image_stat.report")
            

            # filter ddiff
            gm.adf(ddiff, ddiff_sm[1], ddiffm.cc, width, 0.2, 512, 7, 128)
            gm.adf(ddiff_sm[1], ddiff_sm[2], ddiffm.cc, width, 0.3, 256, 7, 64)
            gm.adf(ddiff_sm[2], ddiff_sm[3], ddiffm.cc, width, 0.3, 128, 7, 32)
            gm.adf(ddiff_sm[3], sm4.base, ddiffm.cc, width, 0.3, 128, 7, 16)
            
            # unwrapping (values are << 1 radian)
            gm.cpx_to_real(sm4.base, sm4.phase, width, 4)
     
            # shift to zero (there may be an offset so that using cpx_to_real may not work correctly)
            # mask very low coherence: cc_mask.bmp
            gm.cc_ad(ddiff, MLI1.dat, MLI1.dat, None, None, ddiffm.cc_mask,
                     width, 3, 9)
            gm.rascc_mask(ddiffm.cc_mask, MLI1.dat, width, 1, 1, 0, 1, 1,
                          0.15, 0.0, 0.1, 0.9, 1.0, 0.35, 1, ddiffm.cc_mask)

            gm.image_stat(sm4.phase, width, width_half, nlines_half,
                          200, 200, image_stat)

            mean = get_par("mean", image_stat)
            stdev = get_par("stdev", image_stat)
            
            gm.lin_comb(1, sm4.phase, mean, 0.0, ddiffm.ph_shift,
                        width, 1, nlines)
            
            gm.subtract_phase(sm4.base, ddiffm.ph_shift, sm4.shifted,
                              width, 1.0)
            
            gm.cpx_to_real(sm4.shifted, sm4.shifted_phase, width, 4)
            
            gm.lin_comb(1, sm4.shifted_phase, mean, 1.0, sm4.phase,
                        width, 1, nlines)
            
            # determine trend considering cc_mask using unwrapping with 
            # multi_cpx before unwrapping
            
            gm.create_diff_par(MLI1.par, None, diff_par, 1, 0)
            gm.quad_fit(sm4.phase, diff_par, 5, 5, ddiffm.cc_mask,
                        None, 3, ddiffm.trend)
            
            gm.subtract_phase(ddiff, ddiffm.trend, ddiffm.detrend, width, 1.0)
    
            # filter ddiff_detrended
            gm.adf(ddiffm.detrend, ddiff_sm[1], ddiffm.sm_cc, width, 0.2, 512, 7, 128)
            gm.adf(ddiff_sm[1], ddiff_sm[2], ddiffm.sm_cc, width, 0.3, 256, 7, 64)
            gm.adf(ddiff_sm[2], ddiff_sm[3], ddiffm.sm_cc, width, 0.3, 128, 7, 32)
            gm.adf(ddiff_sm[3], sm4_detrend.base, ddiffm.sm_cc, width, 0.3, 128, 7, 16)
            
            gm.create_offset(MLI1.par, MLI1.par, off1, 1, 1, 1, 0)
            
            gm.rascc_mask(sm4.cc, MLI1.dat, width, 1, 1, 0, 1, 1,
                          0.95, 0.0, 0.1, 0.9, 1.0, 0.35, 1, sm4.cc_mask)
            
            gm.mask_class(ddiffm.cc_mask, sm4_detrend.base,
                          sm4_detrend.mask_tmp, 1, 1, 1, 1, 0, 0.0, 0.0)
            
            gm.mask_class(sm4.cc_mask, sm4_detrend.mask_tmp,
                          sm4_detrend.mask, 1, 1, 1, 1, 0, 0.0, 0.0)
            
            gm.multi_cpx(sm4_detrend.mask, off1, ddiff5.base, off5, 5, 5)
            gm.multi_real(MLI1.dat, off1, MLI5, off5, 5, 5)
            gm.multi_real(ddiffm.cc, off1, ddiff5.cc, off5, 5, 5)
            
            width5 = get_par("interferogram_width", off5)
            
            gm.cpx_to_real(ddiff5.base, ddiff5.phase_tmp, width5, 4)
            gm.fill_gaps(ddiff5.phase_tmp, width5, ddiff5.phase_interp,
                         4, None, 1, 100, 4, 400)
            
            # remove outliers
            width5 = get_par("interferogram_width", off5)
            
            gm.fspf(ddiff5.phase_interp, ddiff5.phase_interp_fspf,
                    width5, 2, 64, 3)
            
            gm.lin_comb(2, ddiff5.phase_interp, ddiff5.phase_interp_fspf,
                        100.0, 1.0, -1.0, ddiff5.phase_interp_outlier,
                        width5)
            
            gm.single_class_mapping(2, ddiff5.phase_interp_outlier,
                                    99.95, 100.05, ddiffm.cc5, 0.15,
                                    1.0, ddiff5.mask, width5)
            
            gm.mask_class(ddiff5.mask, ddiff5.phase_tmp,
                          ddiff5.phase_tmp1, 0, 1, 1, 1, 0, 0.0, 0.0)
            
            gm.fill_gaps(ddiff5.phase_tmp1, width5, ddiff5.phase_interp,
                         4, None, 1, 100, 4, 400)
            
            gm.multi_real(ddiff5.phase_interp, off5,
                          ddiffm.detrended_interp_phase, off1, -5, -5)
            
            gm.fspf(ddiffm.detrended_interp, ddiffm.detrended_fspf,
                    width, 2, 8, 3)
            
            # add the solutions
            gm.lin_comb(2, ddiffm.phase_trend, ddiffm.detrended_fspf,
                        0.0, 1.0, 1.0, ddiffm.phase_fspf, width, 1, nlines)
            
            gm.rasdt_pwr(ddiffm.phase_fspf, MLI1.dat, width, None, None,
                         None, rng_avg, azi_avg, 12.6, None, None, None,
                         ddiffm.phase_fspf_4PI)
            gm.rasdt_pwr(ddiffm.phase_fspf, MLI1.dat, width, None, None,
                         None, rng_avg, azi_avg, 12.6, None, None, None,
                         ddiffm.phase_fspf_half_PI)
            
            if cleaning:
                remove(*ddiff_sm[1:])
                
                remove(sm4.base, ddiffm.cc5,
                       ddiff5.phase_interp, ddiff5.phase_interp_fspf,
                       ddiff5.phase_interp_outlier, ddiff5.phase_tmp,
                       ddiff5.phase_tmp1, ddiffm.cc, ddiffm.detrend,
                       ddiffm.detrend_interp, ddiffm.detrend_fspf,
                       ddiffm.phase_shift, sm4.cc, sm4_detrend.base,
                       sm4_detrend.mask, sm4_detrend.mask_tmp,
                       sm4.phase, sm4.shifted, sm4.shifted_phase,
                       ddiffm.sm_cc, low_phase)
    
    #######################################################################
    
    print("*** end of SSI_INT ***")



def S1_coreg(SLC1, SLC2, RSLC2, hgt=0.1, rng_looks=10, azi_looks=2,
             poly1=None, poly2=None, cc_thresh=0.8, frac_thresh=0.01,
             ph_std_thresh=0.8, cleaning=True, use_inter=False,
             RSLC3=None):
    
    cleaning = 1 if cleaning else 0
    flag1    = 1 if use_inter else 0
    
    date1, date2 = SLC1.date.date2str(), SLC2.date.date2str()
    
    if RSLC3 is None:
        gm.S1_coreg_TOPS(SLC1.tab, date1, SLC2.tab, date2,
                         RSLC2.tab, hgt, rng_looks, azi_looks, poly1, poly2,
                         cc_thresh, frac_thresh, ph_std_thresh, cleaning,
                         flag1)
    else:
        gm.S1_coreg_TOPS(SLC1.tab, date1, SLC2.tab, date2,
                         RSLC2.tab, hgt, rng_looks, azi_looks, poly1, poly2,
                         cc_thresh, frac_thresh, ph_std_thresh, cleaning,
                         flag1, RSLC3.tab, RSLC3.date.date2str())

    
    with open("{}_{}.coreg_quality".format(date1, date2), "r") as f:
        offsets = [float(line.split()[1]) for line in f
                   if line.startswith("azimuth_pixel_offset")]
    
    
    if math.isclose(sum(offsets), 0.0):
        raise RuntimeError("Coregistration of {} to master {} failed!"
                           .format(date2, date1))
    
    

def S1_poly_overlap(SLC, rng_looks, azi_looks, poly, overlap_type="azi"):
    
    if overlap_type == "azi":
        log.info("Generate polygons for azimuth overlaps.")
    elif overlap_type == "rng":
        raise NotImplementedError("Range overlaps not yet implemented.")
        #log.info("Generate polygons for range overlaps.")
    else:
        raise RuntimeError("overlap_type: %s not supported!")
    
    IW1 = SLC.IWs[0]
    azimuth_line_time = IW1.getfloat("azimuth_line_time")
    range_pixel_spacing = IW1.getfloat("range_pixel_spacing")
    IW1_burst_start_time_1 = IW1.getfloat("burst_start_time_1")
    
    log.info("azimuth_line_time: %f\trange_pixel_spacing: %f"
             % (azimuth_line_time, range_pixel_spacing))
    
    log.info("Determine burst overlap regions within each sub-swath.")
    
    start_time = IW1_burst_start_time_1       
    
    for iw in SLC.IWs:
        if iw is None:
            continue
        
        burst_start_time_1 = iw.getfloat("burst_start_time_1")
        rows_offset = int((burst_start_time_1 - start_time) / azimuth_line_time)
        
        if rows_offset < 0:
            start_time = burst_start_time_1

    IW1_number_of_bursts = IW1.getint("number_of_bursts")
    IW1_burst_start_time_1 = IW1.getfloat("burst_start_time_1")
    IW1_burst_start_time_2 = IW1.getfloat("burst_start_time_2")
    IW1_first_valid_line = IW1.getint("first_valid_line_1")
    IW1_last_valid_line = IW1.getint("last_valid_line_1")
    IW1_near_range_slc = IW1.getfloat("near_range_slc")
    IW1_first_valid_sample = IW1.getfloat("first_valid_sample_1")
    IW1_last_valid_sample = IW1.getfloat("last_valid_sample_1")

    IW1_start_line = int(0.5 + (IW1_first_valid_line
                         + (IW1_burst_start_time_1 - start_time)
                         / azimuth_line_time) / azi_looks)

    IW1_stop_line = int(0.5 + (IW1_last_valid_line
                        + (IW1_burst_start_time_1 - start_time)
                        / azimuth_line_time) / azi_looks)
    
    IW1_start_sample = int(0.5 + (IW1_first_valid_sample / rng_looks))
    IW1_stop_sample = int(0.5 + (IW1_last_valid_sample / rng_looks))
    
    IW1_rows_offset = float((IW1_burst_start_time_2 - IW1_burst_start_time_1)
                            / azimuth_line_time / azi_looks)


    if overlap_type == "azi":
        gb.Files.rm(poly)
        
        for ii, iw in enumerate(SLC.IWs):
            if iw is None:
                continue
            
            # IW1 already processed
            if ii == 0:
                number_of_bursts = IW1_number_of_bursts
                first_valid_sample = IW1_first_valid_sample
                last_valid_sample = IW1_last_valid_sample
                start_sample = IW1_start_sample
                stop_sample = IW1_stop_sample
                burst_start_time_1 = IW1_burst_start_time_1
                burst_start_time_2 = IW1_burst_start_time_2
                
                first_valid_line = IW1_first_valid_line
                last_valid_line = IW1_last_valid_line
                
                start_line = IW1_start_line
                stop_line = IW1_stop_line
                
                rows_offset = IW1_rows_offset
            else:
                number_of_bursts = iw.getint("number_of_bursts")
                
                first_valid_sample = iw.getfloat("first_valid_sample_1")
                last_valid_sample = iw.getfloat("last_valid_sample_1")
                near_range_slc = iw.getfloat("near_range_slc")
                
                start_sample = \
                int(0.5 + (first_valid_sample
                    + (near_range_slc - IW1_near_range_slc) / range_pixel_spacing)
                    / rng_looks)
                
                stop_sample = \
                int(0.5 + (last_valid_sample
                    + (near_range_slc - IW1_near_range_slc) / range_pixel_spacing)
                    / rng_looks)
    
                
                burst_start_time_1 = iw.getfloat("burst_start_time_1")
                burst_start_time_2 = iw.getfloat("burst_start_time_2")
                first_valid_line = iw.getint("first_valid_line_1")
                last_valid_line = iw.getint("last_valid_line_1")
    
                start_line = \
                int(0.5 + (first_valid_line + (burst_start_time_1 - start_time)
                    / azimuth_line_time) / azi_looks)
                
                stop_line = \
                int(0.5 + (first_valid_line + (burst_start_time_1 - start_time)
                    / azimuth_line_time) / azi_looks)
    
                rows_offset = \
                float((burst_start_time_2 - burst_start_time_1)
                      / azimuth_line_time / azi_looks)
    
            idx = ii + 1
            
            log.info("IW%d_valid_lines: %d to %d  IW%d_valid_samples: %d to %d."
                     % (idx, start_line, stop_line, idx,
                        start_sample, stop_sample))
            
            log.info("IW%d_rows_offset: %f" % (idx, rows_offset))
            
            for jj in range(1, number_of_bursts + 1):
                
                az1 = int(0.5 + start_line + (jj * rows_offset))
                az2 = int(0.5 + stop_line + ((jj - 1.0) * rows_offset))
                
                log.info("ii, jj: %d,%d  mli_cols: %d to %d   "
                         "mli_rows: %d to %d" % (idx, jj, start_sample,
                         stop_sample, az1, az2))
                
                with open(poly, "a") as f:
                    f.write("\t%d\t%d\t1\n" % (start_sample, az1))
                    f.write("\t%d\t%d\t2\n" % (stop_sample, az1))
                    f.write("\t%d\t%d\t3\n" % (stop_sample, az2))
                    f.write("\t%d\t%d\t4\n" % (start_sample, az2))
    
    log.info("Azimuth overlap polygon file: %s" % poly)

    
def S1_coreg_TOPS(SLC1, SLC2, RSLC2, hgt=0.1, rng_looks=10, azi_looks=2,
                  cleaning=True, itmax=5, flag1=False, poly1=None, poly2=None,
                  RSLC3=None, coreg_dir=".", **kwargs):
    """
    S1_coreg_TOPS: Script to coregister a Sentinel-1 TOPS mode burst SLC to a reference burst SLC v1.5 21-Nov-2016 uw"

    usage: S1_coreg_TOPS <SLC1_tab> <SLC1_ID> <SLC2_tab> <SLC2_ID> <RSLC2_tab> [hgt] [RLK] [AZLK] [poly1] [poly2] [cc_thresh] [fraction_thresh] [ph_stdev_thresh] [cleaning] [flag1] [RSLC3_tab]"
        SLC1_tab    (input) SLC_tab of S1 TOPS burst SLC reference (e.g. 20141015.SLC_tab)"
        SLC1_ID     (input) ID for reference files (e.g. 20141015)"
        SLC2_tab    (input) SLC_tab of S1 TOPS burst SLC slave (e.g. 20141027.SLC_tab)"
        SLC2_ID     (input) ID for slave files (e.g. 20141027)"
        RSLC2_tab   (input) SLC_tab of co-registered S1 TOPS burst SLC slave (e.g. 20141027.RSLC_tab)"
        hgt         (input) height map in RDC of MLI-1 mosaic (float, or constant height value; default=0.1)"
        RLK         number of range looks in the output MLI image (default=10)"
        AZLK        number of azimuth looks in the output MLI image (default=2)"
        poly1       polygon file indicating area used for matching (relative to MLI reference to reduce area used for matching)"
        poly2       polygon file indicating area used for spectral diversity (relative to MLI reference to reduce area used for matching)"
        cc_thresh   coherence threshold used (default = 0.8)"
        fraction_thresh   minimum valid fraction of unwrapped phase values used (default = 0.01)"
        ph_stdev_thresh   phase standard deviation threshold (default = 0.8)"
        cleaning    flag to indicate if intermediate files are deleted (default = 1 --> deleted,  0: not deleted)"
        flag1       flag to indicate if existing intermediate files are used (default = 0 --> not used,  1: used)"
        RSLC3_tab   (input) 3 column list of already available co-registered TOPS slave image to use for overlap interferograms"
        RSLC3_ID    (input) ID for already available co-registered TOPS slave; if indicated then the differential interferogram between RSLC3 and RSLC2 is calculated"

    # History:
    # 14-Jan-2015: checked/updated that SLC and TOPS_par in RSLC2_tab are correctly used
    #              the only use of RSLC2_tab is when calling S1_coreg_overlap script 
    #              S1_coreg_overlap uses only the burst SLC name in RSLC2_tab but not the burst SLC parameter filename or TOPS_par
    #              --> correct even with corrupt TOPS_par in RSLC2_tab
    # 14-Jan-2015: changed script to apply matching offset to refine lookup table
    # 15-Jan-2015: added generation of a quality file $p.coreg_quality
    # 15-Jan-2015: added checking/fixing of zero values in burst 8 parameters of TOPS_par files
    # 29-May-2015: added poly2: area to consider for spectral diversity
    #  9-Jun-2015: checking for availability of LAT programs - if not available use entire lookup table (--> slower)
    # 19-Jun-2015: modified to limit maximum number of offset estimation in matching
    #  9-Sep-2015: corrected reading of parameter RSLC3_tab
    # 23-Nov-2015: updated for modifications in offset_pwr_tracking (--> new threshold is 0.2)
    # 25-Nov-2015: added RSLC3_ID option to also calculate differential interferogram between RSLC3 and RSLC2
    #  1-Dec-2015: introduced a maximum number of interations
    # 10-Jun-2016: printing out some more of the commands used
    # 22-Sep-2016:  added flag indicating if height file exists; modified phase simulation if that is not the case
    #  5-Oct-2016: define $ras raster image file type
    # 21-Nov-2016: adapted program for EWS SLC coregistration (with up to 5 sub-swaths / resp. lines in the SLC_tab)
    """

    ###########################################################################

    if hasattr(gm, "poly_math"):
        print("LAT program poly_math is available.")
        poly_math = True
    else:
        poly_math = False
        raise RuntimeError("LAT program poly_math is not available!")

    if hasattr(gm, "mask_class"):
        print("LAT program mask_class is available")
    else:
        raise RuntimeError("LAT program mask_class is not available!")

    ############################################################################

    if SLC1 == SLC2:
        print("Indicated SLC is reference SLC --> proceed")
        return 1
    
    if SLC2 == RSLC2:
        raise RuntimeError("SLC files are identical for slave and "
                           "resampled slave!")
    
    
    SLC1_ID, SLC2_ID = SLC1.date.date2str(), SLC2.date.date2str()
    SLC1_tab, SLC2_tab, RSLC2_tab = SLC1.tab, SLC2.tab, RSLC2.tab

    outdir = gb.mkdir(pth.join(coreg_dir, SLC2_ID))
    
    p = gb.Base(pth.join(outdir, ""), off="off", off_start="off_start",
                offs="offs", doff="doff", snr="snr", sim_unw="sim_unw",
                diff_ras="diff_ras", diff_par="diff_par")
    
    log.info("Raster image file extension : %s." % gm_ras)

    SLC1_path, SLC2_path = pth.join(coreg_dir, SLC1_ID), pth.join(outdir, SLC2_ID)
    
    
    slave = gb.Multi(
        slc  = gb.SLC(SLC2_path + ".slc"),
        rslc = gb.SLC(SLC2_path + ".rslc"),
        mli  = gb.MLI(SLC2_path + ".mli"),
        rmli = gb.MLI(SLC2_path + ".rmli")
    )
    
    
    ref = gb.Multi(
        rslc = gb.SLC(SLC1_path + ".rslc"),
        rmli = gb.MLI(SLC1_path + ".rmli")
    )
    
    
    MLI = gb.Base(pth.join(outdir, ""), ras="mli.%s" % gm_ras,
                  lt="lt", lt_masked="lt.masked", lt_az_ovr="lt.az_ovr")
    
    REF_MLI = gb.Base(pth.join(SLC1_path), az_ovr=".az_ovr",
                      az_ovr_ras=".az_ovr.%s" % gm_ras, ras=".%s" % gm_ras,
                      az_ovr2=".az_ovr2", az_ovr2_ras=".az_ovr2.%s" % gm_ras,
                      masked=".masked", masked_ras=".masked.%s" % gm_ras)
    
    print("Test if required input/output files and directories exist.")
    
    if not pth.isfile(SLC1_tab):
        raise RuntimeError("SLC1_tab file (%s) does not exist" % SLC1_tab)

    if not pth.isfile(SLC2_tab):
        raise RuntimeError("SLC2_tab file (%s) does not exist!" % SLC2_tab)

    if not pth.isfile(RSLC2_tab):
        raise RuntimeError("RSLC2_tab file (%s) does not exist!" % RSLC2_tab)
    
    if RSLC3 is not None and not pth.isfile(RSLC3.tab):
        raise RuntimeError("RSLC3_tab file (%s) does not exist!" % RSLC3_tab)
    
    if RSLC3 is None:
        RSLC3_tab = ""
    else:
        RSLC3_tab = RSLC3.tab

    if pth.isfile(hgt):
        print("Using the height file (%s)." % hgt)
        hgt_file_flag = True
    else:
        print('Height  parameter "hgt" is not a file (%s), '
              'using a constant height value (%s).' % (hgt, hgt))
        hgt_file_flag = False


    if poly1 is None:
        print("No polygon poly1 indicated.")
    elif not pth.isfile(poly1):
        raise RuntimeError("Polygon file indicated (%s) does not "
                           "exist" % poly1)

    if poly2 is None:
        print("No polygon poly2 indicated.")
    elif not pth.isfile(poly2):
        raise RuntimeError("Polygon file indicated (%s) does not "
                           "exist" % poly2)

    log.info("Required input files exist.")

    ###########################################################################
    

    log.info("Sentinel-1 TOPS coregistration quality\n"
             "######################################\n"
             "{SLC1_ID}\n"
             "S1_coreg_TOPS TODO: print parameters\n"
             "reference: {SLC1_ID} {REF_SLC.dat} {REF_SLC.par} {}\n"
             "slave: {SLC2_ID} {SLC.dat} {SLC.par} {}\n"
             "coregistered_slave: {SLC2_ID} {RSLC.dat} {RSLC.par} {}\n"
             "reference for spectral diversity refinement:       {}\n"
             "polygon used for matching (poly1):            {}\n"
             "polygon used for spectral diversity (poly2):  {}\n"
             .format(SLC1_tab, SLC2_tab, RSLC2_tab, RSLC3_tab, poly1, poly2,
                     SLC1_ID=SLC1_ID, REF_SLC=ref.rslc, SLC2_ID=SLC2_ID,
                     SLC=slave.slc, RSLC=slave.rslc))

    ###########################################################################
    

    if ref.rslc and flag1:
        log.info("Using existing SLC mosaic of reference: %s"
                 % ref.rslc.datpar)
    else:
        SLC1.mosaic(ref.rslc, rng_looks, azi_looks)


    if ref.rmli and flag1:
        log.info("Using existing MLI mosaic of reference: %s" % ref.rmli.datpar)
    else:
        ref.rslc.multi_look(ref.rmli, rng_looks=rng_looks, azi_looks=azi_looks)
    
    REF_MLI_width, REF_MLI_nlines = ref.rmli.rng(), ref.rmli.azi()
    REF_SLC_width, REF_SLC_nlines = ref.rslc.rng(), ref.rslc.azi()
    
    
    if REF_MLI.exist("ras") and flag1:
        log.info("Using existing rasterfile of MLI mosaic of reference: "
                 "%s" % REF_MLI.ras)
    else:
        ref.rmli.raster(raster=REF_MLI.ras)
    
    ###########################################################################
    
    if slave.slc and flag1:
        log.info("Using existing SLC mosaic of slave: %s" % slave.slc)
    else:
        gm.SLC_mosaic_S1_TOPS(SLC2_tab, slave.slc, rng_looks, azi_looks)

    
    if slave.mli and flag1:
        log.info("Using existing MLI mosaic of slave: %s" % slave.mli)
    else:
        slave.slc.multi_look(slave.mli, rng_looks=rng_looks, azi_looks=azi_looks)

    MLI_width, MLI_nlines = slave.mli.rng(), slave.mli.azi()
    

    if MLI.exist("ras") and flag1:
        log.info("Using existing rasterfile of MLI mosaic of slave: %s" % MLI.ras)
    else:
        slave.mli.raster(raster=MLI.ras)

    ###########################################################################
    

    if MLI.exist("lt") and flag1:
        log.info("Using existing lookup table: %s." % MLI.lt)
    else:
        gm.rdc_trans(ref.rmli.par, hgt, slave.mli.par, MLI.lt)
        #ref.rmli.rdc_trans(hgt, slave.mli, MLI.lt)
    
    
    if poly1 is None:
        if MLI.exist("lt_masked") and flag1:
            log.info("Using existing masked lookup table: %s." % MLI.lt_masked)
        else:
            MLI.ln("lt", MLI.lt_masked)
    else:
        if pth.isfile(poly1):
            if MLI.exist("lt_masked") and flag1:
                log.info("Using existing masked lookup table: %s." % MLI.lt_masked)
        else:
            if hasattr(gm, "poly_math"):
                gm.poly_math(ref.rmli.dat, REF_MLI.masked, REF_MLI_width,
                             poly1, None, 1, 0.0, 1.0)
                gm.raspwr(REF_MLI.masked, REF_MLI_width, 1, 0, 1, 1, 1.0,
                          0.35, 1, REF_MLI.masked_ras)
                gm.mask_class(REF_MLI.masked_ras, MLI.lt,
                              MLI.lt_masked, 1, 1, 1, 1, 0, 0.0, 0.0)
            else:
                MLI.lt.ln(MLI.lt_masked)
      
    ###########################################################################
    
    # determine starting and ending rows and cols in polygon file
    # used to speed up the offset estimation
    
    r, a = [0, int(REF_SLC_width)], [0, int(REF_SLC_nlines)]

    if poly1 is not None:
        with open(poly1, "r") as f:
            tmp = tuple(
                    tuple(
                        int(elem) for elem in line.split()[:2]
                    )
                    for ii, line in enumerate(f) if ii >= 0
                )
        
        nrows = len(tmp)

        r, a = [tmp[0][0], tmp[0][0]], [tmp[0][1], tmp[0][1]]
        
        for n in (2, 3, 4):
            if nrows >= n:
                r_ = tmp[n - 1][0]
                
                if r_ < r[0]: r[0] = r_[0]
                if r_ > r[1]: r[1] = r_[1]

        for n in (2, 3, 4):
            if nrows >= n:
                a_ = tmp[n - 1][1]
                
                if a_ < a[0]: a[0] = a_
                if a_ > a[1]: a[1] = a_
        
        log.info("[r1, r2]: %s\t[a1, a2]: %s" % (r, a))
        r, a = [r[0] * rng_looks, r[1] * rng_looks], \
               [a[0] * azi_looks, a[1] * azi_looks]

        log.info("[r1, r2]: %s\t[a1, a2]: %s" % (r, a))
        
    # endif
    
    ###########################################################################
    
    # reduce offset estimation to 64 x 64 samples max
    
    rstep1, rstep2 = 64, int((r[1] - r[0]) / 64.0)
    
    if rstep1 > rstep2:
        rstep = rstep1
    else:
        rstep = rstep2

    azstep1, azstep2 = 32, int((a[1] - a[0]) / 64.0)
    
    if azstep1 > azstep2:
        azstep = atstep1
    else:
        azstep = azstep2
    log.info("rstep, azstep: %d, %d" % (rstep, azstep))

    ###########################################################################
    ###########################################################################
    ###########################################################################

    # Iterative improvement of refinement offsets between master SLC and
    # resampled slave RSLC  using intensity matching (offset_pwr_tracking)
    # Remarks: here only a section of the data is used if a polygon is
    # indicated the lookup table is iteratively refined refined with the 
    # estimated offsets only a constant offset in range and azimuth 
    # (along all burst and swaths) is considered 
    
    log.info("Iterative improvement of refinement offset using matching:")
    
    p.rm("off")

    #ref.rslc.create_offset(slave.slc, p.off, flags.int_cc, rng_looks,
                           #azi_looks, flags.non_inter)

    gm.create_offset(ref.rslc.par, slave.slc.par,
                     p.off, flags.int_cc, rng_looks, azi_looks, flags.non_inter)
    
    if 1:
        for it in range(itmax):
            log.info("Offset refinement using matching iteration %d" % it)
            
            p.cp("off", p.off_start)
            
            gb.raster(MLI.lt_masked, parfile=ref.rmli.par, mode="mph")
            
            out = \
            interp_lt(SLC2_tab, slave.slc.par, SLC1_tab, ref.rslc.par,
                      MLI.lt_masked, ref.rmli.par, slave.mli.par, p.off_start,
                      RSLC2_tab, slave.rslc)
            
            
            with open(pth.join(coreg_dir, "SLC_interp_lt_S1_TOPS.1.out"), "wb") as f:
                f.write(out)
    
            
            p.rm("doff")
            gm.create_offset(ref.rslc.par, slave.rslc.par, p.doff, flags.int_cc,
                             rng_looks, azi_looks, flags.non_inter)
            
            # no oversampling as this is not done well because of the doppler ramp
    
            gm.offset_pwr_tracking(ref.rslc.dat, slave.rslc.dat,
                                   ref.rslc.par, slave.rslc.par,
                                   p.doff, p.offs, p.snr, 128, 64, None, 1, 0.2,
                                   rstep, azstep, r[0], r[1], a[0], a[1])
    
            out = gm.offset_fit(p.offs, p.snr, p.doff, None, None, 0.2, 1, 0)
            
            line = tuple(line for line in out.decode().split("\n")
                         if "final model fit std. dev. (samples) range:" in line)[0]
            
            split = line.split()
            rng_std, azi_std = split[7], split[9]
            
            daz, dr = p.getfloat("doff", "azimuth_offset_polynomial"), \
                      p.getfloat("doff", "range_offset_polynomial")
    
            daz10000, daz_mli = daz * 10000, daz / azi_looks
    
            log.info("daz10000: %f" % daz10000)
            log.info("daz_mli: %f" % daz_mli)
            
            
            # lookup table refinement
            # determine range and azimuth corrections for lookup table (in mli pixels)
            
            dr_mli, daz_mli = dr / rng_looks, daz / azi_looks
            log.info("dr_mli: %f\tdaz_mli: %f" % (dr_mli, daz_mli))
            p.rm("diff_par")
            
            gm.create_diff_par(ref.rmli.par, ref.rmli.par, p.diff_par, 1, 0)
            
            tpl = "%f   0.0000e+00   0.0000e+00   0.0000e+00   0.0000e+00   0.0000e+00"
            
            p.set("diff_par", "range_offset_polynomial", new=tpl % dr_mli)
            p.set("diff_par", "azimuth_offset_polynomial", new=tpl % daz_mli)
            
            it_diff = "%sdiff_par.%d" % (p.base, it)
                             
            p.cp("diff_par", it_diff)
    
            it_lt = "%slt.masked.%d" % (MLI.base, it)
            
            # if poly1 exists then update unmasked and masked lookup table
            if poly1 is not None:
                it_lt_maskeded = "%slt.masked.tmp.%d" % (slave.mli.dat, it)
    
                MLI.move("lt_masked", it_lt_masked)
                gm.gc_map_fine(it_lt, REF_MLI_width, p.diff_par, MLI.lt_masked, 1)
                
                MLI.move("lt", it_lt)
                gm.gc_map_fine(it_lt, REF_MLI_width, p.diff_par, MLI.lt, 1)
            else:
                MLI.move("lt", it_lt)
                gm.gc_map_fine(it_lt, REF_MLI_width, p.diff_par, MLI.lt, 1)
    
            log.info("matching_iteration_%d: %f %f    %f %f (daz dr   "
                     "daz_mli dr_mli)" % (it, daz, dr, daz_mli, dr_mli))
            log.info("matching_iteration_stdev_%d: %s %s (azimuth_stdev "
                     "range_stdev)" % (it, azi_std, rng_std))
            
            if abs(daz10000) < 100:
                break
        # end for
    
    ###########################################################################
    ###########################################################################
    ###########################################################################

    # Iterative improvement of azimuth refinement using spectral diversity method   
    # Remark: here only a the burst overlap regions within the indicated polygon
    # area poly2 are considered
    
    # determine mask for polygon region poly2 that is at the same
    # time part of the burst overlap regions
    
    az_ovr_poly = "%s.az_ovr.poly" % SLC1_ID
    gb.rm(az_ovr_poly)
    
    if poly_math:
        S1_poly_overlap(SLC1, rng_looks, azi_looks, az_ovr_poly, "azi")
        gm.poly_math(ref.rmli.dat, REF_MLI.az_ovr, REF_MLI_width, az_ovr_poly,
                     None, 1, 0.0, 1.0)
        gm.raspwr(REF_MLI.az_ovr, REF_MLI_width, 1, 0, 1, 1, 1.0, 0.35, 1,
                  REF_MLI.az_ovr_ras)
        gm.mask_class(REF_MLI.az_ovr_ras, MLI.lt, MLI.lt_az_ovr,
                      1, 1, 1, 1, 0, 0.0, 0.0)
    else:
        gb.ln(MLIm.lt, MLIm.lt_az_ovr)

    # further reduce lookup table coverage to area specified by polygon poly2
    
    if poly2 is not None:
        if poly_math:
            gm.poly_math(REF_MLI.az_ovr, REF_MLI.az_ovr2, REF_MLI_width, poly2,
                         None, 1, 0.0, 1.0)
            gm.raspwr(REF_MLI.az_ovr2, REF_MLI_width, 1, 0, 1, 1, 1.0, 0.35, 1,
                      REF_MLI.az_ovr2_ras)
            gm.mask_class(REF_MLI.az_ovr2_ras, MLI.lt, MLI.lt_az_ovr,
                          1, 1, 1, 1, 0, 0.0, 0.0)
        else:
            gb.ln(MLI.lt, MLI.lt_az_ovr)

    log.info("Iterative improvement of refinement offset azimuth overlap regions:")
    
    # iterate while azimuth correction >= 0.0005 SLC pixel
    
    for it in range(itmax):
        log.info("offset refinement using spectral diversity in azimuth "
                 "overlap region iteration %d" % it)
        
        p.cp("off", p.off_start)
        
        out = \
        interp_lt(SLC2_tab, slave.slc.par, SLC1_tab, ref.rslc.par,
                  MLI.lt_az_ovr, ref.rmli.par, slave.mli.par, p.off_start,
                  RSLC2_tab, slave.rslc)

        with open(pth.join(coreg_dir, "SLC_interp_lt_S1_TOPS.2.out"), "wb") as f:
            f.write(out)
        
        it_off = "%s.az_ovr.%s" % (p.off, it)
        it_out = it_off + ".out"

        S1_coreg_overlap(SLC1, RSLC2, p.off_start, p.off, outdir=coreg_dir,
                         RSLC3=RSLC3, **kwargs)


        daz = float(Files.get_par("azimuth_pixel_offset", it_off).split()[0])
        daz10000 = daz * 10000
        log.info("daz10000: %f" % daz10000)
        
        p.cp("off", it_off)
        
        log.info("az_ovr_iteration_%d: %f (daz in SLC pixel)" % (it, daz))
        
        if abs(daz10000) < 5:
            break
        
        # more $p.results >>  $p.coreg_quality
    # end for

    ###########################################################################
    ###########################################################################
    ###########################################################################
    
    # resample full data set
    with open(pth.join(coreg_dir, "SLC_interp_lt_S1_TOPS.3.out"), "wb") as f:
        f.write(gm.SLC_interp_lt_S1_TOPS(SLC2_tab, slave.slc.par, SLC1_tab,
                ref.rslc.par, MLI.lt, ref.rmli.par, slave.mli.par, p.off,
                RSLC2_tab, slave.rslc.datpar))
    
    
    # topographic phase simulation 
    if pth.isfile(p.sim_unw) and flag1:
        log.info("Using existing simulated phase: %s." % p.sim_unw)
    else:
        # hgt file exists
        if hgt_file_flag:
            gm.phase_sim_orb(ref.rslc.par, slave.slc.par, p.off, hgt, p.sim_unw,
                             ref.rslc.par, None, None, 1, 1)
        else:
            gm.phase_sim_orb(ref.rslc.par, slave.slc.par, p.off, None, p.sim_unw,
                             ref.rslc.par, None, None, 1, 1)
        

    # calculation of a S1 TOPS differential interferogram
    gm.SLC_diff_intf(ref.rslc.dat, slave.rslc.dat,
                     ref.rslc.par, slave.rslc.par, p.off, p.sim_unw, p.diff,
                     rng_looks, azi_looks, 1, 0, 0.2, 1, 1)
    
    gm.rasmph_pwr24(p.diff, ref.rmli.dat, REF_MLI_width, 1, 1, 0, 1, 1, 1.0,
                    0.35, 1, p.diff_ras)

    log.info("Generated differential interferogram %s." % p.diff)
    log.info("to display use:   eog %s & " % p.diff_ras)


def S1_coreg_overlap(RSLC1, RSLC2, off, off_out, RSLC3=None, outdir=".",
                     **kwargs):
    """
    S1_coreg_overlap <RSLC1_tab> <RSLC2_tab> <pair> <off> <off_out> [cc_thresh] [fraction_thresh] [ph_stdev_thresh] [cleaning] [RSLC3_tab]"
    RSLC1_tab   (input) 3 column list of TOPS master image (SLC, SLC_par, TOPS_par; row order IW1, IW2, IW3)"
    RSLC2_tab   (input) 3 column list of TOPS slave image (SLC, SLC_par, TOPS_par; row order IW1, IW2, IW3)"
    pair        (input) ID used for InSAR (e.g. 20141003_20141015)"
    off         (input) offset parameter file (with refinement offset polynomials)"
    off_out     (output) corrected offset parameter file (with refinement offset polynomials)"
    cc_thresh   coherence threshold used (default = 0.8)"
    fraction_thresh   minimum valid fraction of unwrapped phase values used (default = 0.01)"
    ph_stdev_thresh   phase standard deviation threshold (default = 0.8 radian)"
    cleaning    flag to indicate if intermediate files are deleted (default = 1 --> deleted,  0: not deleted)"
    RSLC3_tab   (input) 3 column list of already available  co-registered TOPS slave image to use for overlap interferograms"
    """
    
    samples, Sum, samples_all, sum_all, sum_weight_all = 0, 0.0, 0, 0.0, 0.0
    
    cc_thresh = kwargs.get("cc_thresh", 0.8)
    
    frac_thresh, ph_std_frac = \
    kwargs.get("frac_thresh", 0.01), kwargs.get("ph_std_thresh", 0.8)
    
    frac10000_thresh, std10000_thresh = \
    10000 * frac_thresh, 10000 * ph_std_frac
    
    cleaning = kwargs.get("cleaning", True)
    
    res = pth.join(outdir, "results")
    
    with open(res, "w") as f:
        f.write("thresholds applied: cc_thresh: %f,  ph_fraction_thresh: %f,"
                 "ph_stdev_thresh (rad): %f" % (cc_thresh, frac_thresh,
                 ph_std_frac))
        f.write("IW  overlap  ph_mean ph_stdev ph_fraction   (cc_mean "
                 "cc_stdev cc_fraction)    weight")

    log.info("Test if required input/output files and directories exist")

    if not RSLC1:
        raise RuntimeError("RSLC1 \n%s\n does not exist" % RSLC1)

    if not RSLC2:
        raise RuntimeError("RSLC2 \n%s\n does not exist" % RSLC2)

    if not pth.isfile(off):
        raise RuntimeError("offset parameter file (%s) does not exist" % off)
    
    isRSLC3 = RSLC3 is not None
    
    if isRSLC3 and not RSLC3:
        raise RuntimeError("RSLC3 (%s) does not exist" % RSLC3_tab)

    ###########################################################################
    
    nIW = RSLC1.num_IWs()
    
    loffsets1 = tuple(iw.lines_offset() for iw in RSLC1.IWs if iw is not None)
    
    log.info("RSLC1: %s" % RSLC1)
    for ii, offset in enumerate(loffsets1):
        log.info("lines_offset_IW%d: %f %d" % (ii + 1, offset.f, offset.i))

    loffsets2 = tuple(iw.lines_offset() for iw in RSLC2.IWs if iw is not None)
    
    log.info("RSLC2: %s" % RSLC2)
    for ii, offset in enumerate(loffsets2):
        log.info("lines_offset_IW%d: %f %d" % (ii + 1, offset.f, offset.i))
    
    
    azimuth_line_time = RSLC1.IWs[0].getfloat("azimuth_line_time")
    dDC = 1739.43 * azimuth_line_time * loffsets1[0].i
    log.info("dDC %f Hz" % dDC)
    
    dt = 0.159154 / dDC
    log.info("dt %f s" % dt)

    dpix_factor = dt / azimuth_line_time
    log.info("dpix_factor %d azimuth pixel" % dpix_factor)
    log.info("azimuth pixel offset = %f  * average_phase_offset" % dpix_factor)

    ###########################################################################
    # determine phase offsets for sub-swath overlap regions of 
    # first/second sub-swaths
    
    for ii in range(nIW):
        IW1, IW2 = RSLC1.IWs[ii], RSLC2.IWs[ii]
        
        number_of_bursts_IW = IW1.getint("number_of_bursts")
        lines_per_burst = IW1.getint("lines_per_burst")
        loffsets = IW1.lines_offset()
        lines_offset = loffsets.i
        lines_overlap = lines_per_burst - lines_offset
        range_samples = IW1.rng()
    
        samples = 0
        Sum = 0.0
        sum_weight = 0.0
    
        jj = 1
        while jj < number_of_bursts_IW:
            p = gb.Base(pth.join(outdir, "IW%d.%d" % (ii, jj), ""),
                        off1="off1", off2="off2", int1="int1", int2="int2",
                        diff="diff", diff_par="diff_par", diff20="diff20",
                        off20="off20", cc="cc20", cc_ras="cc20.%s" % gm_ras,
                        adf="diff20.adf", adf_cc="diff20.adf.cc",
                        phase="diff20.phase")

            starting_line1 = lines_offset + (jj - 1) * lines_per_burst
            starting_line2 = jj * lines_per_burst
            log.info("%d %f %d" % (ii, starting_line1, starting_line2))
        
            R1IW1 = gb.SLC("%s.%d.1" % (IW1.dat, jj))
            R1IW2 = gb.SLC("%s.%d.2" % (IW1.dat, jj))
    
            R2IW1 = gb.SLC("%s.%d.1" % (IW2.dat, jj))
            R2IW2 = gb.SLC("%s.%d.2" % (IW2.dat, jj))
        
            if isRSLC3:
                IW_s = RSLC3.IW[ii]
            else:
                IW_s = IW1

            gm.SLC_copy(IW_s.dat, IW_s.par, IW1.par, R1IW1, None, 1.0, 0,
                        range_samples, starting_line1, lines_overlap)
    
            gm.SLC_copy(IW_s.dat, IW_s.par, IW1.par, R1IW2, None, 1.0, 0,
                        range_samples, starting_line2, lines_overlap)

                
            gm.SLC_copy(IW2.dat, IW2.par, IW1.par, R2IW1, None, 1.0, 0,
                        range_samples, starting_line1, lines_overlap)

            gm.SLC_copy(IW2.dat, IW2.par, IW1.par, R2IW2, None, 1.0, 0,
                        range_samples, starting_line2, lines_overlap)
        
            # calculate the 2 single look interferograms for the burst overlap region i
            # using the earlier burst --> *.int1, using the later burst --> *.int2
            
            p.rm("off1", "off2", "diff_par")
            
            gm.create_offset(R1IW1.par, R1IW1.par, p.off1, 1, 1, 1, 0)
  
            gm.SLC_intf(R1IW1.dat, R2IW1.dat, R1IW1.par, R1IW1.par, p.off1,
                        p.int1, 1, 1, 0, None, 0, 0)
            
            gm.create_offset(R1IW2.par, R1IW2.par, p.off2, 1, 1, 1, 0)
  
            gm.SLC_intf(R1IW2.dat, R2IW2.dat, R1IW2.par, R1IW2.par, p.off2,
                        p.int2, 1, 1, 0, None, 0, 0)

            # calculate the single look double difference interferogram for the burst overlap region i
            # insar phase of earlier burst is subtracted from interferogram of later burst

            gm.create_diff_par(p.off1, p.off2, p.diff_par, 0, 0)
            gm.cpx_to_real(p.int1, "tmp", range_samples, 4)
            gm.sub_phase(p.int2, "tmp", p.diff_par, p.diff, 1, 0)

            # multi-look the double difference interferogram
            # (200 range x 4 azimuth looks)
            gm.multi_cpx(p.diff, p.off1, p.diff20, p.off20, 200, 4)
            
            range_samples20 = p.getint("off20", "interferogram_width")
            azimuth_lines20 = p.getint("off20", "interferogram_azimuth_lines")
            range_samples20_half = range_samples20 / 2.0
            azimuth_lines20_half = azimuth_lines20 / 2.0
            log.info("range_samples20_half: %f" % range_samples20_half)
            log.info("azimuth_samples20_half: %f" % azimuth_samples20_half)
            log.info("to display double difference interferogram use:    "
                     "dismph %s %d " % (p.diff20, range_samples20))
   
            # determine coherence and coherence mask based on unfiltered 
            # double differential interferogram
            
            gm.cc_wave(p.diff20, None, None, p.cc20, range_samples20, 5, 5, 0)
            gm.rascc_mask(p.cc, None, range_samples20, 1, 1, 0, 1, 1,
                          cc_thresh, None, 0.0, 1.0, 1.0, 0.35, 1, p.cc_ras)
            
            # adf filtering of double differential interferogram
            
            gm.adf(p.diff20, p.adf, p.adf_cc, range_samples20, 0.4, 16, 7, 2)
            p.rm("adf_cc")

            # unwrapping of filtered phase considering coherence and mask 
            # determined from unfiltered double differential interferogram
            
            gm.mcf(p.adf, p.cc, p.cc_ras, p.phase, range_samples20, 1, 0, 0,
                   None, None, 1, 1, 512, range_samples20_half, azimuth_lines20_half)

            size = pth.getsize(p.phase)
            log.info("unwrapped phase %s  file size: %f" % size)
            
            if cleaning:
                gb.raster(p.diff20, rng=range_samples20, azi=azimuth_lines20)
                gb.raster(p.adf, rng=range_samples20, azi=azimuth_lines20)
                
                if p.exist("phase") and size > 0:
                    gm.rasrmg(p.phase, "None", range_samples20, 1, 1, 0, 1, 1,
                              0.333, 1.0, 0.35, 0.0, 1,
                              "%s.%s" % (p.phase, gm_ras))
  
            # determine overlap phase average (in radian), 
            # standard deviation (in radian), and valid data fraction
            
            if p.exist(p.cc):
                cc_stat = p.stat("cc", range_samples20)
            else:
                cc_stat = Params({
                    "mean": 0.0,
                    "stdev": 0.0,
                    "fraction_valid": 0.0
                })

            log.info("cc_fraction1000: %s" 
                     %  cc_stat.getfloat("fraction_valid") * 1e3)
            
            if p.exist("phase") and size > 0:
                ph_stat = p.stat("phase", range_samples20)
            else:
                ph_stat = Params({
                    "mean": 0.0,
                    "stdev": 0.0,
                    "fraction_valid": 0.0
                })
  
            # determine fraction10000 and stdev10000 to be used 
            # for integer comparisons
            
            cc_frac1000 = cc_stat.getfloat("fraction_valid") * 1e3
            frac, std = ph_stat.gefloat("fraction_valid"), ph_stat.gefloat("stdev")
            
            if cc_frac1000 == 0.0:
                frac10000 = 0.0
            else:
                frac10000 = frac0 * 10000 / cc_stat.gefloat("fraction_valid")
            
            std10000 = std * 10000
            
            # only for overlap regions with a significant area with high coherence
            # and phase standard deviation < stdev10000_thresh
    
            if frac10000 > frac10000_thresh and std10000 < std10000_thresh:
                # +0.1 to limit maximum weights for very low stdev
                weight = frac / (std + 0.1) / (frac + 0.1)
                Sum += mean * frac
                samples += 1
                sum_weight += frac
                sum_all += mean * frac
                samples_all +=  1
                sum_weight_all += frac
            else:
                weigth = 0.0
            
            frac1000 = frac * 1e3
            
            if frac1000 > 0:
                info = "IW%d %d %f %f %f (%f %f %f)" % (ii, jj,
                        ph_stat.getfloat("mean"), std, frac,
                        cc_stat.getfloat("mean"), cc_stat.getfloat("stdev"),
                        cc_stat.getfloat("fraction_valid"))
            else:
                info = "IW%d %d 0.0 0.0 0.0 (%f %f %f)" % (ii, jj,
                        cc_stat.getfloat("mean"), cc_stat.getfloat("stdev"),
                        cc_stat.getfloat("fraction_valid"))
            
            log.info(info)
            
            with open(res, "a") as f:
                f.write(info + "\n")
            
            jj += 1
            
        # end while
    
    if samples > 0:
        average = Sum / sum_weigth
    else:
        average = 0.0
    
    with open("res", "a") as f:
        f.write("IW%d %f\n" % (ii, average))

    
def check_ionoshpere(self, rng_win=256, azi_win=256, thresh=0.1,
                     rng_step=None, azi_step=None):
    log.info("Checking ionosphere for : %s" % self.datfile)
    
    SLC = self.datfile
    par = self.parfile
    
    if rng_step is None:
        rng_step = int(0.25 * rng_win)

    if azi_step is None:
        azi_step = int(0.25 * azi_win)
    
    log.info("rng_win, azi_win: %d, %d" % (rng_win, azi_win))
    log.info("rng_step, azi_step: %d, %d" % (rng_step, azi_step))
    
    prf = float(self["prf"].split()[0])
    dc = float(self["doppler_polynomial"].split()[0])
    az_bw = float(self["azimuth_proc_bandwidth"].split()[0])
    
    image_format = self["image_format"]
    
    if image_format == "SCOMPLEX":
        data_type = "1"
    elif image_format == "FCOMPLEX":
        data_type = "0"
    else:
        raise ValueError("image format (%s) not supported."
                          % image_format)
    
    center_frequency = dc / prf
    az_bw_frac = az_bw / prf
    
    fa1 = center_frequency - 0.25 * az_bw_frac
    fa2 = center_frequency + 0.25 * az_bw_frac
    
    df1 = 0.5 * az_bw_frac
    df2 = df1
    
    log.info("image_format: %s" % image_format)
    log.info("center_frequency: %.3f" % center_frequency)
    log.info("az_bw: %.3f, prf: %.3f,  az_bw_frac: %.3f"
              % (az_bw, prf, az_bw_frac))
    
    log.info("bpf azimuth filter parameters for lower band:    %.3f %.3f"
                % (fa1, df1))
    log.info("bpf azimuth filter parameters for higher band:   %.3f %.3f"
                % (fa2, df2))

    log.info("Start ionosphere analysis")
    
    slca, slcb, off, offs, ccp = "slc.a", "slc.b", "a_b.off", "a_b.offs",\
                                 "a_b.ccp"
    
    tmps = TMPFiles(slca, slcb, off, offs, ccp)
    
    thresh = 0.1
    
    width = self.rng()
    nlines = self.azi()
    
    if not pth.isfile(str(slca)):
        gm.bpf(SLC, slca, width, 0.0, 1.0, fa1, df1, 0, 0, None, None,
               data_type, 0, 1.0, 128)
        
    if not pth.isfile(str(slcb)):
        gm.bpf(SLC, slcb, width, 0.0, 1.0, fa2, df2, 0, 0, None, None,
               data_type, 0, 1.0, 128)
    
    gm.create_offset(par, par, off, 1, 20, 60, 0)
    
    gm.offset_pwr_tracking(slca, slcb, par, par, off, offs, ccp,
                           rng_win, azi_win, None, 2, thresh,
                           rng_step, azi_step, None, None, None, None, 5)
    
    offset_out = gm.offset_fit(offs, ccp, off, "coffs", "coffsets", thresh, 3)
    off_set_fit = tmps(off + "set_fit.out")
    
    with open(off_set_fit, "wb") as f:
        f.write(offset_out)
    
    off_width = get_par("offset_estimation_range_samples", off)
    
    offs_azi = tmps(offs + ".azimuth")
    offs_rng = tmps(offs + ".range")

    offs_azi_bmp = tmps(offs + ".azimuth.bmp")
    offs_rng_bmp = tmps(offs + ".range.bmp")

    gm.cpx_to_real(offs, offs_azi, off_width, 1)
    gm.cpx_to_real(offs, offs_rng, off_width, 1)
    
    offs_azi_tmp = tmps(offs + ".azimuth.tmp.bmp")
    offs_rng_tmp = tmps(offs + ".range.tmp.bmp")
        
    gm.ras8_float(offs_azi, None, off_width,
                  offs_azi_tmp, 1, 0.0, 240.0, None, None, None,
                  0, -2.0, 2.0, 0, 0, 1.0, 0.35, 1, 1, 0, 1, 1, 1)

    gm.ras8_float(offs_rng, None, off_width,
                  offs_rng_tmp, 1, 0.0, 240.0, None, None, None,
                  0, -2.0, 2.0, 0, 0, 1.0, 0.35, 1, 1, 0, 1, 1, 1)
    
    ccp_bmp = tmps(ccp + ".bmp")
    
    gm.raspwr(ccp, off_width, 1, 0, 1, 1, 1.0, 0.35, 1, ccp_bmp)
    
    gm.comb_hsi(offs_azi_tmp, ccp_bmp, offs_azi_bmp)
    gm.comb_hsi(offs_rng_tmp, ccp_bmp, offs_rng_bmp)
        
    #xv $offs.azimuth.bmp $offs.range.bmp &
    
    threshf = tmps("thresh_file")
    with open(threshf, "w") as f:
        f.write("Parameter              Min     Max\n")
        f.write("range_pixel_offset:      -0.2    0.2\n")
        f.write("azimuth_pixel_offset:    -2.0    2.0\n")
        f.write("estimation_ccp:           0.15   1.0\n")
        f.write("relative_range_offset:   -0.15   0.15\n")
        f.write("relative_azimuth_offset: -0.3    0.3\n")
    
    r = 7


    if 2 * azi_step <= azi_win and 4 * rng_step <= rng_win:
        offs_cond = tmps(offs + ".cond")
        
        gm.condition_offset_estimates(offs, ccp, off, threshf, offs_cond,
                                      r, 1)
        
        offs_cond_azi = tmps(offs + ".cond.azimuth")
        offs_cond_rng = tmps(offs + ".cond.range")
        
        gm.cpx_to_real(offs_cond, offs_cond_azi, off_width, 1)
        gm.cpx_to_real(offs_cond, offs_cond_rng, off_width, 0)
        
        offs_cond_azi_bmp = tmps(offs + ".cond.azimuth.bmp")
        offs_cond_rng_bmp = tmps(offs + ".cond.range.bmp")

        gm.ras8_float(offs_cond_azi, None, off_width, offs_azi_tmp,
                      1, 0.0, 240.0, None, None, None, 0, -2.0, 2.0, 0,
                      0, 1.0, 0.35, 1, 1, 0, 1, 1, 1)
        
        gm.ras8_float(offs_cond_rng, None, off_width, offs_rng_tmp,
                      1, 0.0, 240.0, None, None, None, 0, -2.0, 2.0, 0,
                      0, 1.0, 0.35, 1, 1, 0, 1, 1, 1)
        
        gm.raspwr(ccp, off_width, 1, 0, 1, 1, 1.0, 0.35, 1, ccp_bmp)
        
        gm.comb_hsi(offs_azi_tmp, ccp_bmp, offs_cond_azi_bmp)
        gm.comb_hsi(offs_rng_tmp, ccp_bmp, offs_cond_rng_bmp)
        
        #xv $offs.cond.azimuth.bmp $offs.cond.range.bmp &
        
        offs_int     = tmps(offs + ".cond.interp")
        offs_int_azi = tmps(offs + ".cond.interp.azimuth")
        offs_int_rng = tmps(offs + ".cond.interp.range")

        offs_int_azi_bmp = tmps(offs + ".cond.interp.azimuth.bmp")
        offs_int_rng_bmp = tmps(offs + ".cond.interp.range.bmp")
        
        gm.interp_ad(offs_cond, offs_int, off_width, r, 35, 50, 2, 0, 0)
        
        gm.cpx_to_real(offs_int, offs_int_azi, off_width, 1)
        gm.cpx_to_real(offs_int, offs_int_rng, off_width, 0)
        
        gm.ras8_float(offs_int_azi, None, off_width, offs_azi_tmp,
                      1, 0.0, 240.0, None, None, None, 0, -2.0, 2.0, 0,
                      0, 1.0, 0.35, 1, 1, 0, 1, 1, 1)

        gm.ras8_float(offs_int_rng, None, off_width, offs_rng_tmp,
                      1, 0.0, 240.0, None, None, None, 0, -2.0, 2.0, 0,
                      0, 1.0, 0.35, 1, 1, 0, 1, 1, 1)
        
        gm.raspwr(ccp, off_width, 1, 0, 1, 1, 1.0, 0.35, 1, ccp_bmp)
        
        gm.comb_hsi(offs_azi_tmp, ccp_bmp, offs_int_azi_bmp)
        gm.comb_hsi(offs_rng_tmp, ccp_bmp, offs_int_rng_bmp)
        
        #xv $offs.cond.interp.azimuth.bmp $offs.cond.interp.range.bmp &
    else:
        log.warning("no conditionning and interpolation done for range "
                    "steps larger than 0.25*window size and azimuth steps "
                    "larger than 0.5 window size")
    
    remove("a_b.offs*", "coffs", "coffsets")
    log.info("*** End Ionosphere Analysis")


def prasdt_pwr24(plist, pmask, SLC_par, pdata, par_out, mli, cycle,
                 radius=4.0, rec_num=None, outdir=".", search=3, imode=3):
    
    width = gb.Files.get_par("range_samples", par_out)
    
    tmp = "tmp.prasdt_pwr24"
    gb.rm(tmp)
    
    npt = int([line.split(":")[1] for line in gm.npt(plist).decode().split("\n")
               if "total_number_of_points:" in line][0])
    
    nbytes = pth.getsize(pdata)
    
    nrec = nbytes / npt / 4
    
    base = pth.basename(pdata)
    
    if rec_num is None:
        rec = 0
        while rec != nrec:
            rec += 1
            
            if rec < 10:
                fras = "%s.0%d.bmp" % (base, rec)
            else:
                fras = "%s.%d.bmp" % (base, rec)
            
            gm.pt2data(plist, pmask, SLC_par, pdata, rec, tmp, par_out, 2,
                       imode, radius, search)

            print("data record: %d" % rec)
            
            gm.rasdt_pwr24(tmp, mli, width, 1, 1, 0, 1, 1, cycle, 1.0, 0.4, 1,
                           pth.join(outdir, fras))
    else:
        print("data record: %d" % rec_num)
        fras = "%s.bmp" % base

        gm.pt2data(plist, pmask, SLC_par, pdata, rec_num, tmp, par_out, 2,
                   imode, radius, search)

        gm.rasdt_pwr24(tmp, mli, width, 1, 1, 0, 1, 1, cycle, 1.0, 0.4, 1,
                       pth.join(outdir, fras))


    gb.rm(tmp)
