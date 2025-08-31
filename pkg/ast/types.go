package ast

import "fmt"

// Contains very basic definitions of types, will be reworked

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

func ParseType(val string) (Type, error) {
	switch val {
	case "int":
		return Integer, nil
	case "float":
		return Float, nil
	case "bool":
		return Boolean, nil
	case "string":
		return String, nil
	case "unit":
		return Unit, nil
	default:
		return Undefined, fmt.Errorf("invalid type: %s", val)
	}
}
