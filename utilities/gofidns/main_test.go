package main

import (
	"testing"

	"github.com/unifi-go/gofi/utilities/internal/dns"
)

func TestValidateModes(t *testing.T) {
	tests := []struct {
		name    string
		get     bool
		del     bool
		wantErr bool
	}{
		{"get only", true, false, false},
		{"del only", false, true, false},
		{"no mode", false, false, true},
		{"both modes", true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateModes(tt.get, tt.del)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateModes(%t, %t) err = %v, wantErr = %t", tt.get, tt.del, err, tt.wantErr)
			}
		})
	}
}

func TestValidateIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier dns.DeleteIdentifier
		wantErr    bool
	}{
		{"id only", dns.DeleteIdentifier{ID: "dns1"}, false},
		{"name only", dns.DeleteIdentifier{Name: "a.example.com"}, false},
		{"ip only", dns.DeleteIdentifier{IP: "192.168.1.1"}, false},
		{"none", dns.DeleteIdentifier{}, true},
		{"id and name", dns.DeleteIdentifier{ID: "dns1", Name: "a.example.com"}, true},
		{"all three", dns.DeleteIdentifier{ID: "dns1", Name: "a.example.com", IP: "192.168.1.1"}, true},
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
