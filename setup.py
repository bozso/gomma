import utils as ut
import os.path as pth
import glob

#import argparse as ap


flags = "-ldflags '-s -w'"


root = pth.dirname(pth.abspath(__file__))

def generate_ninja():
    src = glob.glob(pth.join(root, "src", "*.go"))
    main = pth.join(root, "bin", "gamma")
    
    cmd = "go build %s -o ${out} ${in}"
    
    for path in {"src", "bin"}:
        n = ut.Ninja.in_path(path)
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
