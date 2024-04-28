package parser

import (
	"fmt"
	"parrot/internal/code"
	"parrot/internal/compile"
	"parrot/internal/object"
	"parrot/internal/token"
	"strings"
)

// Node represents a node in the AST.
type Node interface {
	String() string
	// TODO position of the node.
	compile.Compilable
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
func (expresmt *ExprStmt) Compile(c *compile.Compiler) error {
	return expresmt.E.Compile(c)
}

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

func (p *Program) Compile(c *compile.Compiler) error {
	for _, stmt := range p.Stmts {
		err := stmt.Compile(c)
		if err != nil {
			return err
		}
	}
	return nil
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

func (ident *Ident) Compile(c *compile.Compiler) error {
	if symbol, ok := c.Resolve(ident.Name); ok {
		c.LoadSymbol(symbol)
		return nil
	}
	return fmt.Errorf("undefined variable %s", ident.Name)
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

func (boolean *Boolean) Compile(c *compile.Compiler) error {
	if boolean.Value {
		c.Op(code.OpTrue)
	} else {
		c.Op(code.OpFalse)
	}
	return nil
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

func (s *String) Compile(c *compile.Compiler) error {
	o := object.NewString(s.Literal)
	c.OpArg(code.OpConstant, c.Const(o))
	return nil
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

func (n *Integer) Compile(c *compile.Compiler) error {
	o := object.Integer(n.Value)
	c.OpArg(code.OpConstant, c.Const(&o))
	return nil
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

func (listexpr *ListExpr) Compile(c *compile.Compiler) error {
	for _, n := range listexpr.List {
		if err := n.Compile(c); err != nil {
			return err
		}
	}
	c.OpArg(code.OpList, uint32(len(listexpr.List)))
	return nil
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

func (prefixexpr *PrefixExpr) Compile(c *compile.Compiler) (err error) {
	switch prefixexpr.TokenType {
	case token.BANG:
		err = prefixexpr.Right.Compile(c)
		if err != nil {
			return
		}
		c.Op(code.OpBang)
		return nil
	case token.MINUS:
		err = prefixexpr.Right.Compile(c)
		if err != nil {
			return
		}
		c.Op(code.OpMinus)
		return nil
	case token.ADD:
		return prefixexpr.Right.Compile(c)
	}
	return fmt.Errorf("runtime error: %s", prefixexpr.Literal)
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

func (i *IndexExpr) Compile(c *compile.Compiler) error {
	if err := i.Left.Compile(c); err != nil {
		return err
	}
	if err := i.Index.Compile(c); err != nil {
		return err
	}
	c.Op(code.OpIndex)
	return nil
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

func (s *SliceExpr) Compile(c *compile.Compiler) error {
	// TODO
	panic("not implemented")
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
	var args []string
	for _, a := range call.args {
		args = append(args, a.String())
	}
	return fmt.Sprintf("%v(%s)", call.fn, strings.Join(args, ", "))
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
	case *object.Function:
		if len(call.args) != len(fn.Params) {
			return object.NewError(
				"wrong number of arguments: expected %d, got %d",
				len(fn.Params),
				len(call.args),
			)
		}
		newEnv := object.NewEnvWrap(fn.Env)
		for i, a := range call.args {
			newEnv.Set(fn.Params[i], a.Eval(env))
		}
		return fn.Body.(*Program).Eval(newEnv)
	default:
		return object.NewError("%q object is not callable", fnObj.Type())
	}
}

func (call *Call) Compile(c *compile.Compiler) (err error) {
	for _, arg := range call.args {
		err = arg.Compile(c)
		if err != nil {
			return err
		}
	}
	err = call.fn.Compile(c)
	if err != nil {
		return err
	}
	c.OpArg(code.OpCall, uint32(len(call.args)))
	return nil
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

func (assign *Assign) Compile(c *compile.Compiler) (err error) {
	symbol := c.Define(assign.Left.String())
	err = assign.Right.Compile(c)
	if err != nil {
		return
	}
	if symbol.Scope == compile.GlobalScope {
		c.OpArg(code.OpSetGlobal, uint32(symbol.Index))
	} else {
		c.OpArg(code.OpSetLocal, uint32(symbol.Index))
	}
	return nil
}

type Function struct {
	Params []*Ident
	Body   *Program
	Name   string
}

func (function *Function) String() string {
	var params []string
	for _, p := range function.Params {
		params = append(params, p.String())
	}
	return fmt.Sprintf("fn(%s) {%v}", strings.Join(params, ","), function.Body)
}

func (function *Function) Eval(env *object.Env) object.Object {
	var params []string

	for _, p := range function.Params {
		params = append(params, p.String())
	}
	fn := &object.Function{
		Params: params,
		Body:   function.Body,
		Env:    env,
	}
	env.Set(function.Name, fn)
	return fn
}

func (function *Function) Compile(c *compile.Compiler) (err error) {
	nc := c.NewForFunction()

	for _, param := range function.Params {
		nc.Define(param.Name)
	}

	err = function.Body.Compile(nc)
	if err != nil {
		return err
	}
	f := object.FunctionCompiled{
		Instructions: nc.OpCodes.Output(),
		ParamsCnt:    int8(len(function.Params)),
		LocalCnt:     nc.SymbolTable.NumDefinitions,
	}
	c.OpArg(code.OpConstant, c.Const(&f))

	if function.Name != "" {
		symbol := c.Define(function.Name)
		if symbol.Scope == compile.GlobalScope {
			c.OpArg(code.OpSetGlobal, uint32(symbol.Index))
		} else {
			c.OpArg(code.OpSetLocal, uint32(symbol.Index))
		}
	}
	return nil
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

func (infixexpr *InfixExpr) Compile(c *compile.Compiler) (err error) {
	err = infixexpr.Left.Compile(c)
	if err != nil {
		return
	}
	err = infixexpr.Right.Compile(c)
	if err != nil {
		return
	}
	var op code.OpCode
	switch infixexpr.TokenType {
	case token.ADD:
		op = code.OpAdd
	case token.MINUS:
		op = code.OpSub
	case token.MUL:
		op = code.OpMul
	case token.DIV:
		op = code.OpDiv
	case token.MOD:
		op = code.OpMod
	case token.GT:
		op = code.OpCmpGT
	case token.GE:
		op = code.OpCmpGE
	case token.EQ:
		op = code.OpCmpEQ
	case token.NOTEQ:
		op = code.OpCmpNE
	case token.LT:
		op = code.OpCmpLT
	case token.LE:
		op = code.OpCmpLE
	case token.AND:
		op = code.OpAnd
	case token.OR:
		op = code.OpOr
	default:
		panic(fmt.Sprintf("unkown BinOp: %s", infixexpr.TokenType))
	}
	c.Op(op)
	return nil
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
