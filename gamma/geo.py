import gamma as gm

from os import path as pth
from logging import getLogger
from collections import namedtuple


__all__ = (
    "DEM",
    "Geocode",
    "geocode"
)


RDC = namedtuple("RDC", "rng azi")

log = getLogger("gamma.geo")

gp = gm.gp


class DEM(gm.DataFile):
    __slots__ = {"lookup"}
    __save__ = {"dat", "par", "lookup"}
    
    _geo2rdc = {
        "dist": 0,
        "nearest_neigh": 1,
        "sqr_dist": 2,
        "const": 3,
        "gauss": 4
    }
    
    _rdc2geo = {
        "nearest_neigh": 0
    }
    

    def __init__(self, datfile, parfile=None, lookup=None):
        self.dat = datfile
        
        if parfile is None:
            parfile = pth.splitext(datfile) + ".dem_par"
        
        self.par, self.lookup = parfile, lookup
    
    
    @classmethod
    def from_json(cls, line):
        return cls(line["dat"], line["par"], line["lookup"])
        
        
    def rng(self):
        return self.int("width")
    
    def azi(self):
        return self.int("nlines")
    
    
    def geo2rdc(self, infile, outfile, width, nlines=0, interp="dist",
                dtype=0):

        _interp = DEM._geo2rdc[interp]

        gp.geocode(self.lookup, infile, self.rng(), outfile, width, nlines,
                   _interp, dtype)


    # TODO: interpolation modes
    def rdc2geo(self, infile, outfile, width, nlines=0, interp=1, dtype=0,
                flip_in=False, flip_out=False, order=5, func=None):
        
        lr_in = -1 if flip_in else 1
        lr_out = -1 if flip_out else 1
        
        gp.geocode_back(infile, self.rng(), self.lookup, outfile, width,
                        nlines, interp, dtype, lr_in, lr_out, order)
    
    
    def calc_rdc(latlon, mpar, hgt, diff_par):
            out = gp.coord_to_sarpix(mpar, None, self.par, latlon[0],
                                     latlon[1], hgt, diff_par).decode()
        
            for line in out.split("\n"):
                if "corrected SLC/MLI range, azimuth pixel (int)" in line:
                    split = line.split(":")[1].split()
                    return RDC(int(split[0]), int(split[1]))

    def raster(self, obj, **kwargs):
        assert obj in {"lookup", "lookup_old"}
        
        kwargs.setdefault("image_format", "FCOMPLEX")
        kwargs.setdefault("datfile", getattr(self, obj))
        kwargs.setdefault("mode", "mph")
        DataFile.raster(self, **kwargs)


class Geocode(gm.DataFile):
    _items = {"hgt", "sim_sar", "zenith", "orient", "inc", "pix", "psi",
              "ls_map", "diff_par", "offs", "offsets", "ccp", "coffs",
              "coffsets"}
    
    __save__ = {"dat", "par"} | _items
    
    rashgt = getattr(gp, "rashgt")
    
    
    def __init__(self, mli=None, **kwargs):
        if mli is not None:
            self.dat = mli.dat
            self.par = mli.par
        
        self.__dict__.update(kwargs)
    
    
    
    @classmethod
    def from_json(cls, line):
        ret = cls(**line)
        ret.dat, ret.par = line["dat"], line["par"]
        
        return ret
    
    
    def raster(self, start_hgt=None, start_pwr=None, m_per_cycle=None,
               **kwargs):
        args = DataFile.parse_ras_args(self, **kwargs)
        
        Geocode.rashgt(args["datfile"], self.mli.dat, args["rng"],
                       start_hgt, start_pwr, args["nlines"], args["arng"],
                       args["aazi"], m_per_cycle, args["scale"], args["exp"],
                       args["LR"], args["raster"], args["debug"])
    


