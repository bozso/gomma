from functools import partial
from glob import iglob
from sys import path
from os.path import join as pjoin

import json

__all__ = ("gamma", "progs", "Project")

progs = pjoin("/home", "istvan", "progs")

path.append(pjoin(progs, "utils"))
import utils

exe = pjoin(progs, "gamma", "bin", "gamma")
cmds = ("select", "import", "batch", "move", "make", "stat", "like")

gamma = utils.cmd_line_prog(exe, *cmds)

class Project(object):
    def __init__(self, *args, **kwargs):
        self.general = kwargs
    
    def select(self, path, *args, **kwargs):
        datas = ["-d" + path for path in iglob(pjoin(path, "*.zip"))]
        gamma.select(" ".join(datas), *args, **self.general, **kwargs)
    
    def data_import(self, *args, **kwargs):
        getattr(gamma, "import")(*args, **self.general, **kwargs)
    

class DataFile(dict):
    __slots__ = ("metafile",)

    def __init__(self, path):
        self.metafile = path
        with open(path, "r") as f:
            self.update(json.load(f))
    
    def like(self, name=None, **kwargs):
        if name is None:
            name = utils.get_tmp()
        
        kwargs["in"] = self.metafile
        gamma.like(out=name, **kwargs)
        
        return DataFile(name)
    
    def stat(self, **kwargs):
        return gamma.stat(self.metafile, **kwargs)
    

    
class SLC(DataFile):
    def SplitInterferometry(self):
        pass

class Lookup(DataFile):
    def geocode(self, mode, infile, outfile=None, like=None, **kwargs):
        kwargs["infile"] = infile
        
        if like is not None:
            outfile = like.like(**kwargs)
            Lookup   string `cli:"*l,lookup" usage:"Lookup table file"`
        
        kwargs["outfile"] = outfile
        kwargs["lookup"] = self.metafile

        gamma.geocode(**kwargs)
    
    def radar2geo(self, **kwargs):
        kwargs["mode"] = "togeo"
        
        return self.geocode(**kwargs)

    def geo2radar(self, **kwargs):
        kwargs["mode"] = "toradar"
        
        return self.geocode(**kwargs)
