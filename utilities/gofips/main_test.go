package main

import "testing"

func TestCheckStrayFlags(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no positionals", nil, false},
		{"filename only", []string{"hosts.conf"}, false},
		{"stray long flag after filename", []string{"hosts.conf", "--dry-run"}, true},
		{"stray short flag after filename", []string{"hosts.conf", "-k"}, true},
		{"bare dash allowed", []string{"-"}, false},
		{"isc fragment not a flag", []string{"host dev { hardware ethernet aa:bb:cc:dd:ee:ff; }"}, false},
	}
	for _, c := range cases {
		err := checkStrayFlags(c.args)
		if (err != nil) != c.wantErr {
			t.Errorf("%s: checkStrayFlags(%v) err = %v, wantErr = %v", c.name, c.args, err, c.wantErr)
		}
	}
}
