import os.path as pth
import functools as ft
import subprocess as sub
import shlex

from glob import iglob
from sys import path
from os.path import join as pjoin

import json

__all__ = ("gamma", "progs", "Project", "DataFile", "SLC", "Lookup")

progs = pjoin("/home", "istvan", "progs")

path.append(pjoin(progs, "utils"))
import utils


class Enforcer(object):
    __slots__ = ("exc",)
    
    def __init__(self, exc):
        self.exc = exc
    
    def __call__(self, cond, *args, **kwargs): 
        print(type(self.exc))
        if not cond:
            raise self.exc(*args, **kwargs)


@ft.lru_cache()
def enforcer(exc):
    return Enforcer(exc)


class Command(object):
    __slots__ = ("path", "tpl", "subcommands")
    
    error_tpl = ("\nNon zero returncode from command: \n'{}'\n"
                 "\nOUTPUT OF THE COMMAND: \n\n{}\nRETURNCODE was: {}")
    
    def __init__(self, *args, **kwargs):
        self.path = pth.join(*args)
        self.tpl = "%s%%s%s%%s" % (
            kwargs.get("prefix", "--"),
            kwargs.get("sep", "=")
        )
        self.subcommands = kwargs.get("subcommands", None)
    
        
    def __call__(self, *args, **kwargs):
        debug = kwargs.pop("_debug_", False)
        
        Cmd = self.path
        
        if len(args) > 0:
            Cmd += " %s" % " ".join(args)
        
        tpl = self.tpl
        
        if len(kwargs) > 0:
            Cmd += " %s" % " ".join(tpl % (key, val)
                                        for key, val in kwargs.items()
                                        if val is not None)
        
        if debug:
            print("Command is '%s'" % Cmd)
            return
        
        try:
            proc = sub.check_output(shlex.split(Cmd), stderr=sub.STDOUT)
        except sub.CalledProcessError as e:
            raise RuntimeError(
                self.error_tpl.format(Cmd, e.output.decode(),
                    e.returncode)
            )
        
        
        return proc
    
    def subcmd(self, cmd, *args, **kwargs):
        err = enforce(TypeError)
        
        err(self.subcommands is not None, 
            "This command line executable does not support subcommands"
        )
        
        err(cmd in self.subcommands,
            "Subcommand '%s' is not supported. Available commands: %s" %
            (cmd, self.subcommands)
        )
        
        return self(" %s" % cmd, *args, **kwargs)


exe = pjoin(progs, "gamma", "bin", "gamma")
cmds = {"select", "import", "batch", "move", "make", "stat", "like"}
    
gamma = Command(exe, subcommands=cmds, prefix="-")

class Project(object):
    def __init__(self, *args, **kwargs):
        self.general = kwargs
    
    def select(self, path, *args, **kwargs):
        datas = ["-d" + path for path in iglob(pjoin(path, "*.zip"))]
        gamma.subcmd("select", " ".join(datas),
            *args, **self.general, **kwargs)
    
    def data_import(self, *args, **kwargs):
        gamma.subcmd("import", *args, **self.general, **kwargs)
    

class DataFile(dict):
    __slots__ = ("metafile",)

    def __init__(self, path):
        self.metafile = path
        with open(path, "r") as f:
            self.update(json.load(f))
    
    @classmethod
    def make(cls, meta, dat=None, par=None, ext=None, dtype="Unknown"):
        if dat is None:
            dat = utils.tmp_file() + ".dat"
        
        gamma.make(meta=meta, dat=dat, par=par, parExt=ext,
            dtype=dtype)
        
        return cls(meta)
    
    def like(self, name=None, **kwargs):
        if name is None:
            name = utils.tmp_file()
        
        kwargs["in"] = self.metafile
        gamma.like(out=name, **kwargs)
        
        return DataFile(name)
    
    def move(self, dirPath):
        gamma.move(meta=self.meta, out=dirPath)
        self.meta = path.join(dirPath, self.meta)
    
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
        
        kwargs["outfile"] = outfile
        kwargs["lookup"] = self.metafile

        gamma.geocode(**kwargs)
    
    def radar2geo(self, **kwargs):
        kwargs["mode"] = "togeo"
        
        return self.geocode(**kwargs)

    def geo2radar(self, **kwargs):
        kwargs["mode"] = "toradar"
        
        return self.geocode(**kwargs)
