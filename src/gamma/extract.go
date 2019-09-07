package gamma;


import (
    "os";
    "io";
    fp "path/filepath";
    zip "archive/zip";
    set "github.com/deckarep/golang-set";
);


type(
    Extracted struct {
        fileSet set.Set;
        root string;
    };
);


func extractFile(src *zip.File, dst string) error {
    handle := Handler("extractFile");
    
    srcName := src.Name;
    
    in, err := src.Open();
    if err != nil {
        return handle(err, "Could not open file '%s'!", srcName);
    }
    defer in.Close()
    
    out, err := os.Create(dst);
    if err != nil {
        return handle(err, "Could not create file '%s'!", dst);
    }
    defer out.Close();
    
    _, err = io.Copy(out, in);
    if err != nil {
        return handle(err, "Could not copy contents of '%s' into '%s'!", 
                            srcName, dst);
    }
    
    return nil;
}


func extract(path, root string, templates []string) ([]string, error) {
    handle := Handler("extract");
    
    file, err := zip.OpenReader(path);
    
    if err != nil {
        return nil, handle(err, "Could not open zipfile: '%s'!", path);
    }
    
    defer file.Close()
    
    ret := make([]string, BufSize);
    
    // go through files in the zipfile
    for _, zipfile := range file.File {
        srcName := zipfile.Name;
        dst := fp.Join(root, srcName);
        
        if _, err := os.Stat(dst); err == nil {
            // TODO: PLus check if matches templates.
            if os.IsNotExist(err) {
                err := extractFile(zipfile, dst);
                
                if err != nil {
                    return nil, handle(err, 
                    "Failed to extract file : '%s' from zip '%v'!",
                    srcName, file);
                }
                
                ret = append(ret, dst);
                
            } else {
                return nil, handle(err, "Stat failed on file : '%s'!", file);
            }
        }
    }
    return ret, nil;
}
