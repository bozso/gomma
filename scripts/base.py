from sys import path
from os.path import join as pjoin

__all__ = ("gamma", "progs")

progs = pjoin("/home", "istvan", "progs")

path.append(pjoin(progs, "utils"))
from utils import cmd_line_prog

exe = pjoin(progs, "gamma", "bin", "gamma")
cmds = ("proc", "init", "batch", "move", "make", "stat")

gamma = cmd_line_prog(exe, *cmds, prefix="")
