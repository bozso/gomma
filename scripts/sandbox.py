from base import gamma as gm, progs
from os.path import join as pjoin

def main():
    test = pjoin(progs, "gamma", "testfiles")
    mli = pjoin(test, "mli.json")
    
    # gm.make(mli, dat=pjoin(test, "vv.mli"), ftype="mli")
    
    gm.stat(mli, "stat")
    
if __name__ == "__main__":
    main()
