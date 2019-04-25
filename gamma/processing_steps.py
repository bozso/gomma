import logging
import os.path as pth
import os

from datetime import datetime
from glob import iglob
from pickle import dump, load
from collections import OrderedDict
from pprint import pprint
from shutil import copy
from sys import stdout
from imp import load_source


try:
    from configparser import SafeConfigParser
except ImportError:
    from ConfigParser import SafeConfigParser


import gamma as gm

gp, narg = gm.gamma_progs, gm.Argp.narg

log = logging.getLogger("gamma.steps")


__all__ = ("Processing", "make_step", "ListIter")

# *********************
# * Utility functions *
# *********************

# _terminal_width = gp.get_terminal_size().columns


def make_all():
    pass


def delim(msg, sym="*", width=80):
    msg = str(msg)
    msg = "%s %s %s" % (sym, msg, sym)

    syms = sym * len(msg)
    
    tpl = "\n{{:^{w}}}\n{{:^{w}}}\n{{:^{w}}}\n".format(w=width)
    
    print(tpl.format(syms, msg, syms))

    
def pk_save(obj, path):
    with open(path, "wb") as f:
        dump(obj, f)


def pk_load(path):
    with open(path, "rb") as f:
        obj = load(f)
    
    return obj


def pk_update(path, _dict):
    with open(path, "rb") as f:
        obj = load(f)
    
    obj.update(_dict)
    
    with open(path, "wb") as f:
        dump(obj, f)


def make_step(name, function, required=True):
    return (name, (function, required))


class ListIter(object):
    def __init__(self, path, converter):
        self.path = path
        self.conv = converter
    
    def __enter__(self):
        self.fp = open(self.path, "r")
        return self
    
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        self.fp.close()

    def __iter__(self):
        for line in self.fp:
            yield self.conv(line)


