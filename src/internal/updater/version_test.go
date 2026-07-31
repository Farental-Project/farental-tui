package updater

import "testing"

func TestCompatible(t *testing.T) {
	tests := []struct {
		client string
		compat string
		want   bool
	}{
		{"1.1.0", "1.1", true},
		{"1.1.7", "1.1", true},
		// The bug the old strings.HasPrefix gate had: "1.10.0" starts with
		// "1.1" but is a different minor version.
		{"1.10.0", "1.1", false},
		{"1.2.0", "1.1", false},
		{"2.1.0", "1.1", false},
		{"1.1.0", "1.1.0", true},
		{"", "1.1", false},
		{"1.1.0", "", false},
		{"nonsense", "1.1", false},
		{"1.1.0", "nonsense", false},
	}

	for _, tt := range tests {
		if got := Compatible(tt.client, tt.compat); got != tt.want {
			t.Errorf("Compatible(%q, %q) = %v, want %v", tt.client, tt.compat, got, tt.want)
		}
	}
}

func TestNewer(t *testing.T) {
	tests := []struct {
		candidate string
		current   string
		want      bool
	}{
		{"1.1.1", "1.1.0", true},
		{"1.2.0", "1.1.9", true},
		{"2.0.0", "1.9.9", true},
		{"1.1.0", "1.1.0", false},
		{"1.1.0", "1.1.1", false},
		{"1.9.0", "1.10.0", false},
		{"1.10.0", "1.9.0", true},
		{"", "1.1.0", false},
		{"1.1.0", "", false},
	}

	for _, tt := range tests {
		if got := Newer(tt.candidate, tt.current); got != tt.want {
			t.Errorf("Newer(%q, %q) = %v, want %v", tt.candidate, tt.current, got, tt.want)
		}
	}
}
