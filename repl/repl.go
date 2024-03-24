package repl

import (
	"errors"
	"fmt"
	"io"
	"parrot/internal/object"
	"parrot/internal/parser"
	"strings"

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

	var accumulatedInput []string

	for {
		line, err := rl.Readline()
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			return
		}
		accumulatedInput = append(accumulatedInput, line)
		input := strings.Join(accumulatedInput, "\n")

		prog, errs := parser.Parse(input)
		if len(errs) > 0 {
			// Here, check if error is due to incomplete input.
			isIncomplete := false
			if errors.Is(errs[len(errs)-1].Err, parser.ErrEof) {
				isIncomplete = true
			}

			if isIncomplete {
				rl.SetPrompt("... ")
				continue
			} else {
				for _, e := range errs {
					fmt.Println(e)
				}
				accumulatedInput = []string{} // Reset accumulated input
				rl.SetPrompt(">>> ")
				continue
			}
		}
		if val := prog.Eval(env); val != nil && val != object.NULLObj {
			fmt.Println(val)
		} else {
			fmt.Printf("val: %+v\n", val)
		}
		accumulatedInput = []string{} // Reset accumulated input after successful execution.
		rl.SetPrompt(">>> ")
	}
}
