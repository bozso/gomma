package gamma;

import (
    "log";
    "fmt";
    //conv "strconv";
    str "strings";
);



var (
    burstCorner CmdFun;
);


const (
    burstTpl = "burst_asc_node_%d";
);



func init() {
    var ok bool;
    
    if burstCorner, ok = Gamma["ScanSAR_burst_corners"]; !ok {
        if burstCorner, ok = Gamma["SLC_burst_corners"]; !ok {
            log.Fatalf("No Fun.");
        }
    }
}


type (
    S1Zip struct {
        path, zip_base, mission, datestr, mode, prod_type, resolution string;
        level, prod_class, pol, abs_orb, DTID, UID string;
        date date;
    };
    
    IWInfo struct {
        num, nburst int;
        extent rect;
        bursts [9]float64;
    };
    
    S1IW struct {
        dataFile;
        TOPS_par ParamFile;
    };


    S1SLC struct {
        nIW int;
        IWs [9]S1IW;
        tab string;
    };
);


var extractRegex = map[string]string {
        "file": "{{mission}}-iw{{iw}}-slc-{{pol}}-.*-{{abs_orb}}-" + 
                "{{DTID}}-[0-9]{3}",
        "tiff":  "measurement/%s.tiff",
        "annot": "annotation/%s.xml",
        "calib": "annotation/%s.xml",
        "noise": "annotation/calibration/noise-%s.xml",
        "preview": "preview/product-preview.html",
        "quicklook": "preview/quick-look.png",
};


func (self *S1Zip) extracTemplates(names []string, pol, iw string) []string  {
    // TODO: finish
    // tpl = extractRegex["file"]
    
    //    tpl = self.extract_regex["file"]\
    //          .format(obj=self, mission=self.mission.lower(), 
    //                  DTID=self.DTID.lower(), pol=pol, iw=iw)
    //    return (Seq(self.extract_regex[name] for name in names)
    //                .map(str.format, tpl=tpl)
    //                .map(self.safe_join))
    
    return []string{"asd"}
}

func makePoint(info Params, max bool) (point, error) {
    handle := Handler("makePoint");
    
    var tpl_lon, tpl_lat string;
    
    if max {
        tpl_lon, tpl_lat = "Max_Lon", "Max_Lat";
    } else {
        tpl_lon, tpl_lat = "Min_Lon", "Min_Lat";
    }
    
    
    x, err := info.Float(tpl_lon);
    
    if err != nil {
        return point{}, handle(err, "Could not get Longitude value!");
    }
    
    y, err := info.Float(tpl_lat);
    
    if err != nil {
        return point{}, handle(err, "Could not get Latitude value!");
    }
    
    return point{x:x, y:y}, nil;
}


func iwInfo(path string) (*IWInfo, error) {
    handle := Handler("iwInfo");
    
    //num, err := conv.Atoi(str.Split(path, "iw")[1][0]);
    num := int(str.Split(path, "iw")[1][0])
    // Check(err, "Failed to retreive IW number from %s", path);
    
    par, err := TmpFile();
    
    if err != nil {
        return nil, handle(err, "Failed to create tmp file!");
    }
    
    TOPS_par, err := TmpFile();
    
    if err != nil {
        return nil, handle(err, "Failed to create tmp file!");
    }
    
    _info, err := Gamma["par_S1_SLC"](nil, path, nil, nil, par, nil, TOPS_par);
    
    if err != nil {
        return nil, handle(err, "Failed to parse parameter files!");
    }
    
    info := FromString(_info, ":");
    TOPS, err := FromFile(TOPS_par, ":");
    
    if err != nil {
        return nil, handle(err, "Could not parse TOPS_par file!");
    }
    
    nburst, err := TOPS.Int("number_of_bursts");
    
    if err != nil {
        return nil, handle(err, "Could not retreive number of bursts!");
    }
    
    var numbers [9]float64;
    
    for ii := 0; ii < nburst; ii++ {
         tpl := fmt.Sprintf(burstTpl, ii);
         
         numbers[ii], err = TOPS.Float(tpl);
         
         if err != nil {
            return nil, handle(err, "Could not get burst number: '%s'",
                                     tpl);
         }
    }
    
    max, err := makePoint(info, true);
    
    if err != nil {
        return nil, handle(err, "Could not create max latlon point!");
    }
    
    min, err := makePoint(info, false);
    
    if err != nil {
        return nil, handle(err, "Could not create min latlon point!");
    }
    
    return &IWInfo{num:num, nburst:nburst, extent:rect{min:min, max:max},
                   bursts:numbers}, nil;
}



func (self *point) inIWs(IWs []IWInfo) bool {
    for _, iw := range IWs {
        if self.inRect(&iw.extent) {
            return true;
        }
    }
    return false;
}


func pointsInSLC(IWs []IWInfo, points [4]point) bool {
    sum := 0;
    
    for _, point := range points {
        if point.inIWs(IWs) {
            sum++;
        }
    }
    return sum == 4;
}



/*

def safe(path):
    return filter(lambda x: ".SAFE" in x, pth.normpath(path).split(pth.sep))
    

class S1Zip(object):
    if hasattr(gm, "ScanSAR_burst_corners"):
        cmd = "ScanSAR_burst_corners"
    else:
        # fallback
        cmd = "SLC_burst_corners"
    
    burst_fun = getattr(gp, cmd)
    
    regex_names = set(extract_regex.keys()) - {"file"}
    
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
    
    

burst_tpl = "burst_asc_node_%d"


def iw_info_from_annot(path):
    iw_num = int(path.split("iw")[1][0])
    
    par, TOPS_par = tmp_file(), tmp_file()
    
    gp.par_S1_SLC(None, path, None, None, par, None, TOPS_par)
    
    info = S1Zip.burst_fun(par, TOPS_par).decode()
    nburst = int(get_par("number_of_bursts", TOPS_par))
    
    numbers = tuple(float(get_par(burst_tpl % (ii + 1), TOPS_par).split()[0])
                    for ii in range(nburst))
    
    
    max_lat, min_lat, max_lon, min_lon = \
    get_par("Max_Lat", info), get_par("Min_Lat", info), \
    get_par("Max_Lon", info), get_par("Min_Lon", info)
    
    return IWInfo(iw_num, 
                  gm.Rect(float(max_lon), float(min_lon),
                          float(max_lat), float(min_lat)),
                  numbers)


def use_extracted(self, names, extracted, **kwargs):
    return (self.extract_templates(names, **kwargs)
                .map(gm.filter_file, namelist=extracted.file_list)
                .chain()
                .map(partial(pth.join, extracted.outpath)))


def iw_info(self, **kwargs):
    return (use_extracted(self, ("annot",), **kwargs)
            .map(iw_info_from_annot))
    

def point_in_IWs(point, IWs):
    return any(gm.point_in_rect(point, rect=iw.rect) for iw in IWs)


def points_in_IWs(IWs, points):
    return points.map(point_in_IWs, IWs=IWs.collect())


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

*/