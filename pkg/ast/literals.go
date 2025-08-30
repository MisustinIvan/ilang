package ast

import (
	"fmt"
	"strconv"
	"strings"
)

// contains functions that help determining types of literal values

func isInteger(x string) bool {
	_, err := strconv.ParseInt(x, 10, 64)
	return err == nil
}

func isFloat(x string) bool {
	_, err := strconv.ParseFloat(x, 10)
	return err == nil
}

func isBoolean(x string) bool {
	return x == "true" || x == "false"
}

func isString(x string) bool {
	return strings.HasPrefix(x, "\"") && strings.HasSuffix(x, "\"")
}

func isUnit(x string) bool {
	return x == "unit"
}

func LiteralType(l string) (Type, error) {
	switch {
	case isInteger(l):
		return Integer, nil
	case isFloat(l):
		return Float, nil
	case isBoolean(l):
		return Boolean, nil
	case isString(l):
		return String, nil
	case isUnit(l):
		return Unit, nil
	default:
		return Unit, fmt.Errorf("literal \"%s\" has unknown type", l)
	}
}
