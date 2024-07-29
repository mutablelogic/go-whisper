package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var (
	flagDir          = flag.String("dir", "${PKG_CONFIG_PATH}", "Destination directory")
	flagPrefix       = flag.String("prefix", "", "Prefix for the package")
	flagVersion      = flag.String("version", "", "Version for the package")
	flagDescription  = flag.String("description", "", "Description for the package")
	flagCompileFlags = flag.String("cflags", "", "Compiler flag")
	flagLinkerFlags  = flag.String("libs", "", "Linker flags")
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "Missing filename")
		os.Exit(-1)
	}
	dest := filepath.Join(os.ExpandEnv(*flagDir), flag.Arg(0))

	var prefix string
	if *flagPrefix != "" {
		var err error
		prefix, err = filepath.Abs(*flagPrefix)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
	}

	w, err := os.Create(dest)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
	defer w.Close()

	// Write the package
	if prefix != "" {
		fmt.Fprintf(w, "prefix=%s\n\n", prefix)
	}
	fmt.Fprintf(w, "Name: %s\n", filepath.Base(dest))
	if *flagDescription != "" {
		fmt.Fprintf(w, "Description: %s\n", *flagDescription)
	} else {
		fmt.Fprintf(w, "Description: No description\n")
	}
	if *flagVersion != "" {
		fmt.Fprintf(w, "Version: %s\n", *flagVersion)
	}
	if *flagCompileFlags != "" {
		fmt.Fprintf(w, "Cflags: %s\n", *flagCompileFlags)
	}
	if *flagLinkerFlags != "" {
		fmt.Fprintf(w, "Libs: %s\n", *flagLinkerFlags)
	}
}
