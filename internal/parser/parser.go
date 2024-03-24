package parser

import (
	"errors"
	"fmt"
	"parrot/internal/lexer"
	"parrot/internal/token"
	"strconv"
)

var ErrEof = errors.New("got eof")

type Error struct {
	Pos int
	Msg string
	Err error
}

type Parser struct {
	l         *lexer.Lexer
	curToken  *token.Token
	peekToken *token.Token
	errs      []*Error
}

func (p *Parser) nextToken() {
	if p.peekToken != nil {
		p.curToken = p.peekToken
	} else {
		nt := p.l.NextToken()
		p.curToken = &nt
	}
	nt := p.l.NextToken()
	p.peekToken = &nt
}

// https://engineering.desmos.com/articles/pratt-parser/
func (p *Parser) parseExpr(rbp int) (e Expr) {
	prefixFn := prefixParsers[p.curToken.Type]
	if prefixFn == nil {
		var err error
		if p.curToken.Type == token.EOF {
			err = ErrEof
		}
		p.errs = append(p.errs, &Error{
			Pos: p.curToken.Pos,
			Msg: fmt.Sprintf("got '%s', want primary expr", p.curToken.Literal),
			Err: err,
		})
		panic(err)
	}
	left := prefixFn(p)
	for p.peekToken.Type != token.SEMICOLON && rbp < bindingPower[p.peekToken.Type] {
		infixFn := infixParsers[p.peekToken.Type]
		if infixFn == nil {
			break
		}
		p.nextToken()
		left = infixFn(p, left)
	}
	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}
	return left
}

func (p *Parser) parseExprStmt() *ExprStmt {
	stmt := &ExprStmt{
		E: p.parseExpr(LowestBP),
	}
	return stmt
}

func (p *Parser) parseStmt() (s Stmt) {
	// TODO
	switch p.curToken.Type {
	}
	return p.parseExprStmt()
}

func (p *Parser) Parse() (prog *Program, errs []*Error) {
	prog = &Program{}
	defer func() {
		_ = recover()
		errs = p.errs
	}()
	for p.curToken.Type != token.EOF {
		if s := p.parseStmt(); s != nil {
			prog.Stmts = append(prog.Stmts, s)
		}
		p.nextToken()
	}
	return prog, p.errs
}

func (p *Parser) peekError(t token.Type) {
	var err error
	if p.peekToken.Type == token.EOF {
		err = ErrEof
	}
	p.errs = append(p.errs, &Error{
		Pos: p.peekToken.Pos,
		Msg: fmt.Sprintf("expected next token to be %v, got %v insted", t, p.peekToken.Type),
		Err: err,
	})
	panic(nil)
}

func (p *Parser) expectPeek(t token.Type) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func Parse(input string) (prog *Program, errs []*Error) {
	p := &Parser{
		l: lexer.New(input),
	}
	p.nextToken()
	return p.Parse()
}

type (
	nudFunc func(p *Parser) Expr
	ledFunc func(p *Parser, left Expr) Expr
)

func identNud(p *Parser) (e Expr) {
	tok := p.curToken
	return &Ident{
		Name: tok.Literal,
		Pos:  tok.Pos,
	}
}

func stringNud(p *Parser) (e Expr) {
	tok := p.curToken
	return &String{
		Literal: tok.Literal,
		Pos:     tok.Pos,
	}
}

func boolNud(p *Parser) (e Expr) {
	tok := p.curToken
	v := false
	if tok.Type == token.TRUE {
		v = true
	}
	return &Boolean{
		Value: v,
		Pos:   tok.Pos,
	}
}

func numberNud(p *Parser) (e Expr) {
	tok := p.curToken
	n, err := strconv.ParseInt(tok.Literal, 0, 64)
	if err != nil {
		p.errs = append(p.errs, &Error{
			Pos: tok.Pos,
			Msg: err.Error(),
		})
		return nil
	}
	return &Integer{
		Value:   n,
		Literal: tok.Literal,
		Pos:     tok.Pos,
	}
}

func lparNud(p *Parser) (e Expr) {
	p.nextToken()
	e = p.parseExpr(LowestBP)
	if !p.expectPeek(token.RPAR) {
		return nil
	}
	return
}

// lbrkNud parse List.
func lbrkNud(p *Parser) (e Expr) {
	tok := p.curToken
	list := &ListExpr{
		LbrackPos: tok.Pos,
	}
	list.List = parseNodeList(p, token.COMMA, token.RBRK)
	list.RbrackPos = p.curToken.Pos
	return list
}

func funcNud(p *Parser) (e Expr) {
	expression := &Function{
		Params: []*Ident{},
		Body:   &Program{},
		Name:   "",
	}
	if p.peekToken.Type != token.LPAR {
		p.nextToken()
		expression.Name = p.curToken.Literal
	}
	if !p.expectPeek(token.LPAR) {
		return nil
	}
	expression.Params = parseFuncParams(p)
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expression.Body = parseBlock(p)
	return expression
}

func parseFuncParams(p *Parser) (params []*Ident) {
	if p.peekToken.Type == token.RPAR {
		p.nextToken()
		return
	}
	p.nextToken()
	params = append(params, &Ident{
		Name: p.curToken.Literal,
		Pos:  p.curToken.Pos,
	})
	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		p.nextToken()
		params = append(params, &Ident{
			Name: p.curToken.Literal,
			Pos:  p.curToken.Pos,
		})
	}
	if !p.expectPeek(token.RPAR) {
		return nil
	}
	return
}

func parseBlock(p *Parser) (block *Program) {
	p.nextToken()
	block = &Program{}
	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		if s := p.parseStmt(); s != nil {
			block.Stmts = append(block.Stmts, s)
		}
		p.nextToken()
	}
	return
}

