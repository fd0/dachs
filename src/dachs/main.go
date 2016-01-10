package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

var opts = &struct {
	Verbose bool `short:"v" long:"verbose" description:"be verbose"`
}{}

// V prints the message when verbose is active.
func V(format string, args ...interface{}) {
	if opts.Verbose {
		return
	}

	fmt.Printf(format, args...)
}

func main() {
	var parser = flags.NewParser(opts, flags.Default)

	_, err := parser.Parse()
	if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
		os.Exit(0)
	}

	if err != nil {
		os.Exit(1)
	}

	fmt.Println("main")
}
