package repl

import (
	"fmt"
	"io"
	"parrot/internal/object"
	"parrot/internal/parser"

	"github.com/chzyer/readline"
)

func EvalREPL() {
	env := object.NewEnv()

	rl, err := readline.New(">>> ")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rl.Close()

	for {
		input, err := rl.Readline()
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			return
		}
		prog, errs := parser.Parse(input)
		if len(errs) > 0 {
			for _, e := range errs {
				fmt.Println(e)
			}
			continue
		}
		if val := prog.Eval(env); val != nil && val != object.NULLObj {
			fmt.Println(val)
		} else {
			fmt.Printf("val: %+v\n", val)
		}
	}
}
