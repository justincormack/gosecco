package ast

// ArithmeticType specifies the different possible arithmetic operations
type ArithmeticType int

// Constants for the different possible types
const (
	PLUS ArithmeticType = iota
	MINUS
	MULT
	DIV
	BINAND
	BINOR
	BINXOR
	LSH
	RSH
	MOD
)

// ArithmeticNames maps the types to names for presentation
var ArithmeticNames = map[ArithmeticType]string{
	PLUS:   "+",
	MINUS:  "-",
	MULT:   "*",
	DIV:    "/",
	BINAND: "&",
	BINOR:  "|",
	BINXOR: "^",
	LSH:    "<<",
	RSH:    ">>",
	MOD:    "%",
}

// Arithmetic represents an arithmetic operation
type Arithmetic struct {
	Op          ArithmeticType
	Left, Right Numeric
}

// Accept implements Expression
func (v Arithmetic) Accept(vs Visitor) {
	vs.AcceptArithmetic(v)
}

// BinaryNegation represents binary negation of a number
type BinaryNegation struct {
	Operand Numeric
}

// Accept implements Expression
func (v BinaryNegation) Accept(vs Visitor) {
	vs.AcceptBinaryNegation(v)
}
