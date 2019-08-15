import logging
import os
import os.path as pth

from datetime import datetime
from glob import iglob
from pickle import dump, load
from collections import namedtuple
from pprint import pprint
from shutil import copy
from sys import stdout
from json import load as jload, dump as jdump, JSONEncoder
from functools import partial
from itertools import tee

try:
    from configparser import SafeConfigParser
except ImportError:
    from ConfigParser import SafeConfigParser



from utils import *
import gamma as gm
import gamma.sentinel1 as s1


__all__ = (
    "Processing",
    "is_debug"
)


ProcStep = namedtuple("ProcStep", "fun opt")


def typedval(obj):
    return {
        "type": type(obj).__name__,
        "value": obj
    }


gp = gm.gp

log = logging.getLogger("gamma.steps")

debug = False

def is_debug():
    return debug


# *********************
# * Utility functions *
# *********************


def delim(msg, sym="*", width=80):
    msg = str(msg)
    msg = "%s %s %s" % (sym, msg, sym)

    syms = sym * len(msg)
    
    tpl = "\n{{:^{w}}}\n{{:^{w}}}\n{{:^{w}}}\n".format(w=width)
    
    print(tpl.format(syms, msg, syms))



trues  = frozenset({"true", "on", "1"})
falses = frozenset({"false", "off", "0"})


class Config(dict):
    def int(self, key, default):
        return int(self.get(key, default))
    
    def float(self, key, default):
        return float(self.get(key, default))
   
    def bool(self, key, default):
        val = self.get(key, default)
        
        if val is None:
            return False
        
        if isinstance(val, str):
            val = val.lower()
            
            if val in trues:
                return True
            elif val in falses:
                return False

            raise ValueError('Could not convert "%s" to boolean!' % val)


class FileEncoder(JSONEncoder):
    def default(self, obj):
        if hasattr(obj, "to_json"):
            return JSONEncoder.default(self, obj.to_json())
        elif hasattr(obj, "__save__"):
            return {
                key: getattr(obj, key)
                for key in getattr(obj, "__save__")
            }
        return JSONEncoder.default(self, obj)


def save(obj, path):
    with open(path, "w") as f:
        jdump(obj, f, indent=4, separators=(",", ": "),
              cls=FileEncoder)
    

class Save(object):
    @staticmethod
    def load_file(path):
        with open(path, "r") as f:
            return jload(f)
    
    @classmethod
    def from_file(cls, path):
        return cls(Save.load_file(path))
    
    
    def save(self, path):
        save(self, path)


class Meta(dict, Save):
    pass
    



