package data

type Importer struct {
    RngKey, AziKey, TypeKey, DateKey string
}

func (i Importer) Import(dat, par string) (f File, err error) {
    f.DatFile, f.ParFile = dat, par
    
    pr, err := newGammaParams(par)
    if err != nil { return }
    
    f.Ra.Rng, err = pr.Int(i.RngKey, 0)
    if err != nil { return }
    
    f.Ra.Azi, err = pr.Int(i.AziKey, 0)
    if err != nil { return }
    
    s, err := pr.Param(i.TypeKey)
    if err != nil { return }
    
    err = f.Dtype.Set(s)
    if err != nil { return }
    
    if k := i.DateKey; len(k) != 0 {
        s, err = pr.Param(k)
        if err != nil { return }
        
        f.Time, err = DateFmt.Parse(s)
    }
    
    return
}

var (
    Simple = Importer{
        RngKey: "range_samples",
        AziKey: "azimuth_lines",
        TypeKey: "image_format",
        DateKey: "date",
    }
)
