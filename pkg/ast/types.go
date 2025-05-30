package ast

type Type int

const (
	Undefined Type = iota
	Float
	Boolean
	String
	Unit
	Integer
)

func (t Type) Size() int {
	switch t {
	case Integer:
		return 8
	case Float:
		return 8
	case Boolean:
		return 8
	case String:
		// pointer
		return 8
	default:
		return 0
	}
}

func (t Type) String() string {
	switch t {
	case Undefined:
		return "undefined"
	case Integer:
		return "int"
	case Float:
		return "float"
	case Boolean:
		return "bool"
	case String:
		return "string"
	case Unit:
		return "unit"
	default:
		return "UNKNOWN"
	}
}

func ParseType(val string) (Type, bool) {
	switch val {
	case "int":
		return Integer, true
	case "float":
		return Float, true
	case "bool":
		return Boolean, true
	case "string":
		return String, true
	case "unit":
		return Unit, true
	default:
		return Type(-1), false
	}
}
