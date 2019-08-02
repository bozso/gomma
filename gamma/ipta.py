import os.path as pth

from tempfile import NamedTemporaryFile
from glob import glob
from math import pi
from collections import namedtuple
from functools import partial
from struct import pack
from logging import getLogger

from ctypes import sizeof, c_float, c_int, c_byte, c_short

import gamma as gm


log = getLogger("gamma.ipta")

gp = gm.gp


__all__ = ("PointData", "merge", "get_llh", "plist_fmt")

plist_fmt = ">2I"

class Data(object):
    sizes = {
        "FCOMPLEX" : 2 * sizeof(c_float),
        "SCOMPLEX" : 2 * sizeof(c_short),
        "FLOAT"    : sizeof(c_float),
        "INT"      : sizeof(c_int),
        "SHORT"    : sizeof(c_short),
        "BYTE"     : sizeof(c_byte),
        "raster"   : None
    }

    dtypes = {
        "FCOMPLEX" : 0,
        "SCOMPLEX" : 1,
        "FLOAT"    : 2,
        "INT"      : 3,
        "SHORT"    : 4,
        "BYTE"     : 5,
        "raster"   : 6
    }

    def __init__(self, path, dtype, keep=True):
        self.keep, self.path, self.dtype  = keep, path, dtype
    
    
    def __del__(self):
        if not self.keep:
            gm.rm(self.path)


    def dcode(self):
        return Data.dtypes[self.dtype.upper()]
    
    
    def size(self):
        return Data.sizes[self.dtype.upper()]
    
    
    def __str__(self):
        return self.path


class Masks(object):
    def __init__(self, root):
        self.root, self.mlist = root, {}

        gm.mkdir(root)
    
    
    def __getitem__(self, key):
        if key in self.mlist:
            return self.mlist[key]
        else:
            _path = pth.join(self.root, key)
            self.mlist[key] = _path
            return _path
            

