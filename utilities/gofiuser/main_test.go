package main

import (
	"testing"

	"github.com/unifi-go/gofi/utilities/internal/users"
)

func TestValidateModes(t *testing.T) {
	tests := []struct {
		name    string
		list    bool
		del     bool
		wantErr bool
	}{
		{"list only", true, false, false},
		{"del only", false, true, false},
		{"no mode", false, false, true},
		{"both modes", true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateModes(tt.list, tt.del)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateModes(%t, %t) err = %v, wantErr = %t", tt.list, tt.del, err, tt.wantErr)
			}
		})
	}
}

func TestValidateIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier users.DeleteIdentifier
		wantErr    bool
	}{
		{"mac only", users.DeleteIdentifier{MAC: "aa:bb:cc:dd:ee:ff"}, false},
		{"name only", users.DeleteIdentifier{Name: "tapo1"}, false},
		{"none", users.DeleteIdentifier{}, true},
		{"both", users.DeleteIdentifier{MAC: "aa:bb:cc:dd:ee:ff", Name: "tapo1"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIdentifier(tt.identifier)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateIdentifier(%+v) err = %v, wantErr = %t", tt.identifier, err, tt.wantErr)
			}
		})
	}
}
