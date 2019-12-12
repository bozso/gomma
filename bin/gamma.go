package main

import (
    "fmt"
    "os"

    gamma "../src"
    "github.com/mkideal/cli"
)

var help = cli.HelpCommand("display help information")

var mainErr = gamma.NewModuleErr("main")


func testError1() error {
    var ferr = mainErr("testError1")
    
    if err := testError2(); err != nil {
        return ferr.Wrap(err, "failed to load some file")
    }
    
    return nil
}


func testError2() error {
    var ferr = mainErr("testError2")
    
    file, err := os.Open("asd")
    if err != nil {
        return ferr.Wrap(err, "failed to open file")
    }
    defer file.Close()
    
    return nil
}

func main() {
    if err := testError1(); err != nil {
        fmt.Fprintf(os.Stderr, "Error occurred in gamma main program:%s\n", err)
        os.Exit(1)
    }
    
    return
    
    if err := cli.Root(gamma.Root,
        cli.Tree(help),
        cli.Tree(gamma.MoveFile),
        cli.Tree(gamma.DataSelect),
        cli.Tree(gamma.DataImport),
        cli.Tree(gamma.Like),
        cli.Tree(gamma.Coreg),
        cli.Tree(gamma.Create),
        cli.Tree(gamma.SplitIFG),
    ).Run(os.Args[1:]); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
