package main

import (
    "fmt"
    //"log"
    "os"
    "../gamma"
    "github.com/mkideal/cli"
)

var help = cli.HelpCommand("display help information")


func main() {
    defer gamma.RemoveTmp()
    
    if err := cli.Root(gamma.Root,
        cli.Tree(help),
        cli.Tree(gamma.Process),
        cli.Tree(gamma.Init),
    ).Run(os.Args[1:]); err != nil {
        fmt.Fprintf(os.Stderr, "Error occurred: %v\n", err)
        os.Exit(1)
    }
}
