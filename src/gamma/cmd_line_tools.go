package gamma

import (
    "log"
)

func (self *S1ProcData) Quicklook() error {
    log.Printf("%v", self.MasterDate)
    
    return nil
}