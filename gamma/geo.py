import gamma as gm

__all__ = [
    "DEM",
    "HGT",
    "Geocode"
]

gp = gm.gamma_progs


class DEM(gm.DataFile):
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
    

    def __init__(self, datfile, parfile=None, lookup=None, lookup_old=None,
                 keep=True):
        self.dat, self.keep = datfile, None
        
        if parfile is None:
            parfile = pth.splitext(datfile) + ".dem_par"
        
        self.par, self.lookup, self.lookup_old = parfile, lookup, lookup_old
        self.keep = keep
    
    
    def rng(self):
        return self.getint("par", "width")
    
    def azi(self):
        return self.getint("par", "nlines")
    
    
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
    
    
    def raster(self, obj, **kwargs):
        assert obj in ("lookup", "lookup_old")
        
        kwargs.setdefault("image_format", "FCOMPLEX")
        kwargs.setdefault("datfile", getattr(self, obj))
        kwargs.setdefault("mode", "mph")
        DataFile.raster(self, **kwargs)


class Geocode(gm.Files):
    _items = ("sim_sar", "zenith", "orient", "inc", "pix", "psi", "ls_map",
              "diff_par", "offs", "offsets", "ccp", "coffs", "coffsets")

    def __init__(self, path, mli, sigma0=None, gamma0=None, **kwargs):
        self.par = mli.par
        
        if sigma0 is None:
            sigma0 = pth.join(path, "sigma0")

        if gamma0 is None:
            gamma0 = pth.join(path, "gamma0")
        
        self.sigma0, self.gamma0 = sigma0, gamma0
        
        
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


class HGT(gm.DataFile):
    rashgt = getattr(gp, "rashgt")
    
    def __init__(self, hgt, mli, keep=True):
        self.keep = None
        self.dat = hgt
        self.mli = mli
        self.par = mli.par
        
        self.keep = keep
    
    
    def __str__(self):
        return self.dat
    
    def raster(self, start_hgt=None, start_pwr=None, m_per_cycle=None,
               **kwargs):
        args = DataFile.parse_ras_args(self, **kwargs)
        
        HGT.rashgt(args["datfile"], self.mli.dat, args["rng"],
                   start_hgt, start_pwr, args["nlines"], args["arng"],
                   args["aazi"], m_per_cycle, args["scale"], args["exp"],
                   args["LR"], args["raster"], args["debug"])
