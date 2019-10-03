from utils import Ninja
from glob import glob

sources = glob("../gamma/*.go")

n = Ninja(open("build.ninja", "w"))

main = "gamma"

n.rule("go", "go build ${in}", "Build executable")
n.newline()

n.build(main, "go", main + ".go", implicit=sources)
n.newline()
