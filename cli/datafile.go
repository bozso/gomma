package cli

import (
    "github.com/bozso/gotoolbox/cli"
    "github.com/bozso/gotoolbox/path"

    "github.com/bozso/gomma/data"
)

type like struct {
    indat data.File
    in, out, ext string 
    Dtype data.Type
}

func (l *like) SetCli(c* cli.Cli) {
    c.Var(&l.indat, "in", "Reference metadata file")
    c.StringVar(&l.out, "out", "", "Output metadata file")
    c.Var(&l.Dtype, "dtype", "Output file datatype")
    c.StringVar(&l.ext, "ext", "dat", "Extension of datafile")
}

func (l like) Run() (err error) {
    out, indat := l.out, l.indat
    
    //var indat DatFile
    //if err = Load(in, &indat); err != nil {
        //return ferr.Wrap(err)
    //}
    
    dtype := l.Dtype

    if dtype == Unknown {
        dtype = indat.Dtype()
    }
    
    if out, err = filepath.Abs(out); err != nil {
        return
    }
    
    outdat := data.File{
        Dat: fmt.Sprintf("%s.%s", out, l.ext),
        Ra: indat.Ra,
        DType: dtype,
    }
    
    if err = outdat.Save(out); err != nil {
        return
    }
    
    return nil
}

type move struct {
    outDir   string
    MetaFile
}

func (m *move) SetCli(c *cli.Cli) {
    m.MetaFile.SetCli(c)
    c.StringVar(&m.outDir, "out", ".", "Output directory")
}

func (m move) Run() (err error) {
    var ferr = merr.Make("move.Run")

    path := m.Meta
    
    var dat DatParFile
    if err = LoadJson(path, &dat); err != nil {
        return ferr.WrapFmt(err,
            "failed to parse json metadatafile '%s'", path) 
    }
    
    out := m.outDir
    
    if dat, err = dat.Move(out); err != nil {
        return ferr.Wrap(err)
    }
    
    if path, err = Move(path, out); err != nil {
        return ferr.Wrap(err)
    }
    
    if err = SaveJson(path, dat); err != nil {
        return ferr.WrapFmt(err, "failed to refresh json metafile")
    }
    
    return nil
}

type create struct {
    Data path.ValidFile
    Param path.File
    Ftype, Ext string
    MetaFile
    data.Type
}

func (cr *create) SetCli(c *cli.Cli) {
    cr.MetaFile.SetCli(c)
    cr.DType.SetCli(c)
    
    c.Var(&cr.Dat, "dat", "", "Datafile path.")
    c.Var(&cr.Par, "par", "", "Parameterfile path.")
    c.StringVar(&cr.Ftype, "ftype", "", "Filetype.")
    c.StringVar(&cr.Ext, "ext", "par", "Extension of parameterfile.")
}

func (c create) Run() (err error) {
    data, err := data.New(c.Data.ToFile()).WithPar(c.Param).Load()
    if err != nil {
        return
    }
    
    if err = SaveJson(c.Meta, &data); err != nil {
        return
    }
    
    return
}

/*
type (
    Stat struct {
        Out string `cli:"*o,out" usage:"Output file`
        Subset
        MetaFile
    }
)

var imgStat = Gamma.Must("image_stat")

func stat(args Args) (err error) {
    s := Stat{}
    
    if err := args.ParseStruct(&s); err != nil {
        return ParseErr.Wrap(err)
    }
    
    var dat DatFile
    
    if err = Load(s.Meta, &dat); err != nil {
        return
    }
    
    //s.Subset.Parse(dat)
    
    _, err = imgStat(dat.Datfile(), dat.Rng(), s.RngOffset, s.AziOffset,
                     s.RngWidth, s.AziLines, s.Out)
    
    return
}
*/