class Processing(object):
    cache_dirs = frozenset({"ifg", "slc", "rslc", "extracted"})
    
    ftypes = {
        name: getattr(gm, name).from_json
        for name in {"S1SLC", "SLC", "MLI", "S1Zip", "DEM", "Geocode"}
    }
    
    _default_log_format = "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
    
    steps = frozenset({"select", "load_slc", "merge",  "make_mli", "make_rmli",
                       "check_geocode"})
    
    
    def __init__(self, args):
        self.args = args
        self.params = SafeConfigParser()
        
        log.info("File containing processing parameters: %s"
                 % (args.conf_file))
        
        self.params.read(args.conf_file)
        
        steps = self.params.sections()
        steps.remove("general")
        steps = set(steps)
        steps.update(Processing.steps)
        
        self.steps = steps
        
        self.metafile = self.params.get("general", "metafile")
        cache_path = self.params.get("general", "CACHE_PATH")
        
        if cache_path is None:
            cache_path = gm.settings["cache_default_path"]
        
        for path in self.cache_dirs:
            mkdir(cache_path, path)
        
        self.cache_path = cache_path
        self.caches = type("CacheDirectories", (object,), 
                            {val: pth.join(cache_path, val)
                             for val in self.cache_dirs})
        self.extractor = \
        partial(gm.extract, outpath=pth.join(self.caches.extracted))
        self.is_extracted = partial(isfile, self.caches.extracted)
        
        if pth.isfile(self.metafile):
            self.meta = Meta.from_file(self.metafile)
        else:
            self.meta = Meta(dirs={}, lists={})
            
        
        self.optional_steps = set(
            step for step in self.steps
            if self.is_optional(step)
        )
        
        
        self.required_steps = set(
            step for step in self.steps
            if step not in self.optional_steps
        )

    
    def parse_steps(self):
        args = self.args
        step = args.step
        steps = self.steps
        
        if step is not None:
            if step not in steps:
                raise ValueError('Single step "%s" is not a valid processing '
                                 'step. Choose from: %s'
                                 % (step, ", ".join(steps)))
            
            log.debug('Single step "%s" is executed.' % step)
            return [step]
        else:
            start = args.start
            stop  = args.stop
            
            if start not in steps:
                raise ValueError('Step "%s" is not a valid processing '
                                 'step. Choose from: %s'
                                 % (start, ", ".join(steps)))
    
            if stop not in _steps:
                raise ValueError('Step "%s" is not a valid processing '
                                 'step. Choose from: %s'
                                 % (stop, ", ".join(steps)))
    
            log.debug('Steps from "%s" to "%s" will be executed.'
                       % (start, stop))
            
            first = steps.index(start)
            last  = steps.index(stop)
            return steps[first:last + 1]


    def show_steps(self):
        print("\nProcessing steps: %s\nOptional steps: %s\n"
              % (", ".join(step for step in self.required_steps),
                 ", ".join(step for step in self.optional_steps)))
    
    
    def is_optional(self, step):
        return self.params.has_option(step, "optional") and \
               self.params.getboolean(step, "optional")
    
    
    def step_fun(self, step):
        return ProcStep(getattr(self, step), self.is_optional(step))

    
    def run_steps(self):
        args = self.args
        Processing.setup_log("gamma", filename=args.logfile,
                             loglevel=args.loglevel)
        
        if args.show_steps:
            self.show_steps()
            return 0
        
        log.info('File containing processing parameters: "%s"'
                 % (args.conf_file))
        
        if args.info:
            pprint(self.params)
            pprint(self.meta)
            return 0
        
        steps = self.parse_steps()
        
        optional_steps = not args.skip_optional
        
        if not optional_steps:
            log.info("Skipping optional steps.")
        
        
        for step in steps:
            step_fun = self.step_fun(step)
            
            try:
                if step_fun.opt or optional_steps:
                    ustep = step.upper()
                    
                    delim("Starting step: %s" % ustep)
                    step_fun.fun()
                    delim("Finished step: %s" % ustep)
            finally:
                self.meta.save(self.metafile)

    
    def filename(self, name):
        return pth.join(self.params.get("general", "output_dir"),
                        "%s.json" % name)
    
    def save(self, name, *args, **kwargs):
        path = self.filename(name)
        
        assert "list" not in kwargs
        
        to_save = {"list": [typedval(elem) for elem in args]}
        
        if len(kwargs) > 0:
            to_save.update({key: typedval(value)
                            for key, value in kwargs.items()})
        
        save(to_save, path)
    
    
    def load(self, name):
        load = Save.load_file(self.filename(name))
        
        _list = load.pop("list")
        ret = {}
        
        print(_list)
        
        if _list is not None:
            ret["list"] = [self.ftypes[elem["type"]](elem["value"])
                           for elem in _list]
        
        if len(load) > 0:
            ret.update({key: self.ftypes[value["type"]](value)
                        for key, value in load.items()})
        
        return ret
        
            
    def load_list(self, name):
        return self.load(name)["list"]
    
    
    def update(self, name, **kwargs):
        load = self.load_file(name)
        
        load.update(kwargs)
        
        save(self.list(name), load)
        
        
    def select_date(self, name, date):
        if not self.is_list(name):
            return None
        
        return tuple(elem for elem in self.inlist(name)
                     if elem.datestr() == date)[0]
    
    def select_master(self, name):
        return self.select_date(name, self.meta.get("master_date"))
    
    
    def master_date(self):
        master_date = self.meta.get("master_date")
        
        if master_date is None:
            raise ValueError("master_date is not defined.")
        
        return master_date

    
    @staticmethod
    def setup_log(logger_name, filename=None, formatter=None,
                  loglevel="debug"):
        
        logger = logging.getLogger(logger_name)
        
        level = getattr(logging, loglevel.upper(), None)
        
        if level == logging.DEBUG:
            debig = True
        
        if not isinstance(level, int):
            raise ValueError("Invalid log level: {}".format(loglevel))
        
        logger.setLevel(level)
        
        if formatter is None:
            formatter = Processing._default_log_format
        
        form = logging.Formatter(formatter, datefmt="%Y.%m.%d %H:%M:%S")
        
        
        if filename is not None:
            fh = logging.FileHandler(filename)
            fh.setFormatter(form)
            logger.addHandler(fh)
        
        consoleHandler = logging.StreamHandler()
        consoleHandler.setFormatter(form)
        
        logger.addHandler(consoleHandler)    
    
        return logger

    
    def section(self, name):
        return Config(self.params.items(name))
    
    
    #                            ********************
    #                            * Processing steps *
    #                            ********************
    
    
    def select(self):
        general, select = self.section("general"), self.section("select")
        
        slc_data = general.get("slc_data")
        
        if slc_data is None:
            raise ValueError('Parameter "slc_data" not defined.')

        
        master_date, output_dir, pol = \
        general.get("master_date"), general.get("output_dir", "."), \
        general.get("pol")

        date_start, date_stop, check_zips = \
        select.get("date_start"), select.get("date_stop"), \
        select.bool("check_zips", False)
        
        
        SLC = Seq(map(gm.S1Zip, ls(slc_data, "S1*_IW_SLC*.zip")))
        
        if date_start is not None and date_stop is not None:
            date_start = datetime.strptime(date_start, "%Y%m%d")
            date_stop = datetime.strptime(date_stop, "%Y%m%d")
            
            filt = lambda x: x.date.start > date_start and \
                          x.date.stop < date_stop
        
        
        elif date_start is not None:
            date_start = datetime.strptime(date_start, "%Y%m%d")
            
            filt = lambda x: x.date.start > date_start

        elif date_stop is not None:
            date_stop = datetime.strptime(date_stop, "%Y%m%d")
            
            filt = lambda x: x.date.stop < date_stop
        else:
            filt = lambda x: x
        
        
        if check_zips:
            log("Checking integrity of zipfiles.")
            filt = lambda x: x.test_zip() and filt
        
        
        zips, SLC = SLC.filter(filt).tee(2)
        
        extracted = self.caches.extracted
        extract_path = partial(pth.join, extracted)
        
        
        names = ("annot", "quicklook")
        
        extract = (SLC.map(gm.make_extract, pol=pol, names=names)
                      .map(gm.extract, filt_fun=partial(isfile, extracted),
                          outpath=extracted)
                      .chain()
                      .collect())
        
        zips.map(gm.S1Zip.burst_info, namelist=extract, pol=pol).collect()
        
        exit()
        
        if master_date is None:
            log.info("No master_date defined, using first date.")
            
            master_slc = SLC.sorted(key=lambda x: x.date.center)[0]
            master_date = master_slc.date.date2str()
            
            self.meta["master_date"] = master_date
        else:
            master_date = general["master_date"]
            
            log.info("Master date already defined: %s", master_date)
            
            master_slc = [slc for slc in SLC
                          if slc.datestr() == master_date][0]


        log.info("Selected master date is %s" % master_date)
        
        self.save("zipfiles", *SLC.collect())
        

    def load_slc(self):
        if gm.ScanSAR:
            copy_fun = getattr(gp, "SLC_copy_ScanSAR")
        else:
            copy_fun = getattr(gp, "SLC_copy_S1_TOPS")

        general = self.section("general")
        
        master_date = self.master_date()
        pol = general.get("pol")
        
        dir_uncrop = self.dir("uncrop")
        dir_crop   = self.dir("crop")
        SLC = self.load_list("s1zip")
        
        uncrop = []
        
        if 0:
            for slc in SLC:
                _uncrop = \
                gm.S1SLC.from_template(slc.date, slc.burst_nums, pol,
                                       fmt=None, dirpath=dir_uncrop)
                
                for IW in _uncrop.IWs:
                    if IW is not None:
                        slc.extract_IW(pol, IW)
                
                uncrop.append(_uncrop)
            self.outlist("uncrop", uncrop)
        
        uncrop = self.inlist("uncrop")
        
        burst_tab = gm.get_tmp()
        
        crop = []
        
        for uc, slc in zip(uncrop,  SLC):
            with open(burst_tab, "w") as f:
                f.write("\n".join("%d %d" % (burst[0], burst[1])
                        for burst in slc.burst_nums if burst is not None) + "\n")
            
            _crop = uc.make_other(dirpath=dir_crop)
            
            copy_fun(uc.tab, _crop.tab, burst_tab)
            crop.append(_crop)
    
    
        self.save("crop", *crop)

    
    def merge(self):

        if gm.ScanSAR:
            merge_fun = getattr(gp, "SLC_cat_ScanSAR")
        else:
            merge_fun = getattr(gp, "SLC_cat_S1_TOPS")

        
        general = self.section("general")
        
        
        output_dir = general.get("output_dir", ".")
        pol        = general.get("pol", "vv")
        slc_dir    = self.dir("merged")

        merged, used_SLC = [], []
        SLC = self.load_list("crop")
        
        
        for SLC1 in SLC:
            if SLC1 in used_SLC:
                continue
            
            date1 = SLC1.date(start_stop=True)
            
            date1str = date1.date2str()
            
            log.info("Processing date %s." % date1str)
            
            SLC2 = search_pair(SLC1, SLC, used_SLC)
            SLC3 = SLC1.make_other(dirpath=slc_dir)
            
            if SLC2 is not None:
                log.info("Merging %s with %s." % (SLC1.tab, SLC2.tab))
                date2 = SLC2.date(start_stop=True)
                
                if date1.center > date2.center:
                    gp.SLC_cat_S1_TOPS(SLC2.tab, SLC1.tab, SLC3.tab)
                    merge_fun(SLC2.tab, SLC1.tab, SLC3.tab)
                else:
                    merge_fun(SLC1.tab, SLC2.tab, SLC3.tab)
    
                used_SLC.append(SLC2)
            else:
                log.info("No need for merge. Copying %s." % SLC1.tab)
                SLC1.cp(SLC3)
    
            # endif
            merged.append(SLC3)
        # endfor
        
        self.save("merged", merged)
        
        # CLEANUP
        #gm.rm("*.SAFE", "*.SLC_tab", "*iw*", "*.slc*")


    def make_mli(self):
        
        general = self.section("general")
        
        output_dir    = general.get("output_dir", ".")
        range_looks   = general.int("range_looks", 1)
        azimuth_looks = general.int("azimuth_looks", 4)
        pol           = general.get("pol")
        
        mli_dir = self.dir("mli")
        
        tpl = pth.join(mli_dir, "%s.%s.mli")
        
        MLI = []
        
        
        for slc in self.load_list("merged"):
            mli = gm.MLI(datfile=tpl % (slc.datestr(), pol))
            
            slc.multi_look(mli, range_looks, azimuth_looks)
            mli.raster()
            
            MLI.append(mli)
        
        
        self.save("mli", MLI)

    
    def mosaic_tops(self):
        
        general = self.section("general")
        
        output_dir      = general.get("output_dir", ".")
        range_looks     = general.int("range_looks", 1)
        azimuth_looks   = general.int("azimuth_looks", 4)
        pol             = general.get("pol", "vv")
        
        
        SLC = []
        
        for s1slc in self.load_list("merged"):
            dat = pth.join(s1slc.IW[0].dat.split(".")[-2], ".dat")
            slc = s1slc.mosaic(rng_looks=rng_looks, azi_looks=azi_looks,
                               datfile=dat)
            
            SLC.append(slc)
        
        self.save("slc_mosaic", slc)

    
    def check_ionoshpere(self):
        
        output_dir = self.params.get("general", "output_dir")
        
        
        if not self.is_list("merged"):
            self.mosaic_tops()
        
        check_iono = self.params.section("check_ionoshpere")
        
        rng_win = check_iono.int("rng_win", 256)
        azi_win = check_iono.int("azi_win", 256)
        thresh  = check_iono.float("iono_thresh", 0.1)
        
        rng_step = check_iono.int("rng_step")
        azi_step = check_iono.int("azi_step")
        
        raise NotImplementedError("This processing step needs to be reworked.")
        
        # SLC[0].check_ionoshpere(rng_win=rng_win, azi_win=azi_win, thresh=thresh,
        #                         rng_step=rng_step, azi_step=azi_step)
        
        

    def make_rmli(self):
        
        general = self.section("general")
        
        output_dir    = general.get("output_dir", ".")
        range_looks   = general.int("range_looks", 1)
        azimuth_looks = general.int("azimuth_looks", 4)
        pol           = general.get("pol")
        
        mli_dir = self.dir("rmli")
        tpl = pth.join(mli_dir, "%s.%s.mli")
        
        datestr = (date.date2str() for date in self.meta["dates"])
        
        RMLI = tuple(MLI(tpl % (date, pol)) for date in datestr)

        for rmli, rslc in zip(MLI, RSLC):
            rslc.multi_look(rmli, rng_looks=range_looks,
                            azi_looks=azimuth_looks)
            rmli.raster()
        
        self.save("RMLI", RMLI)


    def geocode(self):
        general = self.section("general")
        
        rng_looks = general.int("range_looks", 1)
        azi_looks = general.int("azimuth_looks", 4)
        
        geoc = self.section("geocode")
        
        master = self.load("master")
        m_s1slc = master["s1slc"]
        
        
        if "slc" not in master:
            m_slc = m_s1slc.mosaic(rng_looks=rng_looks, azi_looks=azi_looks,
                                   datfile=\
                                   pth.join(self.dir("SLC"),
                                            "%s.slc" % m_s1slc.datestr()))
            
            self.update("master", slc=m_slc)
        else:
            m_slc = master["slc"]
        
        
        if "mli" not in master:
            m_mli = mslc.multi_look(rng_looks=rng_looks, azi_looks=azi_looks)
            
            self.update("master", mli=m_mli)
        else:
            m_mli = master["mli"]
        
        self.save("geocode",
                  **im.geocode(geoc, m_slc, m_mli,
                               rng_looks=rng_looks,
                               azi_looks=azi_looks, 
                               out_dir=output_dir))


    def check_geocode(self):
        geoc = self.load("geocode")
        
        geo, dem = geoc["geo"], geoc["dem"]
        
        mrng = geo.mli.rng()
        
        log.info("Geocoding DEM heights into image coordinates.")
        dem.geo2rdc(dem.dat, geo.hgt, mrng, nlines=hgt.mli.azi(),
                    interp="sqr_dist")
        
        dem.raster("lookup")
        
        # TODO: make gm.raster2
        log.info("Creating quicklook hgt.bmp file.")
        hgt.raster(m_per_cycle=500.0)
        
        # geo.raster("gamma0")

        # gp.dis2pwr(hgt.mli.dat, geo.gamma0, mrng, mrng)


    def coreg(self):
        log.info("Starting COREG_SLCS")
        
        general = self.section("general")
        
        pol       = general.get("pol", "vv")
        rng_looks = general.int("range_looks", 1)
        azi_looks = general.int("azimuth_looks", 4)

        coreg = self.section("coreg")
        
        cc_thresh      = coreg.float("cc_thresh", 0.8)
        frac_thresh    = coreg.float("fraction_thresh", 0.01)
        ph_std_thresh  = coreg.float("ph_stdev_thresh", 0.8)
        itmax          = coreg.int("itmax", 5)
        
        cleaning, flag1, poly1, poly2 = True, True, None, None
        
        
        if self.is_list("geocode"):
            hgt = self.load["geocode"]["geo"].hgt
        else:
            hgt = None
        
        coreg_dir = self.dir("coreg_out")
        rmli_dir  = self.dir("rmli")
        diff_dir  = self.dir("ifg")
        
        # tpl_iw = pth.join(coreg_dir, "{date}_iw{iw}.{pol}.rslc")
        # tpl_tab = pth.join(coreg_dir, "{date}.{pol}.RSLC_tab")
        # fmt = "%Y%m%d"
        
        SLC = self.load_list("merged")
        
        SLC_sort = sorted(SLC, key=lambda x: x.date(start_stop=True).center)
        midx = tuple(ii for ii, slc in enumerate(SLC_sort)
                     if slc.datestr() == master_date)[0]
        
        # number of slave images
        n_sar = len(SLC) - 1
        prev = None
        
        m_s1slc = SLC_sort[midx]
        
        m_slc = m_s1slc.mosaic(rng_looks=rng_looks, azi_looks=azi_looks,
                               datfile=\
                               pth.join(self.dir("SLC"),
                                        "%s.slc" % m_s1slc.datestr()))
        
        m_mli = mslc.multi_look(rng_looks=rng_looks, azi_looks=azi_looks)
        
        master = {
            "s1slc": m_s1slc,
            "slc": m_slc,
            "mli": m_mli
        }
        
        
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
        
        
        for ii, slc in enumerate(itr):
            if ii == midx:
                continue
            
            # log_coreg(ii, n_sar, master_par, parfile, prev)

            SLC_coreg = slc.make_other(dirpath=coreg_dir)
            # gm.S1SLC.from_template(pol, slc.date(start_stop=True), slc.IWs,
            #                        tpl_tab=tpl_tab, fmt=fmt, tpl=tpl_iw)

            gm.S1_coreg(master, slc, SLC_coreg, hgt, rng_looks, azi_looks,
                        poly1, poly2, cc_thresh, frac_thresh, ph_std_thresh,
                        cleaning, flag1, prev, diff_dir)
            
            rslc = pth.join(coreg_dir, slc.date.date2str()) + ".rslc"
            rmli = gm.MLI(pth.join(rmli_dir, slc.date.date2str()) + ".rmli")
            
            # SLC_coreg.mosaic(datfile=rslc, rng_looks=rng_looks,
            #                  azi_looks=azi_looks)
            
            # SLC_coreg.slc.multi_look(rmli, rng_looks=rng_looks,
            #                          azi_looks=azi_looks)
            
            # RSLC.append(SLC_coreg)
            # RMLI.append(rmli)
            # 
            # rmli.gm.ras_extter()
            
                #gs.S1_coreg(mslc, slc, SLC_coreg, hgt, range_looks, azimuth_looks,
                        #poly1, poly2, cc_thresh, frac_thresh, std_thresh,
                        #cleaning, flag1, prev)

            prev = SLC_coreg
            RSLC.append(rslc_coreg)
            RMLI.append(rmli)
        
        self.save("rslc", *RSLC)
        self.save("rmli", *RMLI)
        self.save("master", master)

    
    
    def deramp(self):
        mslc, SLC = self.load("master")["slc"], self.load_list("rslc")
        gen = self.section("general")

        rng_looks = gen.int("range_looks", 1)
        azi_looks = gen.int("azimuth_looks", 4)
        output_dir = gen.get("output_dir", ".")
        
        mslc_d = mslc.deramp(master=True)
        
        SLC_d = tuple(slc.deramp(master=mslc) for slc in SLC)
        
        self.save("deramp", *SLC_d)
        self.update("master", deramp=mslc_d)

    
    
    # IPTA processing from here ?

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

        
    def candidate_list(self):
        ipta_dir = self.params.general.get("ipta_dir", ".")
        pt_select = self.section("pt_select")
        
        sp_dir = gm.mkdir(pth.join(ipta_dir, "sp"))
        
        rng_spec_lk = pt_select.int("rng_spec_lk", 4)
        azi_spec_lk = pt_select.int("azi_spec_lk", 4)
        pwr_thresh = pt_select.float("pwr_thresh", 0)
        cc_thresh = pt_select.float("cc_thresh", 0.4)
        msr_thresh = pt_select.float("msr_thresh", 1.0)
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



def search_pair(slc1, SLCs, used_SLCs):
    
    date1 = slc1.date(start_stop=True)
    
    for slc2 in SLCs:
        date2 = slc2.date(start_stop=True)
        if date1.date2str() == date2.date2str() \
        and date1.center != date2.center and slc2 not in used_SLCs:
            return slc2

    return None
