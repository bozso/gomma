from functools import partial
from glob import iglob
from sys import path
from os.path import join as pjoin

__all__ = ("gamma", "progs", "Project")

progs = pjoin("/home", "istvan", "progs")

path.append(pjoin(progs, "utils"))
from utils import cmd_line_prog

exe = pjoin(progs, "gamma", "bin", "gamma")
cmds = ("select", "import", "batch", "move", "make", "stat")

gamma = cmd_line_prog(exe, *cmds)

class Project(object):
    def __init__(self, *args, **kwargs):
        self.general = kwargs
    
    def select(self, path, *args, **kwargs):
        datas = ["-d" + path for path in iglob(pjoin(path, "*.zip"))]
        gamma.select(" ".join(datas), *args, **self.general, **kwargs)
    
    def data_import(self, *args, **kwargs):
        getattr(gamma, "import")(*args, **self.general, **kwargs)
    
