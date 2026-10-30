package terminfo

import "testing"

func TestColorLevelFromEnvMaxColors(t *testing.T) {
	tests := []struct {
		term string
		exp  ColorLevel
	}{
		{"dumb", ColorLevelNone},
		{"vt100", ColorLevelNone},
		{"xterm", ColorLevelBasic},
		{"xterm-16color", ColorLevelBasic},
		{"xterm-88color", ColorLevelBasic},
		{"xterm-256color", ColorLevelHundreds},
	}
	for _, test := range tests {
		t.Run(test.term, func(t *testing.T) {
			for _, k := range []string{"COLORTERM", "FORCE_COLOR", "TERM_PROGRAM", "TERM_PROGRAM_VERSION"} {
				t.Setenv(k, "")
			}
			t.Setenv("TERM", test.term)
			c, err := ColorLevelFromEnv()
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if c != test.exp {
				t.Errorf("expected %s, got: %s", test.exp, c)
			}
		})
	}
}
