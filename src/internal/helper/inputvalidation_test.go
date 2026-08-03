package helper

import (
	"os"
	"testing"

	"github.com/halsten-dev/lokyn"
)

func TestMain(m *testing.M) {
	lokyn.Init()
	lokyn.SetLanguage("en")
	os.Exit(m.Run())
}

func TestSignedNumericalValidate(t *testing.T) {
	valid := []string{"", "0", "5", "-5", "-"}

	for _, s := range valid {
		if err := SignedNumericalValidate(s); err != nil {
			t.Errorf("SignedNumericalValidate(%q) = %v, want nil", s, err)
		}
	}

	invalid := []string{"a", "5-", "--5", "1.5"}

	for _, s := range invalid {
		if err := SignedNumericalValidate(s); err == nil {
			t.Errorf("SignedNumericalValidate(%q) = nil, want an error", s)
		}
	}
}
