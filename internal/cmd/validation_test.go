package cmd

import (
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"dev", false},
		{"production", false},
		{"alice@example.com", false},
		{"", true},
		{"../secrets", true},
		{"/etc/passwd", true},
		{"\\windowspath", true},
		{"-arg", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateName(tt.name); (err != nil) != tt.wantErr {
				t.Errorf("validateName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestValidateSecretName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"password", false},
		{"db/password", false},
		{"cloud/api/key", false},
		{"", true},
		{"../secret", true},
		{"db/../secret", true},
		{"db\\password", true},
		{"db//password", true},
		{"/db/password", true},
		{"db/password/", true},
		{"-secret", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateSecretName(tt.name); (err != nil) != tt.wantErr {
				t.Errorf("validateSecretName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}
