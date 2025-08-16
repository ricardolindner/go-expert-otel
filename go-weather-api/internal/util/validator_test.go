package util

import "testing"

func TestIsValidCEP(t *testing.T) {
	tests := []struct {
		name string
		cep  string
		want bool
	}{
		{"valid CEP", "12345678", true},
		{"invalid CEP - short", "1234567", false},
		{"invalid CEP - long", "123456789", false},
		{"invalid CEP - letters", "abcdefgh", false},
		{"invalid CEP - mixed", "1234567a", false},
		{"invalid CEP - empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidCEP(tt.cep); got != tt.want {
				t.Errorf("IsValidCEP(%s) = %v, want %v", tt.cep, got, tt.want)
			}
		})
	}
}