func parseNodeList(p *Parser, sep, end token.Type) (list []Expr) {
	p.nextToken()
	if p.curToken.Type == end {
		return list
	}
	list = append(list, p.parseExpr(LowestBP))
	for p.peekToken.Type == sep {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpr(LowestBP))
	}
	if !p.expectPeek(end) {
		return nil
	}
	return list
}

func prefixNud(p *Parser) (e Expr) {
	tok := p.curToken
	p.nextToken()
	prefix := &PrefixExpr{
		TokenType: tok.Type,
		Right:     p.parseExpr(PrefixBP),
		Literal:   tok.Literal,
		Pos:       tok.Pos,
	}
	return prefix
}

func infixLed(p *Parser, left Expr) (e Expr) {
	tok := p.curToken
	bp := bindingPower[tok.Type]
	p.nextToken()
	e = &InfixExpr{
		TokenType: tok.Type,
		Left:      left,
		Right:     p.parseExpr(bp),
		Literal:   tok.Literal,
		Pos:       tok.Pos,
	}
	return
}

func assignLed(p *Parser, left Expr) (e Expr) {
	tok := p.curToken
	bp := bindingPower[tok.Type]
	p.nextToken()
	e = &Assign{
		TokenType: tok.Type,
		Left:      left,
		Right:     p.parseExpr(bp),
		Literal:   tok.Literal,
		Pos:       tok.Pos,
	}
	return
}

func callLed(p *Parser, left Expr) (e Expr) {
	return &Call{
		fn:   left,
		args: parseNodeList(p, token.COMMA, token.RPAR),
	}
}

// refer to starlark-go: parseSliceSuffix()
func indexOrsliceLed(p *Parser, left Expr) (e Expr) {
	tok := p.curToken
	lpos := tok.Pos
	var rpos int
	var lo, hi, step Expr
	if p.peekToken.Type != token.COLON {
		p.nextToken()
		i := p.parseExpr(LowestBP)
		// left[i]
		if p.peekToken.Type == token.RBRK {
			p.nextToken()
			rpos = p.curToken.Pos
			return &IndexExpr{
				Left:      left,
				LbrackPos: lpos,
				Index:     i,
				RbrackPos: rpos,
			}
		}
		lo = i
	}
	// slice or substring left[lo:hi:step]
	if p.peekToken.Type == token.COLON {
		p.nextToken()
		if p.peekToken.Type != token.COLON && p.peekToken.Type != token.RBRK {
			p.nextToken()
			hi = p.parseExpr(LowestBP)
		}
	}
	if p.peekToken.Type == token.COLON {
		p.nextToken()
		if p.peekToken.Type != token.RBRK {
			p.nextToken()
			step = p.parseExpr(LowestBP)
		}
	}
	if !p.expectPeek(token.RBRK) {
		return nil
	}
	rpos = p.curToken.Pos
	return &SliceExpr{
		Left:      left,
		LbrackPos: lpos,
		Lo:        lo,
		Hi:        hi,
		Step:      step,
		RbrackPos: rpos,
	}
}

const (
	LowestBP int = iota
	AssignBP
	OrBP
	AndBP
	EqualsBP
	LessGreaterBP
	SumBP
	ProductBP
	ModuloBP
	PrefixBP
	CallBP
	IndexBP
)

var (
	bindingPower  [token.CntToken]int
	prefixParsers [token.CntToken]nudFunc
	infixParsers  [token.CntToken]ledFunc
)

func init() {
	bindingPower[token.ASSIGN] = AssignBP
	bindingPower[token.OR] = OrBP
	bindingPower[token.AND] = AndBP
	bindingPower[token.EQ] = EqualsBP
	bindingPower[token.NOTEQ] = EqualsBP
	bindingPower[token.IN] = LessGreaterBP
	bindingPower[token.LT] = LessGreaterBP
	bindingPower[token.GT] = LessGreaterBP
	bindingPower[token.LE] = LessGreaterBP
	bindingPower[token.GE] = LessGreaterBP
	bindingPower[token.ADD] = SumBP
	bindingPower[token.MINUS] = SumBP
	bindingPower[token.MUL] = ProductBP
	bindingPower[token.DIV] = ProductBP
	bindingPower[token.MOD] = ModuloBP
	bindingPower[token.LPAR] = CallBP
	bindingPower[token.LBRK] = IndexBP

	prefixParsers[token.NUM] = numberNud
	prefixParsers[token.STR] = stringNud
	prefixParsers[token.IDENT] = identNud
	prefixParsers[token.FALSE] = boolNud
	prefixParsers[token.TRUE] = boolNud
	prefixParsers[token.BANG] = prefixNud
	prefixParsers[token.MINUS] = prefixNud
	prefixParsers[token.ADD] = prefixNud // XXX
	prefixParsers[token.LPAR] = lparNud
	prefixParsers[token.LBRK] = lbrkNud
	prefixParsers[token.FUNCTION] = funcNud

	infixParsers[token.ADD] = infixLed
	infixParsers[token.MINUS] = infixLed
	infixParsers[token.MUL] = infixLed
	infixParsers[token.DIV] = infixLed
	infixParsers[token.MOD] = infixLed

	infixParsers[token.LT] = infixLed
	infixParsers[token.LE] = infixLed
	infixParsers[token.GT] = infixLed
	infixParsers[token.GE] = infixLed
	infixParsers[token.EQ] = infixLed
	infixParsers[token.NOTEQ] = infixLed
	infixParsers[token.IN] = infixLed
	infixParsers[token.OR] = infixLed
	infixParsers[token.AND] = infixLed

	infixParsers[token.ASSIGN] = assignLed

	infixParsers[token.LPAR] = callLed
	infixParsers[token.LBRK] = indexOrsliceLed
}
