package main

import (
    "fmt"
    //"log"
    "os"
    gm "../gamma"
)

var commands = []string{"proc", "list", "init", "batch", "ras", "dis", "iono"}

func main() {
    defer gm.RemoveTmp()
    
    if len(os.Args) < 2 {
        fmt.Printf("Expected one of the following subcommands: %v!\n",
            gm.CommandsAvailable)
        return
    }
    
    mode := os.Args[1]    
    args := gm.NewArgs(os.Args[2:])
    
    
    cmd, ok := gm.Commands[mode]
    
    if !ok {
        fmt.Printf("Expected on of the following subcommands: %v!\n",
            gm.CommandsAvailable)
        return
    }
    
    if err := cmd(args); err != nil {
        fmt.Printf("Failed to execute command '%s'! Error: %s\n", mode, err)
        return
    }
}

type (
    S1Pair struct {
        master, slave gm.S1SLC
    }
    
    SLCPair struct {
        master, slave gm.SLC
    }
)

func iono() error {
    var (
        err error
        orig S1Pair
    )
    
    if orig.master, err = gm.FromTabfile(""); err != nil {
        return gm.Handle(err, "failed to import S1SLC struct")
    }
    
    if orig.slave, err = gm.FromTabfile(""); err != nil {
        return gm.Handle(err, "failed to import S1SLC struct")
    }
    
    var deramp S1Pair

    deramp.master, err = orig.master.DerampRef()
    if  err != nil {
        return gm.Handle(err, "failed to deramp master S1SLC")
    }
    
    deramp.slave, err = orig.slave.DerampSlave(&orig.master, gm.RngAzi{}, false)
    if err != nil {
        return gm.Handle(err, "failed to deramp slave S1SLC")
    }
    
    const (
        rslc, ifg, hgt = "RSLC", "IFG", "dem.dem"
    )
    
    if err = os.MkdirAll(rslc, os.ModePerm); err != nil {
        return gm.Handle(err, "failed to create directory '%s'", rslc)
    }
    
    if err = os.MkdirAll(ifg, os.ModePerm); err != nil {
        return gm.Handle(err, "failed to create directory '%s'", ifg)
    }
    
    mID, sID := orig.master.Format(gm.DateShort), orig.slave.Format(gm.DateShort)
    ID := fmt.Sprintf("%s_%s", mID, sID)
    
    
    coreg := gm.S1Coreg {
        Tab: deramp.master.Tab,
        ID: mID,
        OutDir: ".",
        RslcPath: rslc,
        IfgPath: ifg,
        Hgt: hgt,
        Poly1: "-",
        Poly2: "-",
        Looks: gm.RngAzi{Rng:1, Azi:1},
        Clean: false,
        CoregOpt: gm.CoregOpt{
            CoherenceThresh:  0.8,
            FractionThresh:   0.01,
            PhaseStdevThresh: 0.8,
        },
    }
    
    var out gm.CoregOut
    if out, err = coreg.Coreg(&deramp.slave, nil); err != nil {
        return gm.Handle(err, "coregistration failed")
    }
    
    if !out.Ok {
        return gm.Handle(err, "coregistration failed")
    }
    
    lookup := ID + ".lt_fine"
    
    var slc SLCPair
    
    mopts := gm.MosaicOpts{Looks: gm.RngAzi{}}
    
    if slc.master, err = deramp.master.Mosaic(mopts); err != nil {
        return gm.Handle(err, "failed to mosaic master S1SLC")
    }
    
    if slc.slave, err = deramp.slave.Mosaic(mopts); err != nil {
        return gm.Handle(err, "failed to mosaic slave S1SLC")
    }
    
    var mmli gm.MLI
    if mmli, err = slc.master.MakeMLI(gm.MLIOpt{}); err != nil {
        return gm.Handle(err, "failed to create master MLI")
    }
    
    ssiOpt := gm.SSIOpt{
        Hgt: hgt,
        LtFine: lookup,
        OutDir: ".",
        Mode: gm.IfgUnwrapped,
    }
    
    //ssiOut, err := slc.master.SplitSpectrumIfg(slc.slave, mmli, ssiOpt)
    _, err = slc.master.SplitSpectrumIfg(slc.slave, mmli, ssiOpt)
    
    if err != nil {
        return gm.Handle(err, "SSI_INT failed")
    }
    
    //if err = ssiOut.ifg.Move("."); err != nil {
        //return gm.Handle(err, "failed to move SSI IFG")
    //}
    
    //if err = ssiOut.unw.Move("."); err != nil {
        //return gm.Handle(err, "failed to move SSI unwrapped IFG")
    //}
    
    return nil
}
