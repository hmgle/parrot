package main

import (
	"flag"
	"parrot/repl"
)

func main() {
	var useVM bool

	flag.BoolVar(&useVM, "vm", false, "Use the Tau VM instead of eval method. (faster)")
	flag.Parse()
	if useVM {
		repl.VMREPL()
	} else {
		repl.EvalREPL()
	}
}
