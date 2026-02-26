package cmd

import (
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "dev", false},
		{"valid email", "alice@example.com", false},
		{"valid email with subdomains", "bob@mail.dev.example.com", false},
		{"valid email with plus", "user+label@example.com", false},
		{"empty name", "", true},
		{"path traversal", "../test", true},
		{"slash in name", "dev/prod", true},
		{"backslash in name", "dev\\prod", true},
		{"leading hyphen", "-flag", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateName(tt.input); (err != nil) != tt.wantErr {
				t.Errorf("validateName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSecretName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid secret", "database/password", false},
		{"valid simple secret", "apikey", false},
		{"empty name", "", true},
		{"path traversal", "../test", true},
		{"backslash", "test\\secret", true},
		{"double slash", "test//secret", true},
		{"leading slash", "/test", true},
		{"trailing slash", "test/", true},
		{"leading hyphen", "-flag", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateSecretName(tt.input); (err != nil) != tt.wantErr {
				t.Errorf("validateSecretName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
