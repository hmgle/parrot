package parser

import (
	"fmt"
	"parrot/internal/object"
	"parrot/internal/token"
	"strings"
)

// Node represents a node in the AST.
type Node interface {
	String() string
	// TODO position of the node.
}

// Stmt represents a statement in the AST.
type Stmt interface {
	Node
	stmt()
}

// ExprStmt represents an expression statement.
type ExprStmt struct {
	E Expr
}

func (exprstmt *ExprStmt) String() string {
	return exprstmt.E.String()
}

func (exprstmt *ExprStmt) stmt() {}

type Expr interface {
	Node
	Eval(env *object.Env) object.Object
}

type Program struct {
	Stmts []Stmt
}

func (p *Program) Eval(env *object.Env) object.Object {
	var ret object.Object
	for _, stmt := range p.Stmts {
		switch node := stmt.(type) {
		case *ExprStmt:
			ret = node.E.Eval(env)
			if isError(ret) {
				return ret
			}
		}
	}
	return ret
}

type Ident struct {
	Name string
	Pos  int
}

func (ident *Ident) String() string {
	return ident.Name
}

func (ident *Ident) Eval(env *object.Env) object.Object {
	if v, ok := env.Get(ident.Name); ok {
		return v
	}
	if o, ok := object.ResolveBuiltin(ident.Name); ok {
		return o
	}
	return object.NewError("%d: name %q is not defined", ident.Pos+1, ident)
}

type Boolean struct {
	Value bool
	Pos   int
}

func (boolean *Boolean) String() string {
	if boolean.Value {
		return "true"
	}
	return "false"
}

func (boolean *Boolean) Eval(env *object.Env) object.Object {
	if boolean.Value {
		return object.TRUEObj
	}
	return object.FALSEObj
}

type String struct {
	Literal string
	Pos     int
}

func (s *String) String() string {
	return s.Literal
}

func (s *String) Eval(env *object.Env) object.Object {
	return object.NewString(s.Literal)
}

type Integer struct {
	Value   int64
	Literal string
	Pos     int
}

func (n *Integer) String() string {
	return n.Literal
}

func (n *Integer) Eval(env *object.Env) object.Object {
	v := object.Integer(n.Value)
	return &v
}

// ListExpr represents a list literal: [ List ].
type ListExpr struct {
	List      []Expr
	LbrackPos int
	RbrackPos int
}

func (listexpr *ListExpr) String() string {
	var es []string
	for _, m := range listexpr.List {
		es = append(es, m.String())
	}
	return "[" + strings.Join(es, ", ") + "]"
}

func (listexpr *ListExpr) Eval(env *object.Env) object.Object {
	var objs []object.Object
	for _, e := range listexpr.List {
		o := e.Eval(env)
		objs = append(objs, o)
	}
	return object.NewList(objs...)
}

type PrefixExpr struct {
	TokenType token.Type
	Right     Expr
	Literal   string
	Pos       int
}

func (prefixexpr *PrefixExpr) String() string {
	return prefixexpr.Literal + prefixexpr.Right.String()
}

func (prefixexpr *PrefixExpr) Eval(env *object.Env) object.Object {
	right := prefixexpr.Right.Eval(env)
	switch prefixexpr.TokenType {
	case token.BANG:
		switch right {
		case object.TRUEObj:
			return object.FALSEObj
		case object.FALSEObj:
			return object.TRUEObj
		case object.NULLObj:
			return object.TRUEObj
		default:
			return object.FALSEObj
		}
	case token.MINUS:
		if right.Type() != object.IntType {
			return object.NewError("%s: runtime error: unkown operator: -%s", right.String(), right.Type())
		}
		ret := object.Integer(-(int64(*(right.(*object.Integer)))))
		return &ret
	case token.ADD:
		return right
	}
	return object.NewError("runtime error: %s%s", prefixexpr.Literal, right.String())
}

// IndexExpr represents an index expression: Left[Index].
type IndexExpr struct {
	Left      Expr
	LbrackPos int
	Index     Expr
	RbrackPos int
}

func (i *IndexExpr) String() string {
	return fmt.Sprintf("%v[%v]", i.Left, i.Index)
}

func (i *IndexExpr) Eval(env *object.Env) object.Object {
	lft := i.Left.Eval(env)
	if isError(lft) {
		return lft
	}
	idx := i.Index.Eval(env)
	if isError(idx) {
		return idx
	}
	return evalIndexExpr(lft, idx)
}

func evalIndexExpr(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ListType && index.Type() == object.IntType:
		l := left.(*object.List)
		i := index.(*object.Integer)
		if int(*i) >= len(*l) {
			return object.NewError("index out of range")
		}
		return (*l)[int(*i)]
	case left.Type() == object.StringType && index.Type() == object.IntType:
		s := left.(*object.String)
		i := index.(*object.Integer)
		if int(*i) >= len(*s) {
			return object.NewError("index out of range")
		}
		return object.NewString(string(string(*s)[*i]))
		// TODO map
	default:
		return object.NewError("invalid index operator for types %v and %v", left.Type(), index.Type())
	}
}

