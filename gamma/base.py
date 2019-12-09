from sys import path
from os.path import join as pjoin

__all__ = ("gamma", "progs")

progs = pjoin("/home", "istvan", "progs")

path.append(pjoin(progs, "utils"))
from utils import cmd_line_prog

exe = pjoin(progs, "gamma", "bin", "gamma")
cmds = ("select", "import", "batch", "move", "make", "stat")

gamma = cmd_line_prog(exe, *cmds)
_proc_steps = {"select", "import",}


def proc_step(self, name, *args, **kwargs):
    return getattr(gamma, name)(*args, **self.general, **kwargs)

proc_steps = type("ProcSteps", (object,),
    {key: partial(proc_step, name=key) for key in _proc_steps}
)


class Project(proc_steps):
    def __init__(self, *args, **kwargs):
        self.general = kwargs


