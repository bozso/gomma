package cli

import (
	//"github.com/bozso/gotoolbox/cli"
	"github.com/bozso/gotoolbox/path"

	"github.com/bozso/gomma/data"
)

type create struct {
	Data       path.ValidFile
	Param      path.File
	Ftype, Ext string
	MetaFile
	data.Type
}

/*
func (cr *create) SetCli(c *cli.Cli) {
    cr.MetaFile.SetCli(c)
    cr.Type.SetCli(c)

    c.Var(&cr.Data, "dat", "Datafile path.")
    c.Var(&cr.Param, "par", "Parameterfile path.")
    c.StringVar(&cr.Ftype, "ftype", "", "Filetype.")
    c.StringVar(&cr.Ext, "ext", "par", "Extension of parameterfile.")
}

func (c create) Run() (err error) {
    dat, err := data.New(c.Data.ToPath()).WithParFile(c.Param).Load()
    if err != nil {
        return
    }

    if err = common.SaveJson(c.Meta, &dat); err != nil {
        return
    }

    return
}

*/

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
