import utils as ut
import os.path as path
import glob

root = path.dirname(path.abspath(__file__))

class Project(object):
    flags = "-ldflags '-s -w'"
    
    
    @staticmethod
    def sources(*args):
        return glob.glob(path.join(*args))

    def generate_ninja(self):
        # src_path = path.join(root, "gamma")
        # src = sources(src_path, "*.go")
        
        subdirs = {path.join(root, elem)
            for elem in {
                "base",
                "cli",
                "common",
                "date",
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
        
        
        src = sum(
            (
                self.sources(root, sdir, "*.go")
                for sdir in subdirs
            ),
            []
        )

        main = path.join(root, "gamma")
        
        cmd = "go build %s -o ${out} ${in}"
        
        ninja = path.join(root, "build.ninja")
        
        n = ut.Ninja.in_path(root)
        n.rule("go", cmd % self.flags, "Build executable.")
        n.newline()
    
        n.build(main, "go", main + ".go", implicit=src)
        n.newline()
        
        for sdir in subdirs:
            n = ut.Ninja.in_path(sdir)
            n.subninja(ninja)
    
    
def main():
    p = Project()
    p.generate_ninja()
    
    
    
if __name__ == "__main__":
    main()