class Processing(object):
    _default_log_format = "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
    
    converters = {
        "S1SLC" : gm.S1SLC.from_tabfile,
        "SLC" : gm.SLC.from_line,
        "MLI" : gm.MLI.from_line
    }
    

    def __init__(self, *steps):
        self._steps = OrderedDict(steps)
        
        self.optional_steps = tuple(key for key, item in self._steps.items()
                                    if not item[1])
        
        self.required_steps = tuple(key for key, item in self._steps.items()
                                    if item[1])
        

    def setup_log(self, logger_name, filename=None, formatter=None,
                  loglevel="debug"):
        
        logger = logging.getLogger(logger_name)
        
        level = getattr(logging, loglevel.upper(), None)
        
        if not isinstance(level, int):
            raise ValueError("Invalid log level: {}".format(loglevel))
        
        logger.setLevel(level)
        
        if formatter is None:
            formatter = self._default_log_format
        
        form = logging.Formatter(formatter, datefmt="%Y.%m.%d %H:%M:%S")
        
        
        if filename is not None:
            fh = logging.FileHandler(filename)
            fh.setFormatter(form)
            logger.addHandler(fh)
        
        consoleHandler = logging.StreamHandler()
        consoleHandler.setFormatter(form)
        
        logger.addHandler(consoleHandler)    
    
        return logger

    
    def steps(self):
        return self._steps.keys()
    
    
    def parse_steps(self, args):
        
        step = args.step
        
        _steps = tuple(self._steps.keys())
        
        if step is not None:
            if step not in _steps:
                raise ValueError("Single step \"%s\" is not a valid processing "
                                 "step. Choose from: %s"
                                 % (step, ", ".join(steps)))
            
            log.debug("Single step \"%s\" is executed." % step)
            return [step]
        else:
            start = args.start
            stop  = args.stop
            
            if start not in _steps:
                raise ValueError("Step \"%s\" is not a valid processing "
                                 "step. Choose from: %s"
                                 % (start, ", ".join(steps)))
    
            if stop not in _steps:
                raise ValueError("Step \"%s\" is not a valid processing "
                                 "step. Choose from: %s"
                                 % (stop, ", ".join(steps)))
    
            log.debug("Steps from \"%s\" to \"%s\" will be executed."
                       % (start, stop))
            first = _steps.index(start)
            last = _steps.index(stop)
            return _steps[first:last + 1]


    def show_steps(self):
        print("\nProcessing steps: %s\nOptional steps: %s\n"
              % (", ".join(step for step in self.required_steps),
                 ", ".join(step for step in self.optional_steps)))
    
    
    def add_args(self, conf_file="gp_conf.py", log_file="gamma_proc.log"):
        _steps = tuple(self.steps())
        return (
            narg("step", help="Single processing step to be executed",
                      alt="S", type=str, choices=_steps),
        
            narg("start", help="Starting processing step. Processing steps will "
                 "be executed until processing step defined by --stop is "
                 "reached.", alt="s", type=str, default=_steps[0],
                 choices=_steps),
        
            narg("stop", help="Last processing step to be executed.", alt="e",
                 type=str, default=_steps[-1], choices=_steps),
        
            narg("paramfile", help="Text file that contains the GAMMA processing "
                 "parameters.", alt="p", type=str, default=conf_file),
        
            narg("logfile", help="Log messages will be saved here.", alt="l",
                 type=str, default=log_file),
        
            narg("loglevel", help="Level of logging.", alt="g", type=str,
                 default="info"),
        
            narg("skip_optional", help="If set the processing will skip optional "
                 "steps.", kind="flag", alt="o"),
    
            narg("show_steps", help="If set, just print the processing steps.",
                 kind="flag", alt="t"),
    
            narg("info", help="Dumps information about the processing to "
                 "the terminal.", kind="flag", alt="i")
        )
    
    
    def get_dir(self, name):
        try:
            return self.meta["dirs"][name]
        except KeyError:
            _path = pth.join(self.params.general["output_dir"], name)
            os.mkdir(_path)
            self.meta["dirs"][name] = _path
            return _path


    def get_list(self, name):
        try:
            return self.meta["lists"][name]
        except KeyError:
            _path = pth.join(self.get_dir("list_dir"), "%s.file_list" % name)
            self.meta["lists"][name] = _path
            return _path


    def inlist(self, name, conv):
        return ListIter(self.get_list(name), Processing.converters[conv])

    def outlist(self, name):
        return open(self.get_list(name), "w")


    def parse_args(self, args):
        self.setup_log("gamma", filename=args.logfile, loglevel=args.loglevel)
        self.run_steps(args)
    

    def load(self):
        metafile = self.params.general["metafile"]

        if pth.isfile(metafile):
            self.meta = pk_load(metafile)
        else:
            self.meta = {"dirs": {}, "lists": {}}
    
    
    def save(self):
        pk_save(self.meta, self.params.general["metafile"])
        
    
    def run_steps(self, args):
        
        if args.show_steps:
            self.show_steps()
            return
        
        log.info("File containing processing parameters: %s" 
                 % (args.paramfile))
        
        self.params = load_source("params", args.paramfile)
        

        if args.info:
            pprint(self.params)
            pprint(self.meta)
            return
        
        steps = self.parse_steps(args)
        
        optional_steps = not args.skip_optional
        
        if not optional_steps:
            log.info("Skipping optional steps.")
        
        self.load()
        
        for step in steps:
            step_fun = self._steps[step]
            
            try:
                if step_fun[1] or optional_steps:
                    log.info("Running step: \"%s\"" % step)
                    step_fun[0](self)
            except Exception as e:
                self.save()
                raise e
            
            self.save()

    
    def get_out_master(self):
        general = self.params.general
        
        output_dir  = general.get("output_dir", ".")
        master_date = self.meta.get("master_date")
        
        if master_date is None:
            raise ValueError("master_date is not defined.")
        
        return output_dir, master_date
        

