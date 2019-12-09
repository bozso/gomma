from utils import Ninja
from glob import glob

sources = glob("*.go")

n = Ninja(open("build.ninja", "w"))

n.rule("lint", "golangci-lint run", "Run golinter")
n.newline()
n.build("LINT", "lint", sources)
n.newline()
