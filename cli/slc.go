package cli

/*

import (
    "fmt"
    "path/filepath"
    "strings"

    "github.com/bozso/gamma/utils"
    "github.com/bozso/gamma/data"
    "github.com/bozso/gamma/dem"
    "github.com/bozso/gamma/geo"
)


var SplitIFG = &cli.Command{
    Name: "splitIfg",
    Desc: "Split Beam/Spectrum Interferometry",
    Argv: func() interface{} { return &splitIfg{} },
    Fn: splitIfgFn,
}

type splitIfg struct {
    SBIOpt
    SSIOpt
    SpectrumMode string `cli:"M,mode"`
    Master       string `cli:"*m,master"`
    Slave        string `cli:"*s,slave"`
    Mli          string `cli:"mli"`
}


func splitIfgFn(ctx *cli.Context) (err error) {
    var ferr = merr.Make("splitIfgFn")

    si := ctx.Argv().(*splitIfg)

    ms, ss := si.Master, si.Slave

    var m, s SLC

    if err = Load(ms, &m); err != nil {
        return ferr.Wrap(err)
    }

    if err = Load(ss, &s); err != nil {
        return ferr.Wrap(err)
    }

    mode := strings.ToUpper(si.SpectrumMode)

    switch mode {
    case "BEAM", "B":
        if err = SameShape(m, s); err != nil {
            return ferr.Wrap(err)
        }

        if err = m.SplitBeamIfg(s, si.SBIOpt); err != nil {
            return ferr.Wrap(err)
        }

    //case "SPECTRUM", "S":
        //opt := si.SSIOpt

        //Mli, err := LoadDataFile(si.Mli)
        //if err != nil {
            //return err
        //}

        //mli, ok := Mli.(MLI)

        //if !ok {
            //return TypeErr.Make(Mli, "mli", "MLI")
        //}

        //out, err := m.SplitSpectrumIfg(s, mli, opt)

        //if err != nil {
            //return err
        //}

        // still need to figure out the returned files
        //return nil
    default:
        err = UnrecognizedMode{name:"Split Interferogram", got: mode}
        return ferr.Wrap(err)
    }
    return nil
}
*/
