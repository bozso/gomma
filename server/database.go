package server

import (
    "fmt"
    "github.com/bozso/gotoolbox/path"

    "github.com/bozso/gomma/slc"
    "github.com/bozso/gomma/mli"
    s1 "github.com/bozso/gomma/sentinel1"
    ifg "github.com/bozso/gomma/interferogram"
)

type RecordDB map[string]Record 

type DataFiles struct {
    db RecordDB
}

func (d *DataFiles) AddDataFile()

func (d *DataFiles) RemoveRecord(id string) (err error) {
    rec, ok := d.db[id]
    
    if !ok {
        err = fmt.Errorf("")
    }
    
    if err = rec.c.Finalize(); err != nil {
        return
    }
    
    delete(d.db, id)
}
