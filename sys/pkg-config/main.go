package main

import (
	"flag"
	"fmt"
)

var (
	flagPrefix       = flag.String("prefix", "", "Prefix for the package")
	flagVersion      = flag.String("version", "", "Version for the package")
	flagDescription  = flag.String("description", "", "Description for the package")
	flagCompileFlags = flag.String("cflags", "", "Compiler flag")
	flagLinkerFlags  = flag.String("libs", "", "Linker flags")
)

func main() {
	flag.Parse()
	fmt.Println("generate:", flag.Args())
}