def select_bursts(self):

    delim("Starting SELECT_BURSTS")

    general = self.params.general
    
    slc_data = general.get("slc_data")

    if slc_data is None:
        raise ValueError('Parameter "slc_data" not defined.')


    master_date, output_dir = general.get("master_date"), \
                              general.get("output_dir", ".")

    date_start, date_stop = general.get("date_start"), general.get("date_stop")

    check_zips = general.get("check_zips", False)
    pol = general.get("pol")
    
    
    
    IWs = tuple(general.get("iw%d" % (idx + 1)) for idx in range(3))
    
    IWs = tuple(
        tuple(
            int(elem)
            for elem in IW.split(",")
             ) if IW is not None else None
        for IW in IWs
    )
    

    if IWs[2] is not None and IWs[1] is None:
        raise ValueError("Selected IWs must be contigous. You must have "
                         "selected bursts in IW2 to have bursts in IW3")
    
    log.info("Creating SLC directories, checking dates and creating "
             "zipfile list.")
    
    SLC = (gm.S1Zip(zipfile)
           for zipfile in iglob(pth.join(slc_data, "S1*_IW_SLC*.zip")))
    
    if date_start is not None and date_stop is not None:
        date_start = datetime.strptime(date_start, "%Y%m%d")
        date_stop = datetime.strptime(date_stop, "%Y%m%d")
        
        SLC = (slc for slc in SLC
               if slc.date.start > date_start and slc.date.stop < date_stop)
    
    elif date_start is not None:
        date_start = datetime.strptime(date_start, "%Y%m%d")

        SLC = (slc for slc in SLC if slc.date.start > date_start)


    elif date_stop is not None:
        date_stop = datetime.strptime(date_stop, "%Y%m%d")
        SLC = (slc for slc in SLC if slc.date.stop < date_stop)

    
    if check_zips:
        SLC = (slc for slc in SLC if slc.test_zip())
    
    SLC = tuple(SLC)
    dates = tuple(slc.date for slc in SLC)
    self.meta["dates"] = dates
    
    if master_date is None:
        log.info("No master_date defined, using first date.")
        
        master_slc = sorted(SLC, key=lambda x: x.date.mean)[0]
        master_date = master_slc.date.date2str()
        
        self.meta["master_date"] = master_date
    else:
        master_date = general["master_date"]
        master_slc = [slc for slc in SLC
                      if slc.date.date2str() == master_date][0]


    log.info("Selected master date is %s" % master_date)
    
    # if we have selected bursts
    if any(IWs):
        master_burst_nums = master_slc.get_burst_nums(pol)
        
        master_burst_nums = \
        tuple(None if IW is None
              else tuple((master_burst[IW[0] - 1], master_burst[IW[-1] - 1]))
              for IW, master_burst in zip(IWs, master_burst_nums))
        
        burst_nums = tuple(slc.select_bursts(pol, master_burst_nums)
                           for slc in SLC)
    # end select AOE
    
    
    S1A_bursts = set(
            " ".join(
            "IW%d: %s" % (ii + 1, item)
            for ii, item in enumerate(burst_num)
            )
        for slc, burst_num in zip(SLC, burst_nums) if slc.mission == "S1A"
    )

    S1B_bursts = set(
            " ".join(
            "IW%d: %s" % (ii + 1, item)
            for ii, item in enumerate(burst_num)
            )
        for slc, burst_num in zip(SLC, burst_nums) if slc.mission == "S1B"
    )

    #log.debug("Master bursts:\n{}\n\n".format(S1A_bursts))
    log.info("\nSentinel-1A slave bursts:\n%s" % S1A_bursts)
    log.info("\nSentinel-1B slave bursts:\n%s" % S1B_bursts)
    
    self.meta["SLC_zip"] = SLC
    self.meta["burst_nums"] = burst_nums

    gm.rm("tmp_par", "tmp_TOPS_par")
    
    print(delim("Finished SELECT_BURSTS"))


def import_slc(self):

    if gp.ScanSAR:
        copy_fun = getattr(gp, "SLC_copy_ScanSAR")
    else:
        copy_fun = getattr(gp, "SLC_copy_S1_TOPS")

    print(delim("Starting IMPORT_SLCS"))
    
    output_dir, master_date = self.get_out_master()
    pol = self.params.general.get("pol")
    
    dir_uncrop = self.get_dir("SLC_uncrop")
    dir_crop = self.get_dir("SLC_crop")


    meta = self.meta
    SLC, burst_nums, dates = meta["SLC_zip"], meta["burst_nums"], meta["dates"]
    
    log.info("Importing SLCs.")
    
    tpl = pth.join(dir_uncrop, "{date}_iw{iw}.{pol}.slc")
    tpl_tab = pth.join(dir_uncrop, "{date}.{pol}.SLC_tab")


    SLC_uncrop = \
    tuple(
        gm.S1SLC.from_template(pol, date, burst_num, tpl=tpl,
                               tpl_tab=tpl_tab)
        for burst_num, date in zip(burst_nums, dates)
    )


    tpl = pth.join(dir_crop, "{date}_iw{iw}.{pol}.slc")
    tpl_tab = pth.join(dir_crop, "{date}.{pol}.SLC_tab")
    
    
    SLC_crop = \
    tuple(
        gm.S1SLC.from_template(pol, date, burst_num, tpl=tpl,
                               tpl_tab=tpl_tab)
        for burst_num, date in zip(burst_nums, dates)
    )

    for slc, uncrop in zip(SLC, SLC_uncrop):
        for IW in uncrop.IWs:
            if IW is not None:
                slc.extract_IW(pol, IW)
    
    for uncrop, crop, burst_num in zip(SLC_uncrop, SLC_crop, burst_nums):
        with open("tmp_burst_tab", "w") as f:
            f.write("\n".join("%d %d" % (burst[0], burst[1])
                    for burst in burst_num if burst is not None) + "\n")
        
        copy_fun(uncrop.tab, crop.tab, "tmp_burst_tab")

    self.meta["SLC_uncrop"] = SLC_uncrop
    self.meta["SLC_crop"] = SLC_crop
    
    gm.rm("tmp_burst_tab")
    
    print(delim("Finished IMPORT_SLCS"))


