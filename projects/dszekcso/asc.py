from gamma import Project
from os.path import join
from glob import iglob

datapath = join("/mnt", "Dszekcso", "ASC")
out = join("/mnt", "bozso_i", "dszekcso", "asc")

proj = Project(
    looks="1,1",
    out=out,
    pol="vv",
    master="20161205",
)

def preproc():
    proj.select(datapath, start="20161201", stop="20170131",
        lowerLeft="46.050571,18.649662", upperRight="46.139456,18.878696",
        outfile="preselect.list")

def main():
    preproc()
    
if __name__ == "__main__":
    main()
