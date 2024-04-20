package main

import (
	"flag"
	"parrot/repl"
)

func main() {
	var useVM bool

	flag.BoolVar(&useVM, "vm", false, "Use the VM instead of eval method.")
	flag.Parse()
	if useVM {
		repl.VMREPL()
	} else {
		repl.EvalREPL()
	}
}
