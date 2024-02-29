package object

type BuiltinFn func(args ...Object) Object

func (b BuiltinFn) Type() Type {
	return BuiltinType
}

func (b BuiltinFn) String() string {
	return "<builtin function>"
}

func ResolveBuiltin(name string) (BuiltinFn, bool) {
	for _, b := range Builtins {
		if name == b.Name {
			return b.Builtin, true
		}
	}
	return nil, false
}

var Builtins = []struct {
	Name    string
	Builtin BuiltinFn
}{
	{
		Name: "len",
		Builtin: func(args ...Object) Object {
			if l := len(args); l != 1 {
				return NewError("len: wrong number of arguments, expected 1, got %d", l)
			}
			switch o := args[0].(type) {
			case *List:
				ret := Integer(len(*o))
				return &ret
			case *String:
				ret := Integer(len(*o))
				return &ret
			default:
				return NewError("len: object of type %q has no length", o.Type())
			}
		},
	},
}
