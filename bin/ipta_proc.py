#!/usr/bin/env python

import os.path as pth

from glob import iglob
from collections import namedtuple

import gamma as gm

gp = gm.gamma_progs
mkdir = gm.mkdir


RDC = namedtuple("RDC", ["rng", "azi"])


def calc_rdc(latlon, mpar, dem_par, hgt, diff_par):
    out = gp.coord_to_sarpix(mpar, None, dem_par, latlon[0], latlon[1],
                              hgt, diff_par).decode()
    
    for line in out.split("\n"):
        if "corrected SLC/MLI range, azimuth pixel (int)" in line:
            split = line.split(":")[1]
            return RDC(int(split.split()[0]), int(split.split()[1]))


def calc_off_num(one, two):
    if one < two:
        low, high = one, two
    else:
        low, high = two, one
    
    return low, high - low
    
    
def offs_lines(latlons, mpar, dem_par, hgt, diff_par):
    rdc = [calc_rdc(latlon, mpar, dem_par, hgt, diff_par)
           for latlon in latlons]

    return [*calc_off_num(rdc[0].rng, rdc[1].rng),
            *calc_off_num(rdc[0].azi, rdc[1].azi)]


def date(path):
    return pth.basename(path).split(".")[0]


def main():
    root = "/mnt/storage_A/istvan/dszekcso"
    
    ipta_dir = mkdir(pth.join(root, "ipta"))
    rslc_dir = mkdir(pth.join(ipta_dir, "rslc"))
    rmli_dir = mkdir(pth.join(ipta_dir, "rmli"))
    int_dir  = mkdir(pth.join(ipta_dir, "int"))
    sp_dir   = mkdir(pth.join(ipta_dir, "sp_msr"))
    pt_dir   = mkdir(pth.join(ipta_dir, "pt_combined"))
    
    
    rng_looks, azi_looks = 15, 3
    
    geo_dir = pth.join(root, "geo")
    
    dem_par = pth.join(root, "dem", "dem_seg.dem_par")
    hgt = pth.join(geo_dir, "dem.rdc")
    
    diff_par = pth.join(geo_dir, "diff_par")

    master_date = "20161205"
    master = gm.SLC(datfile=pth.join("deramp", master_date + ".slc.deramp"))
    
    latlons = [(46.00, 18.65), (46.15, 18.85)]
    
    off_ln = offs_lines(latlons, master.par, dem_par, hgt, diff_par)
    

    # Test cropping
    master_crop = gm.SLC(datfile=pth.join(rslc_dir, "%s.rslc" % master_date))
    

    if 0:
        master.copy(master_crop, rng_off=off_ln[0], rng_num=off_ln[1],
                    azi_off=off_ln[2], azi_num=off_ln[3])
        master_crop.raster()
    
    
    RSLC = tuple(gm.SLC(datfile=slc)
                 for slc in iglob(pth.join("deramp", "*.slc.deramp")))
    
    rslc_tab = pth.join(ipta_dir, "rslc.tab")
    
    
    
    cropped = [gm.SLC(datfile=pth.join(rslc_dir, "%s.rslc" % date(slc.dat)))
               for slc in RSLC]
    
    
    # Crop SLCs
    if 0:
        for rslc, crop in zip(RSLC, cropped):
            rslc.copy(crop, rng_off=off_ln[0], rng_num=off_ln[1],
                      azi_off=off_ln[2], azi_num=off_ln[3])
    
    # Master idx
    midx = [ii for ii, slc in enumerate(cropped)
            if date(slc.dat) == master_date][0]
    
    if 0:
        with open(rslc_tab, "w") as f:
            f.write("%s\n" % "\n".join(str(rslc) for rslc in cropped))

    
    # Cropped rasters for checking
    if 0:
        for rslc in cropped:
            rslc.raster(mode="mph")
    

    
    rmli = [gm.MLI(datfile=pth.join(rmli_dir, "%s.rmli" % date(slc.dat)))
            for slc in RSLC]

    mmli = [mli for mli in rmli if date(mli.dat) == master_date][0]
    mslc = [slc for slc in cropped if date(slc.dat) == master_date][0]

    
    rmli_tab = pth.join(ipta_dir, "rmli.tab")
    avg = gm.MLI(datfile=pth.join(ipta_dir, "avg.rmli"), parfile=mmli.par)


    if 0:
        with open(rmli_tab, "w") as f:
            for rslc, _rmli in zip(cropped, rmli):
                rslc.multi_look(_rmli, rng_looks=1, azi_looks=1)
                
                f.write("%s\n" % _rmli.dat)
        
        gp.ave_image(rmli_tab, avg.rng(), avg.dat)
        avg.raster()
    

    bperp, itab = pth.join(ipta_dir, "bperp.txt"), pth.join(ipta_dir, "itab")
    

    # cropped HGT, DEM and diff_par

    dem_dir = pth.join(root, "dem")
    
    diff_par_crop = pth.join(root, "geo", "diff_par_crop")
    
    dem_orig = gm.MLI(datfile=pth.join(dem_dir, "srtm.dem"),
                      parfile=pth.join(dem_dir, "srtm.dem_par"))
    
    dem_crop = gm.MLI(datfile=pth.join(dem_dir, "dem_seg_crop.dem"),
                      parfile=pth.join(dem_dir, "dem_seg_crop.dem_par"))


    hgt_crop = gm.HGT(pth.join(geo_dir, "dem_crop.rdc"), mmli)
    offset_crop = pth.join(geo_dir, "offset_crop")
    

    if 0:
        gp.create_diff_par(mmli.par, None, diff_par_crop, 1, 0)
        
        tmp = gm.Files.get_tmp()
        
        ll_ovs = (3.0, 3.0)
        
        gp.gc_map(mmli.par, None, dem_orig.par, dem_orig.dat,
                   dem_crop.par, dem_crop.dat, tmp, ll_ovs[0], ll_ovs[1])
        
        gm.rm(tmp)
        
        offset = "tmp"
        
        gp.create_offset(master.par, master.par, offset, 1, 1, 1, 0)
        
        gp.multi_real(hgt, offset, hgt_crop, offset_crop, 1, 1,
                      off_ln[2], off_ln[3], off_ln[0], off_ln[1])

        hgt_crop.raster()
        
        gm.rm(offset)
    
    
    avg3 = gm.MLI(datfile=avg.dat + "3")
    hgt3 = gm.HGT(hgt_crop.dat + "3", avg3)
    
    
    offset = pth.join(geo_dir, "offset")
    offset3 = offset + "3"


    if 0:
        gp.create_offset(mmli.par, mmli.par, offset, 1, 1, 1, 0)    
        
        gp.multi_real(hgt_crop, offset, hgt3, offset3, rng_looks, azi_looks)
        gp.multi_real(avg.dat, offset, avg3.dat, offset3, rng_looks, azi_looks)
        
        rng = gm.Files.get_par("interferogram_width", offset3)
        azi = gm.Files.get_par("interferogram_azimuth_lines", offset3)
        
        avg3.raster(rng=rng, azi=azi, mode="pwr", image_format="FLOAT")
        hgt3.raster(rng=rng, azi=azi, mode="hgt", image_format="FLOAT")
    
    
    ifg_tab = pth.join(ipta_dir, "ifg.tab")
    
    ml_rng, ml_azi = gm.Files.get_par("interferogram_width", offset3), \
                     gm.Files.get_par("interferogram_azimuth_lines", offset3)

    if 0:
        gp.base_calc(rslc_tab, master.par, bperp, itab, 1, 1,
                     1.0, 450.0, 1.0, 15.0, None)
    
    if 0:
        with open(bperp, "r") as bp, open(ifg_tab, "w") as f:
            for line in bp:
                split = line.split()
                
                date1 = split[1].strip()
                date2 = split[2].strip()
                
                ifg = gm.IFG(pth.join(int_dir, "%s_%s.diff" % (date1, date2)))
                
                slc1 = gm.SLC(datfile=pth.join(rslc_dir, "%s.rslc" % date1))
                slc2 = gm.SLC(datfile=pth.join(rslc_dir, "%s.rslc" % date2))
                
                gp.create_offset(slc1.par, slc2.par, ifg.par, 1, rng_looks,
                                 azi_looks, 0)
                
                gp.phase_sim_orb(slc1.par, slc2.par, ifg.par, hgt3, 
                                 ifg.sim_unw, mslc.par, None, None, 1, 1)
                
                gp.SLC_diff_intf(slc1.dat, slc2.dat, slc1.par, slc2.par,
                                 ifg.par, ifg.sim_unw, ifg.dat,
                                 rng_looks, azi_looks, 0, 0)
                
                
                ifg.raster(mli=avg3)
                
                f.write("%s\n" % ifg)
    
    
    
    rmli3 = gm.MLI(datfile=mmli.dat + "3")
    
    multi = gm.PointData(mslc, avg, root=gm.mkdir(pth.join(ipta_dir, "pt_ml")))
    
    multi["diff"] = "fcomplex"
    multi["hgt"] = "float"
    
    
    if 0:
        # mkgrid pt3 133 200 15 3 7 1
        gp.mkgrid(multi.plist, ml_rng, ml_azi, rng_looks, azi_looks, 0, 0)
        gp.multi_look(mslc, rmli3, rng_looks, azi_looks, 0, None, 0.000001)
        
        with gm.ListIter(ifg_tab, gm.IFG.from_line) as f:
            for ii, ifg in enumerate(f):
                gp.data2pt(ifg.dat, rmli3.par, multi.plist, multi.slc.par,
                            multi.data["diff"], ii + 1, 0)
        
        multi.raster("diff", radius=5.0)
    
    if 0:
        multi.get_data(hgt3, rmli3.par, "hgt", rec=1, dtype="float")
        # multi.ras_data("hgt", pth.join(ipta_dir, "avg.rmli3.bmp"), rec=(1,None))
        # rewrite gamma script
        multi.raster("hgt", cycle=45.0, radius=5.0)
    
    avg_ras = avg.dat + ".bmp"
    
    rng_looks, azi_looks = 4, 4
    pwr_min, cc_min, msr_min = 0.0, 0.4, 1.0
    

    sp_msr = gm.PointData(mslc, gm.MLI(datfile=avg.dat, parfile=mmli.par),
                          root=gm.mkdir(pth.join(ipta_dir, "pt_ps")))
    
    sp_msr.map_ras, sp_msr.sp_ras = \
    sp_msr.rjoin("ptmap.bmp"), sp_msr.rjoin("sp.bmp")
    
    
    if 0:
        mslc.multi_look(mmli, rng_looks=1, azi_looks=1, scale=0.000001)
        
        gp.ras_dB(avg.dat, avg.rng(), 1, 0, 1, 1, -22.0, 3.5, 0.0, 1, avg_ras)

        sp_msr.sp_all(rslc_tab)

    rng = mslc.rng()
    
    sp_msr.cc_avg  = gm.MLI(datfile=sp_msr.rjoin("sp_msr", "avg.sp_cc"),
                            parfile=mslc.par)
    sp_msr.msr_avg = gm.MLI(datfile=sp_msr.rjoin("sp_msr", "avg.sp_msr"),
                            parfile=mslc.par)
    
    
    if 0:
        gp.ave_image(sp_msr.rjoin("sp_msr", "cc_list"), rng, sp_msr.cc_avg.dat)
        # sp_msr.cc_avg.raster(mode="_linear")
        
        gp.ave_image(sp_msr.rjoin("sp_msr", "msr_list"), rng, sp_msr.msr_avg.dat)
        # sp_msr.msr_avg.raster(mode="_linear")

    
    if 0:
        cc = sp_msr.stat("cc_avg")
        msr = sp_msr.stat("msr_avg")

        print("\tMean\t+-\tstd\nCC:\t%1.4g\t+-\t%1.2g\nMSR:\t%1.4g\t+-\t%1.2g" % 
              (cc.getfloat("mean"), cc.getfloat("stdev"),
               msr.getfloat("mean"), msr.getfloat("stdev")))
    
    if 0:
        gp.single_class_mapping(2,
                                sp_msr.cc_avg.dat, 0.4, 1.0,
                                sp_msr.msr_avg.dat, 0.5, 100.0,
                                sp_msr.map_ras, rng, None, None, 1, 1)

        # rng_looks=1, azi_looks=1, dtype=6 (raster)
        gp.image2pt(sp_msr.map_ras, rng, sp_msr.plist, 1, 1, 6)


    if 0:
        with open(pth.join(ipta_dir, "D_20160928.stn"), "r") as station:
             for line in station:
                if not line.startswith("date_&_time(hour)"):
                    refl = gm.get_llh(line)
                    rdc = calc_rdc([refl["lat"], refl["lon"]], mmli.par,
                                   dem_crop.par, hgt_crop, diff_par_crop)
                    
                    sp_msr.add_point(rdc.rng, rdc.azi)

    if 0:
        print("Selected number of PS points: ", sp_msr.npoints())
        
        # rng_looks=1, azi_looks=1, rgm=(255,255,0), cross_size=15
        gp.ras_pt(sp_msr.plist, None, avg_ras, sp_msr.sp_ras,
                   1, 1, 255, 255, 0, 10)


    sp_msr["slc"] = "fcomplex"
    sp_msr["base"] = "Unknown"
    sp_msr["sim_unw"] = "float"
    sp_msr["int"] = "fcomplex"
    sp_msr["diff"] = "fcomplex"
    sp_msr["hgt"] = "float"
    

    if 0:
        sp_msr.get_data(hgt_crop, mmli.par, "hgt", rec=1)
        sp_msr.raster("hgt", cycle=50.0, radius=5.0)
    
    sp_msr.par = sp_msr.rjoin("SLC_par")


    if 0:
        sp_msr.SLC2pt(rslc_tab)
    
    if 0:
        gp.base_orbit_pt(sp_msr.par, itab, None, sp_msr.data["base"])

        gp.intf_pt(sp_msr.plist, None, itab, None,
                   sp_msr.data["slc"], sp_msr.data["diff"],
                   sp_msr.data["slc"].dcode(), sp_msr.par)
        
        gp.phase_sim_orb_pt(sp_msr.plist, None, sp_msr.par, None,
                            itab, None, sp_msr.data["hgt"],
                            sp_msr.data["sim_unw"], sp_msr.slc.par)
        
        gp.intf_pt(sp_msr.plist, None, itab, None,
                    sp_msr.data["slc"], sp_msr.data["int"], 0)


        gp.sub_phase_pt(sp_msr.plist, None,
                        sp_msr.data["int"], None,
                        sp_msr.data["sim_unw"], sp_msr.data["diff"], 1, 0)
        
    if 0:
        sp_msr.raster("diff", radius=5.0)

    
    pt = gm.PointData(sp_msr.slc, sp_msr.mli, root=pt_dir)


    if 0:
        gm.cat(pt.plist, sp_msr.plist, multi.plist)
    
        print("Combined number of points: %d" % pt.npoints())
    
    pt["hgt_single"] = "float"
    pt["hgt_multi"] = "float"
    pt["hgt"] = "float"
    
    
    if 1:
        tmp_hgt, multi_hgt = "tmp_hgt", "tmp_multi_hgt"
        
        # tmp_hgt values 
        gp.lin_comb_pt(multi.plist, None, multi.data["hgt"].path, 1,
                                           multi.data["hgt"].path, 1,
                                           multi_hgt, 1, 1000000, 0.0, 0.0, 2, 1)
        
        gm.cat(tmp_hgt, sp_msr.data["hgt"].path, multi_hgt)
        
        gp.thres_msk_pt(pt.plist, pt.mask["single"], tmp_hgt, 1, -500.0, 10000.0)
        gp.thres_msk_pt(pt.plist, pt.mask["multi"], tmp_hgt, 1, 999999.0, 1000001.0)
        
        gm.rm(tmp_hgt, multi_hgt)
    

    if 0:
        pt.ras_pt(avg_ras, "pt_single.bmp", mask="single", size=3.0)
        pt.ras_pt(avg_ras, "pt_multi.bmp", mask="multi", size=3.0)
    
    
    if 0:
        gp.expand_data_pt(sp_msr.plist, None, sp_msr.slc.par, sp_msr.data["hgt"],
                           pt.plist, pt.mask["single"], pt.data["hgt_single"],
                           1, 2, 1)
        
        gp.expand_data_pt(multi.plist, None, multi.slc.par, multi.data["hgt"],
                           pt.plist, pt.mask["multi"], pt.data["hgt_multi"],
                           1, 2, 1)
        
        gp.lin_comb_pt(pt.plist, None,
                        pt.data["hgt_single"], 1,
                        pt.data["hgt_multi"], 1,
                        pt.data["hgt"], 1,
                        0.0, 1.0, 1.0, 2, 1)

    if 0:
        pt.raster("hgt", cycle=45.0)


    pt["diff_single"] = "fcomplex"
    pt["diff_multi"] = "fcomplex"
    pt["diff"] = "fcomplex"
    
    
    if 0:
        tmp = gm.Base(gm.Files.get_tmp(),
                      s_real=".s_real", m_real=".m_real",
                      s_imag=".s_imag", m_imag=".m_imag",
                      real=".real", imag=".imag", keep=False)


        gp.expand_data_pt(sp_msr.plist, None, sp_msr.slc.par, sp_msr.data["diff"],
                           pt.plist, pt.mask["single"], pt.data["diff_single"],
                           None, 0, 1)
        
        gp.expand_data_pt(multi.plist, None, multi.slc.par, multi.data["diff"],
                           pt.plist, pt.mask["multi"], pt.data["diff_multi"],
                           None, 0, 1)
        

        # separate into real and imageinary part and combine using lin_comb_pt
        # and combine again into cpx values
        
        gp.cpx_to_real(pt.data["diff_single"], tmp.s_real, 1, 0)
        gp.cpx_to_real(pt.data["diff_single"], tmp.s_imag, 1, 1)

        gp.cpx_to_real(pt.data["diff_multi"], tmp.m_real, 1, 0)
        gp.cpx_to_real(pt.data["diff_multi"], tmp.m_imag, 1, 1)
        
        gp.lin_comb_pt(pt.plist, None,
                        tmp.s_real, None,
                        tmp.m_real, None,
                        tmp.real, None,
                        0.0, 1.0, 1.0, 2, 1)

        gp.lin_comb_pt(pt.plist, None,
                        tmp.s_imag, None,
                        tmp.m_imag, None,
                        tmp.imag, None,
                        0.0, 1.0, 1.0, 2, 1)
        
        gp.real_to_cpx(tmp.real, tmp.imag, pt.data["diff"], 1, 0)
    
    
    if 0:
        pt.raster("diff", radius=5.0)
    
    
    if 0:
        pt.raster("hgt", cycle=45.0, dis=1)
    

    if 0:
        refl = gm.get_llh("IB1  REGI_VIZMU    46-05-18.91302  "
                          "18-45-41.60132  185.6680  1")
        rdc = calc_rdc([refl["lat"], refl["lon"]], mmli.par,
                       dem_crop.par, hgt_crop, diff_par_crop)

        srng, sazi, pmax = 25, 25, 30
        
        gp.prox_prt(pt.plist, pt.mask["single"], pt.data["hgt"],
                     2399, 577, srng, sazi, pmax, 2,
                     pth.join(ipta_dir, "coords.txt"), 1)
    

    # Reference point nr. (in pt_combined)
    nref = 3235

    # Reference point: nr. 11878 (in pt_combined) - Athens processing

    
    pt.par = pt.rjoin("SLC_par")
    
    if 0:
        pt.SLC2pt(rslc_tab)

    pt.base = sp_msr.data["base"]
    pt["diff_rfilt"] = pt.data["diff"].dtype
    
    
    if 0:
        # replace reference point values with spatially filtered 
        # values (to reduce noise)
        gp.spf_pt(pt.plist, None, pt.slc.par, pt.data["diff"],
                   pt.data["diff_rfilt"], None, 0, 25, 0, None, nref, 0)

    
    pt["mcf_unw"] = "float"


    if 0:
        gp.mcf_pt(pt.plist, None, pt.data["diff"], None, None, None,
                  pt.data["mcf_unw"], 1.0, 1.0, nref, 1)
    
    if 0:
        pt.def_pt("diff_rfilt", itab, "const_bperp", nref=nref,
                  log="def_pt.log", def_min=-0.5, def_max=0.5,
                  multi=True, rpatch=50)
    
    
    pt["unw"] = "float"
    pt["sigma"] = "float"
    pt["def"] = "float"
    pt["res"] = "float"
    pt["dh"] = "float"
    
    
    
    if 0:
        pt.raster("sigma")
        pt.raster("res")
        pt.raster("dh")
    
    pt["res_cpx"] = "fcomplex"
    pt["res_unw"] = "float"
    pt["atm1"] = "float"
    pt["res_cpx_spf"] = "fcomplex"
    
    
    if 0:
        gp.unw_to_cpx_pt(pt.plist, pt.mask["def_mask"], pt.data["res"],
                         None, pt.data["res_cpx"])
        
        r_max = 150
        spf_type = 1 # triangular weighted average: 1 - (r/r_max)
        
        gp.fspf_pt(pt.plist, pt.mask["def_mask"], pt.slc.par,
                   pt.data["res_cpx"], pt.data["res_cpx_spf"], None,
                   pt.data["res_cpx"].dcode(), r_max, spf_type, 0)
        
        gp.mcf_pt(pt.plist, pt.mask["def_mask"], pt.data["res_cpx_spf"],
                  None, None, None, pt.data["res_unw"], 1, 1, nref, 0)

    
    if 0:
        pt.raster("res_unw", radius=8.0)
    
    
    if 0:
        tmp = pt.tmp("float")
        
        r_max, spf_type = 100, 2 # quadratic weighted average: 1 - (r/r_max)**2
        
        gp.fspf_pt(pt.plist, pt.mask["def_mask"], pt.slc.par,
                   pt.data["res_unw"], tmp, None,
                   pt.data["res_unw"].dcode(), r_max, spf_type, 0)
        
        r_max, weight, pt_num = 150, 1, 0
        # 0: 1.0
        # 1: 1.0-radius/r_max
        
        gp.expand_data_pt(pt.plist, pt.mask["def_mask"], pt.slc.par,
                          tmp, pt.plist, None, pt.data["atm1"], None,
                          tmp.dcode(), r_max, weight, pt_num)
    
    
    if 0:
        pt.raster("atm1", radius=8.0)


    pt["hgt1"] = "float"


    if 0:
        gp.lin_comb_pt(pt.plist, None,
                       pt.data["hgt"], 1,
                       pt.data["dh"], 1,
                       pt.data["hgt1"], None,
                       0.0, 1.0, 1.0, pt.data["hgt"].dcode(), 1)
    
    if 0:
        pt.raster("hgt", radius=6.0, cycle=50.0)
        pt.raster("hgt1", radius=6.0, cycle=50.0)

    
    pt["sim_unw_1"] = "float"
    pt["diff1"] = "fcomplex"
    pt["diff1_rfilt"] = "fcomplex"
    
    
    if 0:
        ph_flag, bflag = 1, 0
        # flattened, initial baseline
        
        tmp = pt.tmp(pt.data["diff"].dtype)
        
        
        gp.phase_sim_pt(pt.plist, None, pt.par, None, itab, None, pt.base,
                        pt.data["dh"], pt.data["sim_unw_1"], None, ph_flag, bflag)
        
        gp.sub_phase_pt(pt.plist, None, pt.data["diff"], None,
                        pt.data["sim_unw_1"], tmp, 1, 0)

        gp.sub_phase_pt(pt.plist, None, tmp, None,
                        pt.data["atm1"], pt.data["diff1"], 1, 0)
        
        
        gp.spf_pt(pt.plist, None, pt.slc.par, pt.data["diff1"],
                  pt.data["diff1_rfilt"], None, 0, 25, 0, None, nref, 0)
        
        pt.def_pt("diff1_rfilt", itab, "const_time", nref=nref,
                  def_min=-1.0, def_max=1.0, log="def_pt.log")
    
    
    if 0:
        pt.raster("def", radius=6.0)
        pt.raster("sigma", radius=6.0)
        pt.raster("res", radius=6.0)
    
    
    pt["disp"] = "float"
    
    
    if 0:
        gp.dispmap_pt(pt.plist, pt.mask["def_mask"], pt.par, itab,
                      pt.data["unw"], pt.data["hgt1"], pt.data["disp"], 0, 0)
    
    if 1:
        gp.vu_disp(pt.plist, pt.mask["def_mask"], pt.par, itab,
                   pt.data["disp"], pt.data["def"], pt.data["hgt1"],
                   pt.data["sigma"], None, None, None,
                   pth.join(pt.data_dir, "def.01.bmp"))
        
        # pt.raster("def", radius=6.0, dis=1)
    