def merge_slcs(self):

    if gp.ScanSAR:
        merge_fun = getattr(gp, "SLC_cat_ScanSAR")
    else:
        merge_fun = getattr(gp, "SLC_cat_S1_TOPS")

    
    general = self.params.general
    
    output_dir = general.get("output_dir", ".")
    pol        = general.get("pol", "vv")
    slc_dir    = self.get_dir("SLC_merged")

    
    SLC, dates = self.meta["SLC_crop"], self.meta["dates"]
    SLC_merged, used_SLC = [], []
    
    tpl_iw = pth.join(slc_dir, "{}_iw{}.{}.slc")
    tpl_tab = pth.join(slc_dir, "{}.{}.SLC_tab")
    
    list_merged = self.get_list("merged")
    
    with self.outlist("merged", "w") as f:
        for SLC1 in SLC:
            if SLC1 in used_SLC:
                continue
            
            date1str = SLC1.date.date2str()
            log.info("Processing date %s." % date1str)
            SLC2 = gm.search_pair(SLC1, SLC, used_SLC)
    
            SLC3 = gm.S1SLC(
            tuple(
                    gm.S1IW(tpl_iw.format(date1str, ii + 1, pol), num=ii + 1)
                    if IW is not None else None
                    for ii, IW in enumerate(SLC1.IWs)
                ), tpl_tab.format(date1str, pol), date=SLC1.date
            )
            
            
            if SLC2 is not None:
                log.info("Merging %s with %s." % (SLC1.tab, SLC2.tab))
                
                if SLC1.date.mean > SLC2.date.mean:
                    gp.SLC_cat_S1_TOPS(SLC2.tab, SLC1.tab, SLC3.tab)
                    merge_fun(SLC2.tab, SLC1.tab, SLC3.tab)
                else:
                    merge_fun(SLC1.tab, SLC2.tab, SLC3.tab)
    
                used_SLC.append(SLC2)
            else:
                log.info("No need for merge. Copying %s." % SLC1.tab)
                SLC1.cp(SLC3)
    
            # endif
            f.write("%s\n", SLC3.tab)
        # endfor
    # close
    
    self.meta["dates"] = tuple(slc.date for slc in SLC_merged)
    self.meta["merged"] = list_merged

    # CLEANUP
    #gm.rm("*.SAFE", "*.SLC_tab", "*iw*", "*.slc*")


def quicklook_mli(self):
    
    general = self.params.general
    
    output_dir    = general.get("output_dir", ".")
    range_looks   = general.get("range_looks", 1)
    azimuth_looks = general.get("azimuth_looks", 4)
    pol           = general.get("pol")
    
    log.info("CREATING QUICKLOOK MLIs.")
    mli_dir = self.get_dir("MLI")
    
    list_merged = self.meta["merged"]
    
    tpl = pth.join(mli_dir, "%s.%s.mli")
    
    with self.inlist("merged", "S1SLC") as SLC, self.outlist("mli", "w") as f:
        for mli, slc in zip(MLI, SLC):
            # create MLI file paths
            mli = gm.MLI(tpl % (slc.date2str(), pol))
            
            # multi looking
            slc.multi_look(mli, range_looks, azimuth_looks)
            
            # quicklook gm.ras_extter
            gm.gm.ras_extter(mli.dat, parfile=mli.par, avg_fact=750)
            
            # write to file
            f.write("%s\n" % mli)
    

def mosaic_tops(self):
    
    general = self.params.general
    
    output_dir      = general.get("output_dir", ".")
    range_looks     = general.get("range_looks", 1)
    azimuth_looks   = general.get("azimuth_looks", 4)
    pol             = general.get("pol", "vv")
    
    log.info("MOSAICING SENTINEL IWs.")
    
    for S1slc in self.meta["SLC_merged"]:
        S1slc.mosaic(rng_looks=rng_looks, azi_looks=azi_looks)


def check_ionoshpere(self):
    
    output_dir = self.params.general.get("output_dir", ".")
    
    try:
        SLC = self.meta["SLC_mosaic"]
    except KeyError:
        mosaic_tops(self)
        SLC = self.meta["SLC_mosaic"]
    
    
    check_iono = self.params.check_ionosphere
    
    rng_win = check_iono.get("rng_win", 256)
    azi_win = check_iono.get("azi_win", 256)
    thresh  = check_iono.get("iono_thresh", 0.1)
    
    rng_step = check_iono.get("rng_step")
    azi_step = check_iono.get("azi_step")
    
    SLC[0].check_ionoshpere(rng_win=rng_win, azi_win=azi_win, thresh=thresh,
                            rng_step=rng_step, azi_step=azi_step)
    