class PointData(object):
    modes = {
        "const": 0,
        "const_bperp": 1,
        "const_bperp_time": 2,
        "bperp": 3,
        "bperp_time": 4,
        "const_time": 5,
        "time": 6,
    }

    dtypes = {
        "FCOMPLEX" : 0,
        "SCOMPLEX" : 1,
        "FLOAT"    : 2,
        "INT"      : 3,
        "SHORT"    : 4,
        "BYTE"     : 5,
        "RASTER"   : 6,
        "UNKNOWN"   : -1
    }
    
    
    __save__ = {"root", "data_dir", "slc", "mli", "data", "mask"}
    
    
    def __init__(self, slc, mli, root="."):
        self.root, self.data_dir = root, pth.join(root, "data")
        
        gm.mkdir(root)
        gm.mkdir(self.data_dir)
        
        self.plist, self.slc, self.data, self.mask, self.mli = \
        pth.join(root, "plist"), slc, {}, Masks(pth.join(root, "mask")), mli
    
    
    def __setitem__(self, key, dtype):
        assert dtype.upper() in PointData.dtypes.keys(), \
               "Unrecognized dtype: %s" % dtype
        
        self.data[key] = Data(pth.join(self.data_dir, key), dtype)

    
    def __repr__(self):
        return "<PointData point_list: %r point_datas: %r point_masks: %r>" \
               % (self.plist, self.pdata, self.pmask)

    def __str__(self):
        return "<PointData point_list: %s point_datas: %s point_masks: %s>" \
               % (self.plist, self.pdata, self.pmask)
    

    def tmp(self, dtype, **kwargs):
        return Data(gm.Files.get_tmp(**kwargs), dtype, keep=False)
    
    def get_mask(self, mask):
        if mask is None:
            return None
        else:
            return self.mask[mask]


    def rjoin(self, *path):
        return pth.join(self.root, *path)

    
    def add_point(self, rng, azi):
        with open(self.plist, "ab") as f:
            f.write(pack(plist_fmt, rng, azi))
    
    
    def sp_all(self, slc_list, pwr_min=0.0, cc_min=0.4, msr_min=1.2,
               rng_looks=4, azi_looks=4, rng_ovr=1):
        
        self.sp_msr_dir = gm.mkdir(self.rjoin("sp_msr"))
        
        self.cc_list, self.msr_list = \
        pth.join(self.sp_msr_dir, "cc_list"), \
        pth.join(self.sp_msr_dir, "msr_list"),
        
        if 0:
            with gm.ListIter(slc_list, gm.SLC.from_line) as f, \
                 open(self.cc_list, "w") as f_cc, \
                 open(self.msr_list, "w") as f_msr:
                
                for ii, slc in enumerate(f):
                    tpl = pth.join(self.sp_msr_dir, slc.datestr() + ".%s")
                    
                    cc, msr, pt = tpl % "sp_cc", tpl % "sp_msr", tpl % "sp_pt"
                    
                    gp.sp_stat(slc.dat, None, cc, msr, pt, slc.rng(),
                               pwr_min, cc_min, msr_min, rng_looks, azi_looks,
                               None, None, None, None, None, None,
                               PointData.dtypes[slc.img_fmt()], rng_ovr)
            
                    f_cc.write(cc + "\n")
                    f_msr.write(msr + "\n")
            
            
            mpar = self.slc.par
            
            self.cc_avg, self.msr_avg = \
            gm.MLI(datfile=pth.join(self.sp_msr_dir, "avg.sp_cc"), parfile=mpar), \
            gm.MLI(datfile=pth.join(self.sp_msr_dir, "avg.sp_msr"), parfile=mpar)
            
            rng = self.cc_avg.rng()
            
            gp.ave_image(self.cc_list, rng, self.cc_avg.dat)
            self.cc_avg.raster(mode="_linear")
            
            if msr_min > 0.0:
                gp.ave_image(self.msr_list, rng, self.msr_avg.dat)
                self.cc_avg.raster(mode="_linear")

    
    # TODO: is mask needed?
    def SLC2pt(self, slc_tab, mask=None):
        
        with gm.ListIter(slc_tab, gm.SLC.from_line) as f:
            SLCs = tuple(slc for slc in f)
        
        fmts = tuple(slc.img_fmt() for slc in SLCs)
        
        assert gm.all_same(fmts), "All image_format of SLCs should " \
                                  "be the same!"
        
        log.info("SLC data format: %s" % fmts[0])
        
        self["slc"] = fmts[0]
        
        gm.rm(self.data["slc"].path)
        
        fmt = self.data["slc"].dcode()
        
        for ii, slc in enumerate(SLCs):
            gp.data2pt(slc.dat, slc.par, self.plist, self.slc.par,
                       self.data["slc"], ii + 1, fmt)
            gp.SLC_par_pt(slc.par, self.par, ii + 1, 1)
    
    
    def plot_pt(self, out, **kwargs):
        gm.output(out, **kwargs)
        
        fmt = gm.get_format(self.mli.img_fmt(), shape=self.mli)
        
        gm.cmd(
        """
        set view map
        plot '%s' %s with image notitle
        """ % (self.mli.dat, fmt)
        )
        
        gm.plot(**kwargs)
    
    def get_data(self, datfile, parfile, pdata, rec=None, dtype=None):
        if dtype is None:
            dtype = datfile.img_fmt()
        
        gp.data2pt(datfile, parfile, self.plist, self.slc.par, self.data[pdata],
                   rec, self.data[pdata].dcode())


    def npoints(self, mask=None):
        if mask is not None:
            mask = self.mask[mask]
            out = gp.npt(self.plist, mask).decode()
        else:
            out = gp.npt(self.plist).decode()
        
        
        for line in out.split("\n"):
            if line.startswith("total_number_of_points:"):
                return int(line.split(":")[1])
    
    
    
    def def_pt(self, diff, itab, model, mask=None, bflag="init", nref=0,
               res="res", dh="dh", defo="def", unw="unw", sigma="sigma",
               mask_out="def_mask", dh_max=60.0, def_min=-0.01, def_max=0.01,
               sigma_max=(1.2, 0.75), dh_err=None, def_err=None, ppc_err=None,
               bmax=-1, dtmax=-1, multi=False, noise_min=0, ref_mode=1,
               radius=7, rpatch=100, **kwargs):
    
        assert bflag in ("init", "prec")
        bflag = 1 if bflag == "prec" else 0
        
        
        diff = self.data[diff]
        mask = self.get_mask(mask)
        
        dtype = diff.dtype.upper()
        
        assert dtype in ("FLOAT", "FCOMPLEX")
        
        diff_type = 0 if dtype == "FLOAT" else 1
        
        self[res] = "float"
        self[dh] = "float"
        self[defo] = "float"
        self[unw] = "float"
        self[sigma] = "float"
        
        model = PointData.modes[model]
        
        if multi:
            gp.multi_def_pt(self.plist, mask, self.par, None, itab, self.base,
                            bflag, diff.path, diff_type, nref,
                            self.data[res], self.data[dh], self.data[defo],
                            self.data[unw], self.data[sigma],
                            self.mask[mask_out], dh_max, def_min, def_max,
                            rpatch, sigma_max[0], sigma_max[1], model,
                            noise_min, ref_mode, bmax, dtmax, radius, **kwargs)
        else:
            gp.def_mod_pt(self.plist, mask, self.par, None, itab, self.base,
                            bflag, diff.path, diff_type, nref,
                            self.data[res], self.data[dh], self.data[defo],
                            self.data[unw], self.data[sigma],
                            self.mask[mask_out], dh_max, def_min, def_max,
                            sigma_max[0], model, None, None, None, bmax,
                            dtmax, **kwargs)
        
        
        
        
    def ras_data(self, data, ras=None, out=None, mask=None, rec=None,
                 rng_looks=1, azi_looks=1, imsize=(800,800),
                 cycle=None, psize=1, cmap="hls.cm", dflg=1, wflg=0,
                 debug=False):
        
        datfile = self.data[data]
        
        dtype = datfile.dcode()
        
        out_tpl = gm.Files.get_tmp()
        
        if out is None:
            out = "%s.%s" % (datfile, gm.ras_ext)

        if rec is None:
            rec_num = 1
            nrec = None
        elif isinstance(rec, int):
            rec_num = 1
        else:
            rec_num, nrec = rec[0], rec[1]
        
        if mask is not None:
            mask = self.mask[mask]

        gp.ras_data_pt(self.plist, mask, datfile, rec_num, nrec,
                       ras, out_tpl, dtype, rng_looks, azi_looks,
                       cycle, psize, cmap, dflg, wflg, debug=debug)
        
        files = glob("%s*.bmp" % out_tpl)
        
        tmp = "tmp.bmp"
        
        try:
            gm.montage(tmp, *files, geometry="+2+2", size=imsize)
            gm.make_colorbar(tmp, out, cmap, stop=pi)
        except Exception as e:
            gm.rm(out_tpl + "*", tmp)
            raise e
        else:
            gm.rm(out_tpl + "*", tmp)
    
    
    def ras_pt(self, inras, outras, rng_looks=1, azi_looks=1,
               rgb="red", size=1, zero=False, mask=None,
               mask_flag=0, rec=None):
        """Plot point coordinates on raster file."""
        
        rgb = gm.colors[rgb]
        
        zero = 1 if zero else 0
        
        if mask is not None:
            mask = self.mask[mask]
        
        if rec is None:
            gp.ras_pt(self.plist, mask, inras, outras, rng_looks, azi_looks,
                      rgb, size, zero, mask_flag)
        else:
            gp.ras_pt(self.plist, mask, inras, outras, rng_looks, azi_looks,
                      rgb, size, zero, mask_flag, rec)
    
    
    def raster(self, data, mli=None, par_out=None, mode=None, mask=None,
               rec=None, radius=4.0, cycle=2*pi, outdir=None,
               search=3, imode=3, avg_fact=750, dis=False, debug=False):
        
        if outdir is None:
            outdir = self.data_dir
        
        
        _data = self.data[data]
        is_complex = _data.dtype.lower() in ("fcomplex", "scomplex")
        
        pdata, dtype = _data.path, _data.dcode()
        
        width = self.slc.rng()
        
        tmp = "tmp.prasdt_pwr24"
        gm.rm(tmp)
        
        npt = self.npoints()
        
        nbytes = pth.getsize(pdata)
        
        nrec = int(nbytes / npt / _data.size())
        
        base = pth.basename(pdata)

        if mli is None:
            mli = self.mli.dat
        
        if par_out is None:
            par_out = self.mli.par
        
        if mask is not None:
            mask = self.mask[mask]
        

        if dis:
            arg1 = "%s %d 1 1 0" % (mli, int(width))
            arg2 = "1.0 0.4"
        else:
            avg = self.slc.avg_fact(avg_fact)
            arg1 = "%s %d 1 1 0 %d %d" % (mli, int(width), avg[0], avg[1])
            arg2 = "1.0 0.4 1"
        
        
        if rec is None:
            start, stop = 1, nrec + 1
        elif isinstance(rec, int):
            start, stop = rec, rec + 1
        else:
            start, stop = rec[0], rec[1]

        
        for rec in range(start, stop):
            if rec < 10:
                fras = "%s.0%d.bmp" % (base, rec)
            else:
                fras = "%s.%d.bmp" % (base, rec)
            
            gp.pt2data(self.plist, mask, self.slc.par, pdata, rec, tmp,
                       par_out, dtype, imode, radius, search)

            print("data record: %d" % rec)
            
            
            if is_complex:
                if dis:
                    gp.dismph_pwr24(tmp, arg1, arg2)
                else:
                    gp.rasmph_pwr24(tmp, arg1, arg2, pth.join(outdir, fras))
            else:
                if dis:
                    gp.disdt_pwr24(tmp, arg1, arg2)
                else:
                    gp.rasdt_pwr24(tmp, arg1, cycle, arg2, pth.join(outdir, fras))
    
        gm.rm(tmp)


    def disp(self, data, mli=None, par_out=None, mode="mph", mask=None,
             rec=None, radius=4.0, cycle=2*pi, outdir=".", search=3,
             imode=3, dtype="fcomplex", debug=False):
        
        dtype = PointData.dtypes[dtype.upper()]
        
        plot_cmd = getattr(gp, "ras%s_pwr24" % mode)

        width = self.slc.rng()
        
        tmp = "tmp.prasdt_pwr24"
        gm.rm(tmp)
        
        npt = self.npoints()
        
        pdata = self.data[data].path
        
        nbytes = pth.getsize(pdata)
        nrec = int(nbytes / npt / 4)
        
        base = pth.basename(pdata)

        if mli is None:
            mli = self.mli.dat
        
        if par_out is None:
            par_out = self.mli.par
        
        if mask is not None:
            mask = self.mask[mask]
        
        arg1 = "%s %d 1 1 0 " % (mli, int(width))
        arg2 = "1.0, 0.4"
        
        if rec is None:
            start, stop = 1, nrec
        elif isinstance(rec, int):
            start, stop = rec, rec + 1
        else:
            start, stop = rec[0], rec[1]
            
        
        for rec in range(start, stop):
            if rec < 10:
                fras = "%s.0%d.bmp" % (base, rec)
            else:
                fras = "%s.%d.bmp" % (base, rec)
            
            gp.pt2data(self.plist, mask, self.slc.par, pdata, rec, tmp,
                       par_out, dtype, imode, radius, search)

            print("data record: %d" % rec)
            
            if mode == "mph":
                gp.dismph_pwr24(tmp, arg1, arg2)
            else:
                gp.disdt_pwr24(tmp, arg1, cycle, arg2)
    
        gm.rm(tmp)


    
def merge(*args, **kwargs):
    out = kwargs.get("out")
    n_occur = int(kwargs.get("n_occur", 1))
    rng_tol = int(kwargs.get("rng_tol", 0))
    azi_tol = int(kwargs.get("azi_tol", 0))
    
    if out is None:
        raise RuntimeError("out must be given")
    
    
    with NamedTemporaryFile() as f:
        f.write("%s\n" % "\n".join(pdata.plist for pdata in args))
        gp.merge_pt(f.name, out.plist, n_occur, rng_tol, azi_tol)



def get_llh(line):
    
    lat = line[19:35].split('-')
    lon = line[35:51].split('-')

    lat = float(lat[0]) + float(lat[1]) / 60.0 + float(lat[2]) / 3600.0
    lon = float(lon[0]) + float(lon[1]) / 60.0 + float(lon[2]) / 3600.0

    if line[61] == '1':
        is_reference = True
    else:
        is_reference = False

    return {"short_id": line[:3].strip(),
            "lon": lon,
            "lat": lat,
            "height": float(line[51:61]),
            "is_reference": is_reference}
    