# phase_sim_pt pt - pSLC_par - itab - pbase pdh1 pdh1_sim_unw - 1 0
# /bin/rm ptmp1
# sub_phase_pt pt - pdiff -  pdh1_sim_unw ptmp1 1 0
# 
# # subtract the estimated atmospheric path delays
# sub_phase_pt pt - ptmp1 -  patm1 pdiff1 1 0
# 
# # replace reference point values with spatially filtered values (to reduce noise)
# spf_pt pt - 20150826.vv.rslc.par pdiff1 pdiff1a - 0 25 0 - 11878 0
# 
# # now after subtracting the atmospheric phase we should be able to use def_mod_pt
# # with one spatial reference point
# # Again we use model 1 to only estimate a height correction and to determine an update
# # to the atmospheric phase
# def_mod_pt pt - pSLC_par - itab pbase 0 pdiff1a 1 11878 pres2 pdh2 - punw2 psigma2 pmask2 90. -0.1 0.1 1.2 1 pdh_err2 -
# 53423


        
# update heights
# lin_comb_pt pt - pdem_combined 1 pdh1 1 phgt1 - 0.0 1. 1. 2 1
# pdisdt_pwr24 pt - 20150826.vv.rslc.par phgt1 1 20150826.vv.rmli.par ave.vv.rmli 100. 3
    
    if 0:
        pt.plot_pt("test.png", term="png")
        # a = pt.data["hgt"]
        # gm.test(a.path, a.dtype, "histo.png", term="pngcairo", debug=0)
         
        #gm.histo(a.path, a.dtype, "histo.png", term="pngcairo", nbin=100)
    
    
    
    
    return 0


    pwr = gm.PointData(mslc.par,
                       gm.MLI(datfile=avg.dat, parfile=mmli.par),
                       root=gm.mkdir(pth.join(ipta_dir, "pt_pwr")))
    
    pwr.msr, pwr.ras = pwr.rjoin("MSR"), pwr.rjoin("pt.bmp")
    msr_min, pwr_min = 2.3, 1.0
    
    
    
    if 0:
        gm.pwr_stat(rslc_tab, mslc.par, pwr.msr, pwr.plist,
                    msr_min, pwr_min, None, None, None, None, 1, 2)
        
        print("Selected number of PS points: ", pwr.npoints())
        
        msr = gm.MLI(datfile=pwr.msr, parfile=mslc.par)
        msr.raster(mode="pwr")
        
        msr.report()
        
        pwr.ras_pt(avg_ras, pwr.ras, size=10)


    if 0:
        from struct import pack
        
        with open("test.bin", "wb") as f:
            f.write(pack(">ff", 1.0, 2.0))
        
        gm.output("test.png", term="png")
        gm.cmd("plot 'test.bin' binary endian=big format='%float'")
        gm.plot()


if __name__ == "__main__":
    main()