def quicklook_rmli(self):
    
    general = self.params.general
    
    output_dir    = general.get("output_dir", ".")
    range_looks   = general.get("range_looks", 1)
    azimuth_looks = general.get("azimuth_looks", 4)
    pol           = general.get("pol")
    
    log.info("CREATING QUICKLOOK RMLIs.")
    mli_dir = gm.mkdir(pth.join(output_dir, "RMLI"))
    
    RSLC = self.meta["SLC_coreg"]
    
    tpl = pth.join(mli_dir, "%s.%s.mli")
    
    datestr = (date.date2str() for date in self.meta["dates"])
    
    RMLI = tuple(MLI(tpl % (date, pol)) for date in datestr)

    for rmli, rslc in zip(MLI, RSLC):
        rslc.multi_look(rmli, rng_looks=range_looks, azi_looks=azimuth_looks)

    for rmli in RMLI:
        gm.ras_extter(rmli.dat, parfile=rmli.par, comp_fact=750)
    
    self.meta["RMLI"] = RMLI


def geocode_master(self):

    log.info("Starting GEOCODE_MASTER.")
    
    output_dir, master_date = self.get_out_master()
    general = self.params.general
    
    rng_looks = general.get("range_looks", 1)
    azi_looks = general.get("azimuth_looks", 4)
    
    
    geoc = self.params.geocoding
    
    vrt_path = geoc.get("dem_path")

    if vrt_path is None:
        raise ValueError("dem_path is not defined!")
    
    dem_lat_ovs = geoc.get("dem_lat_ovs", 1.0)
    dem_lon_ovs = geoc.get("dem_lon_ovs", 1.0)

    n_rng_off = geoc.get("n_rng_off", 64)
    n_azi_off = geoc.get("n_azi_off", 32)

    rng_ovr = geoc.get("rng_overlap", 100)
    azi_ovr = geoc.get("azi_overlap", 100)

    npoly = geoc.get("npoly", 4)

    itr = geoc.get("iter", 0)
    
    demdir = gm.mkdir(pth.join(output_dir, "dem"))
    geodir = gm.mkdir(pth.join(output_dir, "geo"))
    
    merged = self.meta["SLC_merged"]

    if "master" in self.meta:
        master = self.meta["master"]
    else:
        master = {}
        self.meta["master"] = {}
    
    if "S1SLC" in master:
        master_S1slc = master["S1SLC"]
    else:
        master_S1slc = tuple(slc for slc in merged
                             if slc.date.date2str() == master_date)[0]
        master["S1SLC"] = master_S1slc
    
    
    if master_S1slc.slc is None:
        master_S1slc.mosaic(rng_looks=rng_looks, azi_looks=azi_looks)

        
    if "MLI" in master:
        mmli = master["MLI"]
    else:
        mlidir = gm.mkdir(pth.join(output_dir, "MLI"))
        mmli = gm.MLI(pth.join(mlidir, "%s.mli" % master_date))
        
        master_S1slc.slc.multi_look(mmli, rng_looks=rng_looks,
                                    azi_looks=azi_looks)
        
        master["MLI"] = mmli
    
    
    self.meta["master"].update(master)
    
    
    dem_orig = gm.MLI(pth.join(demdir, "srtm.dem"),
                      parfile=pth.join(demdir, "srtm.dem_par"))
    
    
    if not dem_orig.exist("dat"):
        log.info("Creating DEM from %s." % vrt_path)
        
        gp.vrt2dem(vrt_path, mmli.par, dem_orig, 2, None)
    else:
        log.info("DEM already imported.")


    geo_path = gm.mkdir(pth.join(geodir))

    mli_rng, mli_azi = int(mmli.rng()), int(mmli.azi())
    
    rng_patch, azi_patch = int(mli_rng / n_rng_off + rng_ovr / 2), \
                           int(mli_azi / n_azi_off + azi_ovr / 2)
    
    # make sure the number of patches are even
    if rng_patch % 2: rng_patch += 1
    
    if azi_patch % 2: azi_patch += 1

    dem = gm.DEM(pth.join(demdir, "dem_seg.dem"),
                 parfile=pth.join(demdir, "dem_seg.dem_par"),
                 lookup=pth.join(geo_path, "lookup"),
                 lookup_old=pth.join(geo_path, "lookup_old"))
    
    
    geo = gm.Geocode(geo_path, mmli, sim_sar="sim_sar", zenith="zenith",
                     orient="orient", inc="inc", pix="pix", psi="psi",
                     ls_map="ls_map", diff_par="diff_par", offs="offs",
                     offsets="offsets", ccp="ccp", coffs="coffs",
                     coffsets="coffsets")
    
    
    if not (dem.exist("lookup") and dem.exist("par")):
        log.info("Calculating initial lookup table.")
        gp.gc_map(mmli.par, None, dem_orig.par, dem_orig.dat,
                  dem.par, dem.dat, dem.lookup, dem_lat_ovs, dem_lon_ovs,
                  geo.sim_sar, geo.zenith, geo.orient, geo.inc, geo.psi,
                  geo.pix, geo.ls_map, 8, 2)
    else:
        log.info("Initial lookup table already created.")

    dem_segpent_width = dem["width"]
    dem_segpent_lines = dem["lines"]

    gp.pixel_area(mmli.par, dem.par, dem.dat, dem.lookup, geo.ls_map,
                  geo.inc, geo.sigpa0, geo.gamma0, 20)
    
    gp.create_diff_par(mmli.par, None, geo.diff_par, 1, 0)
    
    log.info("Refining lookup table.")

    if itr >= 1:
        log.info("ITERATING OFFSET REFINEMENT.")

        for ii in range(itr):
            log.info("ITERATION %d / %d" % (ii + 1, itr))

            geo.rm("diff_par")

            # copy previous lookup table
            dem.cp("lookup", dem.lookup_old)

            gp.create_diff_par(mmli.par, None, geo.diff_par, 1, 0)

            gp.offset_pwrm(geo.sigpa0, mmli.dat, geo.diff_par, geo.offs,
                           geo.ccp, rng_patch, azi_patch, geo.offsets, 2,
                           n_rng_off, n_azi_off, 0.1, 5, 0.8)

            gp.offset_fitm(geo.offs, geo.ccp, geo.diff_par, geo.coffs,
                           geo.coffsets, 0.1, npoly)

            # update previous lookup table
            gp.gc_map_fine(dem.lookup_old, dem_segpent_width, geo.diff_par,
                           dem.lookup, 1)

            # create new simulated ampliutides with the new lookup table
            gp.pixel_area(mmli.par, dem.par, dem.dat, dem.lookup, geo.ls_map,
                          geo.inc, geo.sigpa0, geo.gamma0, 20)

        # end for
        log.info("ITERATION DONE.")
    # end if
    
    hgt = gm.HGT(pth.join(geo_path, "dem.rdc"), mmli)
    
    self.meta.update({"geo": geo, "dem_orig": dem_orig, "dem": dem,
                      "hgt": hgt})

    self.meta["SLC_merged_nomaster"] = merged


