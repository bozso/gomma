from gamma import Project
from os.path import join
from glob import iglob

data = join("/mnt", "Dszekcso", "ASC")

proj = Project(
    looks="1,1",
    out="",
    pol="vv",
    master="20161205",
)

def preproc():
    datas = ["-d" + path for path in iglob(join(data, "*.zip"))]
    
    print(datas)
    
    proj.select(datas, start="", stop="", lowerLeft="", upperRight="",
        outfile="preselect.list")

def main():
    preproc()
    
if __name__ == "__main__":
    main()
