package main

import (
    "fmt"
    "os"

    "../gamma"
    "github.com/mkideal/cli"
)

var help = cli.HelpCommand("display help information")


func main() {
    defer gamma.RemoveTmp()
    
    if err := cli.Root(gamma.Root,
        cli.Tree(help),
        cli.Tree(gamma.Init),
        cli.Tree(gamma.DataSelect),
        cli.Tree(gamma.DataImport),
    ).Run(os.Args[1:]); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