def check_geocode(self):

    hgt, geo, dem = self.meta["hgt"], self.meta["geo"], self.meta["dem"]
    
    mrng = hgt.mli.rng()
    
    log.info("Geocoding DEM heights into image coordinates.")
    dem.geo2rdc(dem.dat, hgt.dat, mrng, nlines=hgt.mli.azi(), interp="sqr_dist")
    
    dem.gm.ras_extter("lookup")
    
    # TODO: make gm.ras_extter2
    log.info("Creating quicklook hgt.bmp file.")
    hgt.gm.ras_extter(m_per_cycle=500.0)
    
    geo.gm.ras_extter("gamma0")

    gp.dis2pwr(hgt.mli.dat, geo.gamma0, mrng, mrng)


def coreg_slcs(self):
    log.info("Starting COREG_SLCS")
    
    master, SLC, dates = self.meta["master"], self.meta["SLC_merged"], \
                         self.meta["dates"]
    
    mdate = master["S1SLC"].date
    
    output_dir, master_date = self.get_out_master()
    general = self.params.general
    
    pol       = general.get("pol", "vv")
    rng_looks = general.get("range_looks", 1)
    azi_looks = general.get("azimuth_looks", 4)

    coregp = self.params.coreg
    
    cc_thresh   = float(coregp.get("cc_thresh", 0.8))
    frac_thresh = float(coregp.get("fraction_thresh", 0.01))
    ph_std_thresh  = float(coregp.get("ph_stdev_thresh", 0.8))
    itmax       = float(coregp.get("itmax", 5))
    
    cleaning, flag1, poly1, poly2 = True, True, None, None
    
    hgt = self.meta["hgt"].dat
    
    coreg_dir = gm.mkdir(pth.join(output_dir, "coreg_out"))
    rmli_dir = gm.mkdir(pth.join(output_dir, "RMLI"))
    diff_dir = gm.mkdir(pth.join(output_dir, "IFG"))
    
    tpl_iw = pth.join(coreg_dir, "{date}_iw{iw}.{pol}.rslc")
    tpl_tab = pth.join(coreg_dir, "{date}.{pol}.RSLC_tab")
    fmt = "%Y%m%d"
    
    SLC_sort = sorted(SLC, key=lambda x: x.date.mean)
    midx = tuple(ii for ii, slc in enumerate(SLC_sort) if slc.date == mdate)[0]
    
    # number of slave images
    n_sar = len(SLC) - 1
    prev = None

    
    RSLC, RMLI = [], []
    
    if midx == 0:
        # master date at the start of SLCs
        itr = SLC_sort
    elif midx == n_sar:
        # master date at the end of SLCs
        itr = reversed(SLC_sort)
    else:
        raise NotImplementedError("Master date in the middle not implemented yet.")
        # master date in the "middle" of SLCs

        # from master date to the end
        range(master_idx + 1, n_sar)
        # from master date to the start
        range(master_idx - 1, -1, -1)
        
    
    log.info("Master: %s." % master["S1SLC"].tab)
    
    for ii, slc in enumerate(itr):
        if ii == midx:
            continue
        
        # log_coreg(ii, n_sar, master_par, parfile, prev)

        SLC_coreg = \
        gm.S1SLC.from_template(pol, slc.date, slc.IWs, tpl_tab=tpl_tab,
                               fmt=fmt, tpl=tpl_iw)

        gm.coreg(master, slc, SLC_coreg, hgt, rng_looks, azi_looks, poly1,
                 poly2, cc_thresh, frac_thresh, ph_std_thresh, cleaning,
                 flag1, prev, diff_dir)
        
        rslc = pth.join(coreg_dir, slc.date.date2str()) + ".rslc"
        rmli = gm.MLI(pth.join(rmli_dir, slc.date.date2str()) + ".rmli")
        
        SLC_coreg.mosaic(datfile=rslc, rng_looks=rng_looks,
                         azi_looks=azi_looks)
        
        SLC_coreg.slc.multi_look(rmli, rng_looks=rng_looks,
                                 azi_looks=azi_looks)
        
        RSLC.append(SLC_coreg)
        RMLI.append(rmli)
        
        rmli.gm.ras_extter()
        
            #gs.S1_coreg(mslc, slc, SLC_coreg, hgt, range_looks, azimuth_looks,
                    #poly1, poly2, cc_thresh, frac_thresh, std_thresh,
                    #cleaning, flag1, prev)

        prev = SLC_coreg
    
    self.meta.update({"RSLC": tuple(RSLC), "RMLI": tuple(RMLI)})



