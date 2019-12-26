package main

import (
    "fmt"
    "os"

    gamma "../src"
    "github.com/mkideal/cli"
)

var help = cli.HelpCommand("display help information")

func main() {
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
