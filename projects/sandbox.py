import gamma as gm
import os.path as path

def main():
    test = path.join("..", "testfiles")
    mli = path.join(test, "mli.json")
    
    dat = gm.DataFile(mli)
    
    print(dat.like())
    
if __name__ == "__main__":
    main()
