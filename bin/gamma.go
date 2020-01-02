package main

import (
    "fmt"
    "os"

    gamma "../src"
    "github.com/mkideal/cli"
)

var help = cli.HelpCommand("display help information")

func main() {
    cli = NewCli("gamma")
    cli.SetupGammaCli()
    
    if err := cli.Parse(args[1:]); err != nil {
        fmt.Fprintf(os.Stderr, err)
        os.Exit(1)
    }
}
