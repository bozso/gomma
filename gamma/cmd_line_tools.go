package gamma

import (
    //"log"
    "fmt"
    fp "path/filepath"
    //str "strings"
)

var (
    imv = MakeCmd("eog")
)

func (self *S1ProcData) Quicklook(root string) error {
    handle := Handler("S1ProcData.Quicklook")
    zips, info := self.Zipfiles, &ExtractOpt{root:fp.Join(root, "sentinel1")}
    
    for _, zip := range zips {
        s1, err := NewS1Zip(zip, root)
        
        if err != nil {
            return handle(err, 
                "Failed to parse Sentinel-1 information from zipfile '%s'!",
                s1.Path)
        }
        
        image, err := s1.Quicklook(info)
        
        if err != nil {
            return handle(err, "Failed to retreive quicklook file in zip '%s'!",
                s1.Path)
        }
        
        fmt.Println(image)
    } 
    
    return nil
}