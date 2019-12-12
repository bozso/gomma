import gamma as gm
import os.path as path

def main():
    test = path.join("..", "testfiles")
    mli = path.join(test, "mli.json")
    
    dat = gm.DataFile.make(path.join(test, "mli.json"), dat=path.join(test, "vv.mli"))
    
if __name__ == "__main__":
    main()
