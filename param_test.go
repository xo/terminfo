package terminfo

import "testing"

func TestPrintfFormattedValues(t *testing.T) {
	for _, tt := range []struct {
		format string
		params []any
		want   string
	}{
		{"%p1%02d", []any{3}, "03"},
		{"%p1%02dX", []any{3}, "03X"},
		{"%p1%02d;%p2%03d!", []any{3, 4}, "03;004!"},
		{"%p1%:04x", []any{15}, "000f"},
		{"%p1%:-4s!", []any{"ab"}, "ab  !"},
		{"%p1%3c!", []any{byte('a')}, "  a!"},
	} {
		t.Run(tt.format, func(t *testing.T) {
			if got := Printf([]byte(tt.format), tt.params...); got != tt.want {
				t.Fatalf("Printf = %q, want %q", got, tt.want)
			}
		})
	}
}
