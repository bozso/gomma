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
    zips, info := self.Zipfiles, extractInfo{root:fp.Join(root, "extracted")}
    
    for _, zip := range zips {
        image, err := zip.Quicklook(info)
        if err != nil {
            return handle(err, "Failed to retreive quicklook file in zip '%s'!",
                zip.Path)
        }
        fmt.Println(image)
    } 
    
    return nil
}