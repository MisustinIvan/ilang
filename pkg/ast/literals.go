package ast

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

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

func LiteralType(l string) Type {
	switch {
	case isInteger(l):
		return Integer
	case isFloat(l):
		return Float
	case isBoolean(l):
		return Boolean
	case isString(l):
		return String
	case isUnit(l):
		return Unit
	default:
		fmt.Printf("Literal %s has unknown type\n", l)
		os.Exit(-1)
		return Unit
	}
}
