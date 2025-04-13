package main

import (
	"fmt"
	"os"
)

func main() {

	var args []string
	cmd := "-h"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}

	switch cmd {
	case "oss-fuzz-corpus":
		ossFuzzCorpusSubCommand(args)
	case "unicode-case-folding-map":
		unicodeCaseFoldingMapSubCommand(args)
	case "emb-structs":
		embStructsSubCommand(args)
	case "-h":
		fallthrough
	default:
		fmt.Fprintf(os.Stderr, `Usage: _tools <subcommand> [options]
subcommands:
  oss-fuzz-corpus
  unicode-case-folding-map
  emb-structs
`)
		os.Exit(1)
	}
}

func usage(u func(), err error) {
	u()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}
	os.Exit(1)
}
