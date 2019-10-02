from utils import Ninja
from glob import glob

sources = glob("../gamma/*.go")

n = Ninja(open("build.ninja", "w"))

server = "istvan@zafir.ggki.hu"
target = "/home/istvan/packages/usr/bin"

main = "gamma"

n.rule("go", "go build ${in}", "Build executable")
n.newline()

n.rule("deploy", 'lxterminal -e "scp ${in} %s:%s"' % (server, target),
       "Deploy to server")
n.newline()




n.build(main, "go", main + ".go", implicit=sources)
n.newline()

n.build("DEPLOY", "deploy", main)
n.newline()
