package ast

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type StringVisitorSuite struct{}

var _ = Suite(&StringVisitorSuite{})

func (s *StringVisitorSuite) Test_Variable(c *C) {
	c.Assert(ExpressionString(Variable{"foo1"}), Equals, "foo1")
}

func (s *StringVisitorSuite) Test_Argument(c *C) {
	sv := &StringVisitor{}

	Argument{3}.Accept(sv)

	c.Assert(sv.String(), Equals, "arg3")
}

func (s *StringVisitorSuite) Test_NumericLiteral(c *C) {
	sv := &StringVisitor{}

	NumericLiteral{42}.Accept(sv)

	c.Assert(sv.String(), Equals, "42")
}

func (s *StringVisitorSuite) Test_BooleanLiteral(c *C) {
	sv := &StringVisitor{}

	BooleanLiteral{true}.Accept(sv)

	c.Assert(sv.String(), Equals, "true")

	sv = &StringVisitor{}

	BooleanLiteral{false}.Accept(sv)

	c.Assert(sv.String(), Equals, "false")
}

func (s *StringVisitorSuite) Test_Comparison(c *C) {
	sv := &StringVisitor{}

	Comparison{Op: GT, Left: NumericLiteral{42}, Right: NumericLiteral{1}}.Accept(sv)

	c.Assert(sv.String(), Equals, "(> 42 1)")

	sv = &StringVisitor{}

	Comparison{Op: EQL, Left: Argument{1}, Right: NumericLiteral{1}}.Accept(sv)

	c.Assert(sv.String(), Equals, "(== arg1 1)")
}

func (s *StringVisitorSuite) Test_Arithmetic(c *C) {
	sv := &StringVisitor{}

	Arithmetic{Op: LSH, Left: NumericLiteral{42}, Right: NumericLiteral{3}}.Accept(sv)

	c.Assert(sv.String(), Equals, "(<< 42 3)")

	sv = &StringVisitor{}

	Arithmetic{Op: PLUS, Left: Argument{42}, Right: NumericLiteral{1}}.Accept(sv)

	c.Assert(sv.String(), Equals, "(+ arg42 1)")
}

func (s *StringVisitorSuite) Test_BinaryNegation(c *C) {
	sv := &StringVisitor{}

	BinaryNegation{Operand: NumericLiteral{42}}.Accept(sv)

	c.Assert(sv.String(), Equals, "^42")
}

func (s *StringVisitorSuite) Test_Call(c *C) {
	sv := &StringVisitor{}

	Call{Name: "foo1", Args: []Any{BinaryNegation{NumericLiteral{42}}, BooleanLiteral{false}, Argument{3}}}.Accept(sv)

	c.Assert(sv.String(), Equals, "foo1(^42, false, arg3)")
}

func (s *StringVisitorSuite) Test_Inclusion(c *C) {
	sv := &StringVisitor{}

	Inclusion{Positive: false, Left: BinaryNegation{Argument{0}}, Rights: []Numeric{NumericLiteral{23}, Argument{3}}}.Accept(sv)

	c.Assert(sv.String(), Equals, "notIn(^arg0, 23, arg3)")
}

func (s *StringVisitorSuite) Test_And(c *C) {
	sv := &StringVisitor{}

	And{Left: Comparison{Op: GT, Left: NumericLiteral{42}, Right: NumericLiteral{1}}, Right: Comparison{Op: EQL, Left: NumericLiteral{42}, Right: NumericLiteral{42}}}.Accept(sv)

	c.Assert(sv.String(), Equals, "(&& (> 42 1) (== 42 42))")
}

func (s *StringVisitorSuite) Test_Or(c *C) {
	sv := &StringVisitor{}

	Or{Left: Comparison{Op: GT, Left: NumericLiteral{42}, Right: Argument{1}}, Right: Comparison{Op: EQL, Left: NumericLiteral{42}, Right: NumericLiteral{42}}}.Accept(sv)

	c.Assert(sv.String(), Equals, "(|| (> 42 arg1) (== 42 42))")
}

func (s *StringVisitorSuite) Test_Negation(c *C) {
	sv := &StringVisitor{}

	Negation{Or{Left: Comparison{Op: GT, Left: NumericLiteral{42}, Right: Argument{1}}, Right: Comparison{Op: EQL, Left: NumericLiteral{42}, Right: NumericLiteral{42}}}}.Accept(sv)

	c.Assert(sv.String(), Equals, "!(|| (> 42 arg1) (== 42 42))")
}
