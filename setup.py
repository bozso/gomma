import utils as ut
import os.path as path
import glob

#import argparse as ap


flags = "-ldflags '-s -w'"

root = path.dirname(path.abspath(__file__))

def sources(*args):
    return glob.glob(path.join(*args))

def generate_ninja():
    src_path = path.join(root, "gamma")
    src = sources(src_path, "*.go")
    
    subdirs = {path.join(src_path, elem)
        for elem in {
            "base",
            "command_line",
            "common",
            "data",
            "dem",
            "geo",
            "interferogram",
            "plot",
            "sentinel1",
            "utils",
            path.join("utils", "params"),
        }
    }
    
    for sdir in subdirs:
        src += sources(src_path, sdir, "*.go")
    
    main = path.join(root, "gamma")
    
    cmd = "go build %s -o ${out} ${in}"
    
    subdirs |= {src_path, "bin"}
    
    for sdir in subdirs:
        n = ut.Ninja.in_path(sdir)
        n.rule("go", cmd % flags, "Build executable.")
        n.newline()
    
        n.build(main, "go", main + ".go", implicit=src)
        n.newline()
    
    #n = Ninja.in_path("bin")
    #n.rule("go", "go build ${in}", "Build executable.")
    #n.newline()
    
    #n.build("go", main)
    #n.newline()
    
    
    
def main():
    #if "ninja" in sys.args:
    generate_ninja()
    
    
    
if __name__ == "__main__":
    main()
