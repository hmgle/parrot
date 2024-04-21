package repl

import (
	"errors"
	"fmt"
	"io"
	"parrot/internal/compile"
	"parrot/internal/object"
	"parrot/internal/parser"
	"parrot/internal/vm"
	"strings"

	"github.com/chzyer/readline"
)

func VMREPL() {
	rl, err := readline.New(">>> ")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rl.Close()

	var accumulatedInput []string

	machine := vm.New()
	c := compile.New()
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
		err = c.Compile(prog)
		if err != nil {
			fmt.Printf("err: %+v\n", err)
			continue
		}
		machine.Next(c.Constants, c.OpCodes.Output())
		c.OpCodes = []compile.Instruction{}
		c.Constants = []object.Object{}
		if err = machine.Run(); err != nil {
			fmt.Printf("runtime error: %v\n", err)
			continue
		}
		if val := machine.LastPoppedStackElem(); val != nil && val != object.NULLObj {
			fmt.Println(val)
		}
		accumulatedInput = []string{} // Reset accumulated input after successful execution.
		rl.SetPrompt(">>> ")
	}
}

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
		}
		accumulatedInput = []string{} // Reset accumulated input after successful execution.
		rl.SetPrompt(">>> ")
	}
}
