package token

type Type int

type Token struct {
	Type    Type
	Literal string
	Pos     int
}

const (
	ERR    Type = iota
	EOF         // "eof"
	LPAR        // "("
	RPAR        //  ")"
	LBRK        // "["
	RBRK        // "]"
	LBRACE      // "{"
	RBRACE      // "}"
	IN          // "in"

	OR   // or
	AND  // and
	BANG // "NOT"
	LT   // "LT"
	GT   // "GT"
	LE   // "LE"
	GE   // "GE"
	EQ   // "EQ"
	ASSIGN
	NOTEQ // "NE"
	IDENT
	NUM // "number"
	STR // "string"

	LEN // "len"
	REG // "regexp"

	TRUE  // "true"
	FALSE // "false"

	DOT    // "."
	DOTDOT // ".."

	ADD   // "add"
	MINUS // "minus"
	MUL   // "mul"
	DIV   // "div"
	MOD   // "mod"

	NEGSIGN
	POSSIGN

	COLON     // :
	COMMA     // ,
	SEMICOLON // ;

	RANGE
	INDEX
	FUNCTION

	CntToken
)

var typeNames = [...]string{
	EOF:    "eof",
	LPAR:   "(",
	RPAR:   ")",
	LBRK:   "[",
	RBRK:   "]",
	LBRACE: "{",
	RBRACE: "}",
	IN:     "in",

	OR:     "or",
	AND:    "and",
	BANG:   "NOT",
	LT:     "<",
	GT:     ">",
	LE:     "<=",
	GE:     ">=",
	EQ:     "==",
	ASSIGN: "=",
	NOTEQ:  "!=",
	IDENT:  "identifier",
	NUM:    "number",
	STR:    "string",

	LEN: "len",
	REG: "regexp",

	TRUE:  "true",
	FALSE: "false",

	DOT:    ".",
	DOTDOT: "..",

	ADD:   "add",
	MINUS: "minus",
	MUL:   "mul",
	DIV:   "div",
	MOD:   "mod",

	NEGSIGN: "-",
	POSSIGN: "+",

	COLON:     ":",
	COMMA:     ",",
	SEMICOLON: ";",

	RANGE:    "range",
	INDEX:    "index",
	FUNCTION: "function",
}

func (t Type) String() string {
	return typeNames[t]
}

var keywords = map[string]Type{
	"and": AND,
	"or":  OR,

	"true":  TRUE,
	"false": FALSE,

	"in":     IN,
	"regexp": REG,
	"fn":     FUNCTION,
}

func LookupKeyWord(identifier string) Type {
	if tok, ok := keywords[identifier]; ok {
		return tok
	}
	return IDENT
}
