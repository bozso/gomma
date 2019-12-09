from sys import path

path.append("/home/istvan/progs/utils")

from utils import Ninja
from glob import glob

sources = glob("../src/*.go")

n = Ninja(open("build.ninja", "w"))

main = "gamma"

n.rule("go", "go build ${in}", "Build executable")
n.newline()

n.build(main, "go", main + ".go", implicit=sources)
n.newline()
