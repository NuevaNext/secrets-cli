package cmd

import (
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"valid-name", false},
		{"user@example.com", false},
		{"", true},
		{"../traversal", true},
		{"/absolute", true},
		{"\\backslash", true},
		{"-leading-hyphen", true},
		{"name/with/slash", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateName(tt.name); (err != nil) != tt.wantErr {
				t.Errorf("validateName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSecretName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"valid-secret", false},
		{"database/password", false},
		{"deep/path/to/secret", false},
		{"", true},
		{"../traversal", true},
		{"path/../traversal", true},
		{"\\backslash", true},
		{"-leading-hyphen", true},
		{"/leading-slash", true},
		{"trailing-slash/", true},
		{"double//slash", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateSecretName(tt.name); (err != nil) != tt.wantErr {
				t.Errorf("validateSecretName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