def deramp(self):
    log.info("Starting DERAMP_SLCS.")
    
    mslc = self.meta["master"]["S1SLC"]
    gen = self.params.general

    rng_looks = gen.get("range_looks", 1)
    azi_looks = gen.get("azimuth_looks", 4)
    output_dir = gen.get("output_dir", ".")
    
    deramp_dir = gm.mkdir(pth.join(output_dir, "deramp"))
    
    
    mslcd = gm.S1SLC.from_SLC(mslc, ".deramp")
    _slc = pth.join(deramp_dir, "%s.slc.deramp" % mslc.date.date2str())
    
    gp.S1_deramp_TOPS_reference(mslc.tab)
    mslcd.mosaic(datfile=_slc, rng_looks=rng_looks, azi_looks=azi_looks)
    
    self.meta["master"]["S1SLC_deramp"] = mslcd
    
    
    RSLC = self.meta["RSLC"]
    #RSLC = self.meta["RLSC"]
    
    deramped = [gm.S1SLC.from_SLC(slc, ".deramp") for slc in RSLC]
    
    for rslc, dslc in zip(RSLC, deramped):
        date = rslc.date.date2str()
        
        gp.S1_deramp_TOPS_slave(rslc.tab, date, mslc.tab, rng_looks,
                                azi_looks, 0)
        
        _slc = pth.join(deramp_dir, "%s.slc.deramp" % date)
        
        dslc.mosaic(datfile=_slc, rng_looks=rng_looks, azi_looks=azi_looks)
    
    deramped.append(mslcd)
    
    self.meta["RSLC_deramped"] = deramped
    self.meta["master_idx"] = -1