def geocode(params, m_slc, m_mli, rng_looks=1, azi_looks=1, out_dir="."):
    
    demdir = pth.join(out_dir, "dem")
    geodir = pth.join(out_dir, "geo")
    
    djoin, gjoin = gm.make_join(demdir), gm.make_join(geodir)
    
    dem_orig = DEM(datfile=pth.join(demdir, "srtm.dem"))
    
    vrt_path = params.get("dem_path")

    if vrt_path is None:
        raise ValueError("dem_path is not defined!")
    
    dem_lat_ovs = params.getfloat("dem_lat_ovs", 1.0)
    dem_lon_ovs = params.getfloat("dem_lon_ovs", 1.0)

    n_rng_off = params.getint("n_rng_off", 64)
    n_azi_off = params.getint("n_azi_off", 32)

    rng_ovr = params.getint("rng_overlap", 100)
    azi_ovr = params.getint("azi_overlap", 100)

    npoly = params.getint("npoly", 4)
    itr = params.getint("iter", 0)
    
    
    if not dem_orig.exist("dat"):
        log.info("Creating DEM from %s." % vrt_path)
        
        gp.vrt2dem(vrt_path, mmli.par, dem_orig, 2, None)
    else:
        log.info("DEM already imported.")

    mli_rng, mli_azi = m_mli.rng(), m_mli.azi()
    
    rng_patch, azi_patch = int(mli_rng / n_rng_off + rng_ovr / 2), \
                           int(mli_azi / n_azi_off + azi_ovr / 2)
    
    
    # make sure the number of patches are even
    if rng_patch % 2: rng_patch += 1
    
    if azi_patch % 2: azi_patch += 1

    dem = DEM(djoin("dem_seg.dem"), parfile=djoin("dem_seg.dem_par"),
              lookup=gjoin("lookup"), lookup_old=gjoin("lookup_old"))
    
    
    items = {
        name: gjoin(name)
        for name in {"hgt", "sim_sar", "zeinth", "orient", "inc", "pix",
                     "psi", "ls_map", "diff_par", "offs", "offsets", "ccp",
                     "coffs", "coffsets"}
    }
    
    geo = gm.Geocode(mli=m_mli, **items)    
    
    if not dem.exist("lookup", "par"):
        log.info("Calculating initial lookup table.")
        gp.gc_map(mmli.par, None, dem_orig.par, dem_orig.dat,
                  dem.par, dem.dat, dem.lookup, dem_lat_ovs, dem_lon_ovs,
                  geo.sim_sar, geo.zenith, geo.orient, geo.inc, geo.psi,
                  geo.pix, geo.ls_map, 8, 2)
    else:
        log.info("Initial lookup table already created.")

    dem_s_width = dem["width"]
    dem_s_lines = dem["lines"]

    gp.pixel_area(m_mli.par, dem.par, dem.dat, dem.lookup, geo.ls_map,
                  geo.inc, geo.sigma0, geo.gamma0, 20)
    
    gp.create_diff_par(m_mli.par, None, geo.diff_par, 1, 0)
    
    log.info("Refining lookup table.")

    if itr >= 1:
        log.info("ITERATING OFFSET REFINEMENT.")

        for ii in range(itr):
            log.info("ITERATION %d / %d" % (ii + 1, itr))

            geo.rm("diff_par")

            # copy previous lookup table
            dem.cp("lookup", dem.lookup_old)

            gp.create_diff_par(m_mli.par, None, geo.diff_par, 1, 0)

            gp.offset_pwrm(geo.sigma0, m_mli.dat, geo.diff_par, geo.offs,
                           geo.ccp, rng_patch, azi_patch, geo.offsets, 2,
                           n_rng_off, n_azi_off, 0.1, 5, 0.8)

            gp.offset_fitm(geo.offs, geo.ccp, geo.diff_par, geo.coffs,
                           geo.coffsets, 0.1, npoly)

            # update previous lookup table
            gp.gc_map_fine(dem.lookup_old, dem_s_width, geo.diff_par,
                           dem.lookup, 1)

            # create new simulated ampliutides with the new lookup table
            gp.pixel_area(m_mli.par, dem.par, dem.dat, dem.lookup, geo.ls_map,
                          geo.inc, geo.sigma0, geo.gamma0, 20)

        # end for
        log.info("ITERATION DONE.")
    # end if
    
    
    return {
        "geo": geo,
        "dem_orig": dem_orig,
        "dem": dem
    }
