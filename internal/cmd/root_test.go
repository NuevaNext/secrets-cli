package cmd

import (
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"valid", false},
		{"valid-name", false},
		{"valid.name", false},
		{"", true},
		{"..", true},
		{"../etc/passwd", true},
		{"/etc/passwd", true},
		{"\\backslashes", true},
		{"-leading-hyphen", true},
	}

	for _, tt := range tests {
		err := validateName(tt.name)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestValidateSecretName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"valid", false},
		{"valid/secret", false},
		{"valid/secret/path", false},
		{"", true},
		{"..", true},
		{"../secret", true},
		{"secret/..", true},
		{"//double/slash", true},
		{"secret//double", true},
		{"\\backslashes", true},
		{"/leading/slash", true},
		{"trailing/slash/", true},
		{"-leading-hyphen", true},
		{"secret/-leading-hyphen-in-part", false}, // This is okay as it's not leading the whole argument
	}

	for _, tt := range tests {
		err := validateSecretName(tt.name)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateSecretName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}
