import gamma as gp

from logging import getLogger

log = getLogger("gamma.interferometry")

gp = gp.gamma_progs


__all__ = ("coherence", "base_plot")


cc_weights = {
    "constant": 0,
    "gaussian": 1,
}



def base_plot(midx, RSLCs, bperp_lims=(0.0, 150.0),
              delta_T_lims=(0.0, 15.0), SLC_tab="SLC_tab",
              bperp="bperp", itab="itab"):

    with open(SLC_tab, "w") as f:
        f.write("%s\n" % "\n".join(str(rslc) for rslc in RSLCs))
    
    mslc_par = RSLCs[midx].par
    
    gp.base_calc(SLC_tab, mslc_par, bperp, itab, 1, 1, bperp_lims[0],
                 bperp_lims[1], delta_T_lims[0], delta_T_lims[1])
    
    gp.base_plot(SLC_tab, mslc_par, itab, bperp, 1)


def coherence(ifg, cc, slope_win=5, weight_type="gaussian", corr_thresh=0.4,
              box_lims=(3.0,9.0)):
    wgt_flag = cc_weights[weight_type]
    
    #log.info("CALCULATING COHERENCE AND CREATING QUICKLOOK IMAGES.")
    #log.info('Weight type is "%s"'.format(weight_type))
    
    width = ifg.rng()
    
    log.info("Estimating phase slope.", end=" ")
    gp.phase_slope(ifg.dat, slope, width, slope_win, corr_thresh)

    log.info("Calculating coherence.", end=" ")
    gp.cc_ad(ifg.dat, mli1, mli2, slope, None, ifg.cc, width, box_lims[0],
             box_lims[1], wgt_flag)

    log.info("Creating quicklook bmp.", end=" ")
    ifg.rascc()
    log.info("DONE.")
