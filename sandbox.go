package main

import (
	"fmt"
	"os"
)

func Main() (err error) {
	f, err := os.Open("/tmp")
	if err != nil {
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	fmt.Printf("%#v\n", stat.IsDir())

	return nil
}

func main() {
	if err := Main(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
