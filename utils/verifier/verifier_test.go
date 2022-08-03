package verifier

import (
	"testing"
)

func TestPhone(t *testing.T) {
	for _, unit := range [] struct {
		phone    string
		expected bool
	}{
		{"13344445555", true},
		{"15500004444", true},
		{"18644449999", true},
		{"17612345678", true},
		{"13300004444", true},

		{"12300004444", false},
		{"133123456789", false},
		{"62548745", false},
		{"asdasfsdfcd", false},
		{"3.141592653", false},
		{"11111111111", false},
		{"99999999999", false},
	} {
		if actually := IsPhone(unit.phone); actually != unit.expected {
			t.Error(unit.phone + " is not expected!")
		}
	}
}