// SliceExpr represents a slice or substring expression: Left[Lo:Hi]
type SliceExpr struct {
	Left         Expr
	LbrackPos    int
	Lo, Hi, Step Expr
	RbrackPos    int
}

func (s *SliceExpr) String() string {
	return fmt.Sprintf("%v[%v:%v]", s.Left, s.Lo, s.Hi)
}

func (s *SliceExpr) Eval(env *object.Env) object.Object {
	lftObj := s.Left.Eval(env)
	if isError(lftObj) {
		return lftObj
	}
	var (
		loObj   object.Object = object.NULLObj
		hiObj   object.Object = object.NULLObj
		stepObj object.Object = object.NULLObj
	)
	if s.Lo != nil {
		loObj = s.Lo.Eval(env)
		if isError(loObj) {
			return loObj
		}
	}
	if s.Hi != nil {
		hiObj = s.Hi.Eval(env)
		if isError(hiObj) {
			return hiObj
		}
	}
	if s.Step != nil {
		stepObj = s.Step.Eval(env)
		if isError(stepObj) {
			return stepObj
		}
	}
	return evalSliceExpr(lftObj, loObj, hiObj, stepObj)
}

func evalSliceExpr(leftObj, loObj, hiObj, stepObj object.Object) object.Object {
	if leftObj.Type() != object.ListType && leftObj.Type() != object.StringType {
		return object.NewError("runtime error: invalid slice operator for types %v", leftObj.Type())
	}
	if loObj.Type() != object.IntType && loObj.Type() != object.NULLType {
		return object.NewError("TypeError: slice indices must be integers or None, but not %s", loObj.Type())
	}
	if hiObj.Type() != object.IntType && hiObj.Type() != object.NULLType {
		return object.NewError("runtime error: invalid slice HI value for types %v", hiObj.Type())
	}
	if stepObj.Type() != object.IntType && stepObj.Type() != object.NULLType {
		return object.NewError("runtime error: invalid slice STEP value for types %v", stepObj.Type())
	}
	var (
		lo   int
		hi   int
		step int
	)
	if loObj.Type() == object.IntType {
		lo = int(*loObj.(*object.Integer))
	}
	if hiObj.Type() == object.IntType {
		hi = int(*hiObj.(*object.Integer))
	}
	if stepObj.Type() == object.IntType {
		step = int(*stepObj.(*object.Integer))
	} else {
		step = 1
	}
	if step == 0 {
		return object.NewError("ValueError: slice step cannot be zero")
	}
	switch {
	case leftObj.Type() == object.ListType:
		var objs []object.Object
		left := *(leftObj.(*object.List))
		l := len(left)
		if hiObj.Type() == object.NULLType {
			hi = l
		}
		lo = fixOffset(l, lo)
		hi = fixOffset(l, hi)
		if step == 1 {
			objs = left[lo:hi]
		} else if step > 0 && hi >= lo {
			for i := lo; i < hi; i += step {
				objs = append(objs, left[i])
			}
		} else if step < 0 && hi <= lo {
			for i := 0; i > hi; i += step {
				objs = append(objs, left[i])
			}
		}
		return object.NewList(objs...)
	case leftObj.Type() == object.StringType:
		left := string(*(leftObj.(*object.String)))
		l := len(left)
		if hiObj.Type() == object.NULLType {
			hi = l
		}
		lo = fixOffset(l, lo)
		hi = fixOffset(l, hi)
		if step == 1 {
			return object.NewString(left[lo:hi])
		}
		s := []byte{}
		if step > 0 && hi >= lo {
			for i := lo; i < hi; i += step {
				s = append(s, left[i])
			}
		} else if step < 0 && hi <= lo {
			for i := 0; i > hi; i += step {
				s = append(s, left[i])
			}
		}
		return object.NewString(string(s))
	}
	return object.NewError("TypeError: '%s' object has no attribute '__getitem__'", leftObj.String())
}

type Call struct {
	fn   Expr
	args []Expr
}

func (call *Call) String() string {
	panic("not implemented") // TODO: Implement
}

func (call *Call) Eval(env *object.Env) object.Object {
	fnObj := call.fn.Eval(env)
	switch fn := fnObj.(type) {
	case object.BuiltinFn:
		var args []object.Object
		for _, a := range call.args {
			args = append(args, a.Eval(env))
		}
		return fn(args...)
	default:
		panic("call not implemented") // TODO: Implement
	}
}

type Assign struct {
	TokenType token.Type
	Left      Expr
	Right     Expr
	Literal   string
	Pos       int
}

func (assign *Assign) String() string {
	return assign.Left.String() + assign.Literal + assign.Right.String()
}

