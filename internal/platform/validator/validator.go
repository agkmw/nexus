package validator

import (
	"regexp"
	"slices"
)

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) AddErrors(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddErrors(key, message)
	}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// =============================================================================

func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool, len(values))

	for _, v := range values {
		uniqueValues[v] = true
	}

	return len(uniqueValues) == len(values)
}

func IsPermitted[T comparable](permittedValues []T, value T) bool {
	return slices.Contains(permittedValues, value)
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
