package terminfo

import (
	"strings"
	"testing"
)

func TestPrintfFormattedValues(t *testing.T) {
	for _, test := range []struct {
		format string
		params []any
		want   string
	}{
		{"%p1%02d", []any{3}, "03"},
		{"%p1%02dX", []any{3}, "03X"},
		{"%p1%02d;%p2%03d!", []any{3, 4}, "03;004!"},
		{"%p1%:04x", []any{15}, "000f"},
		{"%p1%:-4s!", []any{"ab"}, "ab  !"},
		{"%p1%3c!", []any{byte('a')}, "a!"},
		// the values below are the output of tput in ncurses 6.6
		{"%p1%:-4d!", []any{97}, "97  !"},
		{"%p1%:#-5x!", []any{97}, "0x61 !"},
		{"%p1%#o!", []any{97}, "0141!"},
		{"%p1%#x!", []any{97}, "0x61!"},
		{"%p1% d!", []any{97}, " 97!"},
		{"%p1%.3d!", []any{97}, "097!"},
		{"%p1%:+d!", []any{97}, "d!"},
		{"%p1%-d!", []any{97}, "d!"},
		{"%p1%5.2.3d!", []any{97}, "97!"},
		{"%p1%10001d!", []any{97}, "97!"},
		{"%p1%10000d!", []any{97}, strings.Repeat(" ", 9998) + "97!"},
		{"%p1%:%%999999XX", []any{97}, "%61X"},
	} {
		t.Run(test.format, func(t *testing.T) {
			if got := Printf([]byte(test.format), test.params...); got != test.want {
				t.Fatalf("Printf = %q, want %q", got, test.want)
			}
		})
	}
}

// FuzzPrintf checks that Printf does not panic, and that a width cannot make
// the output much longer than ncurses allows.
func FuzzPrintf(f *testing.F) {
	for _, z := range []string{"%p1%d", "%p1%:-4d", "%:%%999999XX", "%p1%p2%+%d", "%?%p1%t;%e%;"} {
		f.Add(z)
	}
	f.Fuzz(func(t *testing.T, z string) {
		if got := Printf([]byte(z), 97, "ab"); len(got) > 10000*len(z) {
			t.Errorf("Printf(%q) returned %d bytes", z, len(got))
		}
	})
}