func (assign *Assign) Eval(env *object.Env) object.Object {
	lv := assign.Left.String()
	rv := assign.Right.Eval(env)
	return env.Set(lv, rv)
}

type InfixExpr struct {
	TokenType token.Type
	Left      Expr
	Right     Expr
	Literal   string
	Pos       int
}

func (infixexpr *InfixExpr) String() string {
	return infixexpr.Left.String() + infixexpr.Literal + infixexpr.Right.String()
}

func (infixexpr *InfixExpr) Eval(env *object.Env) object.Object {
	left := infixexpr.Left.Eval(env)
	if isError(left) {
		return left
	}
	right := infixexpr.Right.Eval(env)
	if isError(right) {
		return right
	}
	switch {
	case left.Type() == object.BoolType && right.Type() == object.BoolType:
		return evalBooleanInfix(infixexpr, left, right)
	case left.Type() == object.IntType && right.Type() == object.IntType:
		return evalNumberInfix(infixexpr, left, right)
	case left.Type() == object.StringType && right.Type() == object.StringType:
		return evalStringInfix(infixexpr, left, right)
	}
	return object.NewError(
		"%d: runtime error: unkown operator: %s %s %s",
		infixexpr.Pos, left.String(), infixexpr.Literal, right.String(),
	)
}

func evalBooleanInfix(infixexpr *InfixExpr, left, right object.Object) object.Object {
	leftVal := bool(*left.(*object.Boolean))
	rightVal := bool(*right.(*object.Boolean))
	switch infixexpr.TokenType {
	case token.AND:
		ret := object.Boolean(leftVal && rightVal)
		return &ret
	case token.OR:
		ret := object.Boolean(leftVal || rightVal)
		return &ret
	case token.EQ:
		ret := object.Boolean(leftVal == rightVal)
		return &ret
	case token.NOTEQ:
		ret := object.Boolean(leftVal != rightVal)
		return &ret
	}
	return object.NewError("pos: %d: runtime error: unkown operator: %s %s %s",
		infixexpr.Pos, left.String(), infixexpr.Literal, right.String())
}

func evalNumberInfix(infixexpr *InfixExpr, left, right object.Object) object.Object {
	leftVal := int64(*left.(*object.Integer))
	rightVal := int64(*right.(*object.Integer))
	switch infixexpr.TokenType {
	case token.ADD:
		ret := object.Integer(leftVal + rightVal)
		return &ret
	case token.MINUS:
		ret := object.Integer(leftVal - rightVal)
		return &ret
	case token.MUL:
		ret := object.Integer(leftVal * rightVal)
		return &ret
	case token.DIV:
		ret := object.Integer(leftVal / rightVal)
		return &ret
	case token.MOD:
		ret := object.Integer(leftVal % rightVal)
		return &ret
	case token.LT:
		ret := object.Boolean(leftVal < rightVal)
		return &ret
	case token.LE:
		ret := object.Boolean(leftVal <= rightVal)
		return &ret
	case token.GT:
		ret := object.Boolean(leftVal > rightVal)
		return &ret
	case token.GE:
		ret := object.Boolean(leftVal >= rightVal)
		return &ret
	case token.EQ:
		ret := object.Boolean(leftVal == rightVal)
		return &ret
	case token.NOTEQ:
		ret := object.Boolean(leftVal != rightVal)
		return &ret
	}
	return object.NewError("pos: %d: runtime error: unkown operator: %s %s %s",
		infixexpr.Pos, left.String(), infixexpr.Literal, right.String())
}

func evalStringInfix(infixexpr *InfixExpr, left, right object.Object) object.Object {
	leftVal := string(*left.(*object.String))
	rightVal := string(*right.(*object.String))
	switch infixexpr.TokenType {
	case token.ADD:
		ret := object.String(leftVal + rightVal)
		return &ret
	case token.LT:
		ret := object.Boolean(leftVal < rightVal)
		return &ret
	case token.LE:
		ret := object.Boolean(leftVal <= rightVal)
		return &ret
	case token.GT:
		ret := object.Boolean(leftVal > rightVal)
		return &ret
	case token.GE:
		ret := object.Boolean(leftVal >= rightVal)
		return &ret
	case token.EQ:
		ret := object.Boolean(leftVal == rightVal)
		return &ret
	case token.NOTEQ:
		ret := object.Boolean(leftVal != rightVal)
		return &ret
	}
	return object.NewError("pos: %d: runtime error: unkown operator: %s %s %s",
		infixexpr.Pos, left.String(), infixexpr.Literal, right.String())
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERRORType
	}
	return false
}

func fixOffset(l, offset int) (fix int) {
	if offset < 0 {
		fix = l + offset
	} else {
		fix = offset
	}
	if fix < 0 {
		fix = 0
	} else if fix > l {
		fix = l
	}
	return fix
}