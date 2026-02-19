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
		{"valid_name", false},
		{"", true},
		{"..", true},
		{"path/to", true},
		{"path\\to", true},
		{"-flag", true},
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
		{"valid", false},
		{"valid/path", false},
		{"valid/path/to/secret", false},
		{"", true},
		{"..", true},
		{"path/../to", true},
		{"path\\to", true},
		{"//double", true},
		{"/leading", true},
		{"trailing/", true},
		{"-leading-hyphen", true},
		{"valid/-not-leading-hyphen", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateSecretName(tt.name); (err != nil) != tt.wantErr {
				t.Errorf("validateSecretName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
