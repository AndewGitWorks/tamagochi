package domain

import (
	"errors"
	"regexp"
)

var (
	ErrNotFound    = errors.New("entity was not found")
	ErrConflict    = errors.New("entity conflicts with existing data")
	ErrValidation  = errors.New("invalid entity data")
	ErrOrderLocked = errors.New("only pending orders can be edited")
)

// Prices are decimal strings so JSON and SQL never round them through float64.
var moneyPattern = regexp.MustCompile(`^(0|[1-9][0-9]{0,7})(\.[0-9]{1,2})?$`)

func ValidPrice(value string) bool { return moneyPattern.MatchString(value) }
