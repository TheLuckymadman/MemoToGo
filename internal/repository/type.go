package repository

import (
	"fmt"
	"strings"
)

type DBInitMode int

const (
	External DBInitMode = iota
	Internal
	ForceInternal
)

func (m *DBInitMode) UnmarshalText(text []byte) error {
	s := strings.ToLower(strings.TrimSpace(string(text)))
	switch s {
	case "external", "0":
		*m = External
	case "internal", "1":
		*m = Internal
	case "forceinternal", "force", "2":
		*m = ForceInternal
	}
	return nil
}

func (m *DBInitMode) Parse(s string) error {
	return m.UnmarshalText([]byte(s))
}

func (m *DBInitMode) String() (string, error) {
	switch *m {
	case External:
		return "external", nil
	case Internal:
		return "internal", nil
	case ForceInternal:
		return "forcereset", nil
	default:
		return "", fmt.Errorf("incorrect db init mode: %d", *m)
	}
}
