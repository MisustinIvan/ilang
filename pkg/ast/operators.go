package ast

type BinaryOperator int

const (
	Addition BinaryOperator = iota
	Subtraction
	Multiplication
	Division
	Equality
	Inequality
	LesserThan
	GreaterThan
	LesserOrEqualThan
	GreaterOrEqualThan
	LeftShift
	RightShift
	LogicAnd
	LogicOr
)

func (o BinaryOperator) String() string {
	s := "UNKNOWN"
	switch o {
	case Addition:
		s = "Addition"
	case Subtraction:
		s = "Subtraction"
	case Multiplication:
		s = "Multiplication"
	case Division:
		s = "Division"
	case Equality:
		s = "Equality"
	case Inequality:
		s = "Inequality"
	case LesserThan:
		s = "LesserThan"
	case GreaterThan:
		s = "GreaterThan"
	case LesserOrEqualThan:
		s = "LesserOrEqualThan"
	case GreaterOrEqualThan:
		s = "GreaterOrEqualThan"
	case LeftShift:
		s = "LeftShift"
	case RightShift:
		s = "RightShift"
	case LogicAnd:
		s = "LogicAnd"
	case LogicOr:
		s = "LogicOr"
	}

	return s
}

var BinaryOperators = map[string]BinaryOperator{
	"+":  Addition,
	"-":  Subtraction,
	"*":  Multiplication,
	"/":  Division,
	"==": Equality,
	"!=": Inequality,
	"<":  LesserThan,
	">":  GreaterThan,
	"<=": LesserOrEqualThan,
	">=": GreaterOrEqualThan,
	"<<": LeftShift,
	">>": RightShift,
	"&&": LogicAnd,
	"||": LogicOr,
}

var BinaryOperatorApplies = map[BinaryOperator]map[Type]bool{
	Addition:           {Integer: true},
	Subtraction:        {Integer: true},
	Multiplication:     {Integer: true},
	Division:           {Integer: true},
	Equality:           {Integer: true, Boolean: true},
	Inequality:         {Integer: true, Boolean: true},
	LesserThan:         {Integer: true},
	GreaterThan:        {Integer: true},
	LesserOrEqualThan:  {Integer: true},
	GreaterOrEqualThan: {Integer: true},
	LeftShift:          {Integer: true},
	RightShift:         {Integer: true},
	LogicAnd:           {Integer: true, Boolean: true},
	LogicOr:            {Integer: true, Boolean: true},
}

type UnaryOperator int

const (
	Negation UnaryOperator = iota
	Inversion
)

func (o UnaryOperator) String() string {
	switch o {
	case Negation:
		return "Negation"
	case Inversion:
		return "Inversion"
	default:
		return "UNKNOWN"
	}
}

var UnaryOperators = map[string]UnaryOperator{
	"!": Negation,
	"-": Inversion,
}

var UnaryOperatorApplies = map[UnaryOperator]map[Type]bool{
	Negation:  {Boolean: true},
	Inversion: {Integer: true},
}
