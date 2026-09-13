package terminfo

import "testing"

func TestPrintfFormattedValues(t *testing.T) {
	for _, tt := range []struct {
		format string
		params []interface{}
		want   string
	}{
		{"%p1%02d", []interface{}{3}, "03"},
		{"%p1%02dX", []interface{}{3}, "03X"},
		{"%p1%02d;%p2%03d!", []interface{}{3, 4}, "03;004!"},
		{"%p1%:04x", []interface{}{15}, "000f"},
		{"%p1%:-4s!", []interface{}{"ab"}, "ab  !"},
		{"%p1%3c!", []interface{}{byte('a')}, "  a!"},
	} {
		t.Run(tt.format, func(t *testing.T) {
			if got := Printf([]byte(tt.format), tt.params...); got != tt.want {
				t.Fatalf("Printf = %q, want %q", got, tt.want)
			}
		})
	}
}