def base_plot(self):
    ipta_dir = self.params.general.get("ipta_dir", ".")
    gm.mkdir(ipta_dir)
    
    bperp = pth.join(ipta_dir, "bperp")
    itab = pth.join(ipta_dir, "itab")
    SLC_tab = pth.join(ipta_dir, "SLC_tab")
    
    ifg_sel = self.params.ifg_select
    
    bperp_lims = ifg_sel.get("bperp_lims", (0.0, 150.0))
    delta_T_lims = ifg_sel.get("delta_T_lims", (0.0, 15.0))
    
    
    deramps = tuple(rslcd.slc for rslcd in self.meta["RSLC_deramped"])
    
    gi.base_plot(self.meta["master_idx"], deramps, bperp_lims, delta_T_lims,
                 SLC_tab, bperp, itab)
    
    self.meta.update({"bperp": bperp, "itab": itab, "SLC_tab": SLC_tab,
                      "deramps": deramps})


def avg_mli(self):
    ipta_dir = self.params.general.get("ipta_dir", ".")
    dmli_dir = gm.mkdir(pth.join(ipta_dir, "DMLI"))

    gen = self.params.general

    rng_looks = gen.get("range_looks", 1)
    azi_looks = gen.get("azimuth_looks", 4)

    
    deramps = self.meta["deramps"]
    
    tpl = pth.join(dmli_dir, "%s.mli")
    
    mli = tuple(gm.MLI(tpl) % slc.date.date2str() for slc in deramps)
    
    for slc, mli in zip(deramps, mli):
        slc.multi_look(MLI, rng_looks=rng_looks, azi_looks=azi_looks)
    
    
    ave = gm.MLI(pth.join(ipta_dir, "avg.rmli"), parfile=mli[midx].par)
    
    imlist = pth.join(ipta_dir, "dmli_list")
    
    with open(imlist, "w") as f:
        f.write("%s\n" % "\n".join(MLI.dat for MLI in mli))
    
    gp.ave_image(imlist, ave.rng(), ave.dat)
    
    ave.gm.ras_extter()

    self.meta.update({"avg_mli": ave, "DMLI": mli})

_dtypes = {
    "FCOMPLEX": 0,
    "SCOMPLEX": 1,
    "FLOAT": 2,
    "INT": 3,
    "SHORT": 4,
    "BYTE": 5,
    "SUN": 6,
    "BMP": 6,
    "TIFF": 6
}

def candidate_list(self):
    ipta_dir = self.params.general.get("ipta_dir", ".")
    pt_select = self.params.pt_select
    
    sp_dir = gm.mkdir(pth.join(ipta_dir, "sp"))
    
    rng_spec_lk = pt_select.get("rng_spec_lk", 4)
    azi_spec_lk = pt_select.get("azi_spec_lk", 4)
    pwr_thresh = pt_select.get("pwr_thresh", 0)
    cc_thresh = pt_select.get("cc_thresh", 0.4)
    msr_thresh = pt_select.get("msr_thresh", 1.0)
    cc_lims = pt_select.get("cc_lims", (0.37, 1.0))
    msr_lims = pt_select.get("msr_lims", (0.5, 100.0))
    rng_ovr = pt_select.get("rng_ovr", 1)
    
    gp.mk_sp_all(self.meta["SLC_tab"], sp_dir, rng_spec_lk, azi_spec_lk,
                 int_thresh, cc_thresh, msr_thesh, rng_ovr, debug=True)
    
    sp = gm.Base(pth.join(sp_dir, "ave"), cc=".sp_cc", msr=".sp_msr",
                 ptmap="%s.%s" % ("ptmap", _default_image_fmt))
    
    width = self.meta["deramps"][0].rng()
    
    # TODO: Fix this!
    gm.single_class_mapping(sp.cc, cc_lim[0], cc_lim[1],
                            sp.msr, msr_lims[0], msr_lims[1],
                            width=width, avg_rng=1, avg_azi=1, ras_ext=sp.ptmap)
    
    pt2 = pth.join(ipta_dir, "pt2")
    pdata = gm.PointData(pth.join(ipta_dir, "pt"))
    
    
    gp.image2pt(sp.ptmap, width, pdata, 1, 1,
                _dtypes[pth.splitext(sp.ptmap)[1].upper()])
    
    amli = self.meta["avg_mli"]
    
    gp.gm.ras_ext_pt(pt2, None, amli.gm.ras_ext, "%s.%s" % (pt2, gm.ras_ext), 5, 1, 255, 255, 0, 3)
    
    merge(pt1, pt2, pdata.plist)
    
    self.meta["pdata"] = pdata


