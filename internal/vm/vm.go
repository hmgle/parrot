package vm

import "parrot/internal/object"

type VM struct {
	stack []object.Object
	ip    int
}

func New() *VM {
	return &VM{
		stack: []object.Object{},
		ip:    0,
	}
}
