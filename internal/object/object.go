package object

import (
	"fmt"
	"strconv"
	"strings"
)

type Type string

const (
	NULLType     Type = "null"
	ERRORType    Type = "error"
	IntType      Type = "int"
	BoolType     Type = "bool"
	StringType   Type = "string"
	ListType     Type = "list"
	FunctionType Type = "function"
	BuiltinType  Type = "builtin"

	FunctionCompiledType Type = "functioncompiled"
)

func (t Type) String() string {
	return string(t)
}

type Object interface {
	Type() Type
	String() string
}

var (
	NULLObj  = &NULL{}
	TRUEObj  = NewBoolean(true)
	FALSEObj = NewBoolean(false)
)

type NULL struct{}

func (null *NULL) Type() Type     { return NULLType }
func (null *NULL) String() string { return "null" }

type Error string

func (e *Error) Type() Type     { return ERRORType }
func (e *Error) String() string { return fmt.Sprintf("error: %s", string(*e)) }

func NewError(f string, a ...any) Object {
	ret := Error(fmt.Sprintf(f, a...))
	return &ret
}

type Integer int64

func (i *Integer) Type() Type     { return IntType }
func (i *Integer) String() string { return strconv.FormatInt(int64(*i), 10) }

type Boolean bool

func (b *Boolean) Type() Type     { return BoolType }
func (b *Boolean) String() string { return strconv.FormatBool(bool(*b)) }

func NewBoolean(b bool) Object {
	ret := Boolean(b)
	return &ret
}

type String string

func (s *String) Type() Type     { return StringType }
func (s *String) String() string { return string(*s) }
func (s *String) Quoted() string {
	return strconv.Quote(string(*s))
}

func NewString(s string) Object {
	o := String(s)
	return &o
}

type List []Object

func (l *List) Type() Type { return ListType }
func (l *List) String() string {
	var elements []string
	for _, e := range *l {
		if s, ok := e.(*String); ok {
			elements = append(elements, s.Quoted())
		} else {
			elements = append(elements, e.String())
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

func NewList(elems ...Object) Object {
	l := append(List{}, elems...)
	return &l
}

type Function struct {
	Params []string
	Body   any
	Env    *Env
}

func (function *Function) Type() Type {
	return FunctionType
}

func (function *Function) String() string {
	return fmt.Sprintf("fn(%s) { %v }", strings.Join(function.Params, ", "), function.Body)
}

type FunctionCompiled struct {
	Instructions []byte
	ParamsCnt    int8
	LocalCnt     int
}

func (functioncompiled *FunctionCompiled) Type() Type {
	return FunctionCompiledType
}

func (functioncompiled *FunctionCompiled) String() string {
	return "<functioncompiled>"
}
