package gamma

import (
    "log"
    "fmt"
    fp "path/filepath"
    str "strings"
)

var (
    imv = MakeCmd("eog")
)

func (self *S1ProcData) Quicklook(root string) error {
    var err error
    zips, info := self.Zipfiles, extractInfo{root:fp.Join(root, "extracted")}
    images := make([]string, len(zips))
    
    for ii, zip := range zips {
        if images[ii], err = zip.Quicklook(info); err != nil {
            return fmt.Errorf("In S1ProcData.Quicklook: Failed to retreive " +
                "quicklook file in zip '%s'!", zip.Path)
        }
    } 
    
    
    imv(str.Join(images, " "))
    
    return nil
}