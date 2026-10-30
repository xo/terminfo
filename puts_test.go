package terminfo

import (
	"bytes"
	"strings"
	"testing"
)

func TestPuts(t *testing.T) {
	const flash = "\x1b[?5h$<100/>\x1b[?5l"
	nul := func(n int) string { return strings.Repeat("\x00", n) }
	tests := []struct {
		name  string
		bools map[int]bool
		nums  map[int]int
		strs  map[int][]byte
		s     string
		lines int
		baud  int
		exp   string
	}{
		{name: "no marker", s: "abc", baud: 9000, exp: "abc"},
		{name: "optional delay", s: "a$<5>b", baud: 9000, exp: "ab"},
		{name: "mandatory delay", s: "a$<5/>b", baud: 9000, exp: "a" + nul(5) + "b"},
		{name: "mandatory delay at baud 0", s: "a$<5/>b", exp: "ab"},
		{name: "delay for each line", s: "$<2*/>", lines: 3, baud: 9000, exp: nul(6)},
		{name: "star and slash in any order", s: "$<2/*>", lines: 3, baud: 9000, exp: nul(6)},
		{name: "decimal delay", s: "$<1.59/>", baud: 90000, exp: nul(10)},
		{name: "delay that starts with a dot", s: "a$<.5/>b", baud: 90000, exp: "ab"},
		{name: "zero delay", s: "a$<0/>b", baud: 9000, exp: "ab"},
		{
			name: "flash always delays",
			strs: map[int][]byte{FlashScreen: []byte(flash)},
			s:    flash, baud: 9600,
			exp: "\x1b[?5h" + nul(106) + "\x1b[?5l",
		},
		{
			name: "bell always delays",
			strs: map[int][]byte{Bell: []byte("\a$<10>")},
			s:    "\a$<10>", baud: 9000,
			exp: "\a" + nul(10),
		},
		{
			name: "at the padding baud rate",
			nums: map[int]int{PaddingBaudRate: 9000},
			s:    "a$<5>b", baud: 9000,
			exp: "a" + nul(5) + "b",
		},
		{
			name: "below the padding baud rate",
			nums: map[int]int{PaddingBaudRate: 9600},
			s:    "a$<5>b", baud: 9000,
			exp: "ab",
		},
		{
			name:  "xon/xoff turns off the optional delay",
			bools: map[int]bool{XonXoff: true},
			nums:  map[int]int{PaddingBaudRate: 1200},
			s:     "a$<5>b", baud: 9000,
			exp: "ab",
		},
		{
			name: "pad character",
			strs: map[int][]byte{PadChar: []byte("*")},
			s:    "a$<3/>b", baud: 9000,
			exp: "a***b",
		},
		{name: "delay limit", s: "$<40000/>", baud: 9, exp: nul(30)},
		{name: "negative baud", s: "a$<5/>b", baud: -9000, exp: "ab"},
		{name: "dollar without a marker", s: "a$b$", baud: 9000, exp: "a$b$"},
		{name: "marker without a number", s: "a$<x>b", baud: 9000, exp: "a$<x>b"},
		{name: "marker without an end", s: "a$<5/b", baud: 9000, exp: "a$<5/b"},
		{name: "marker at the end", s: "a$<", baud: 9000, exp: "a$<"},
		{
			name:  "no pad character sleeps",
			bools: map[int]bool{NoPadChar: true},
			s:     "a$<1/>b$<1/>c", baud: 9000,
			exp: "abc",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ti := &Terminfo{Bools: test.bools, Nums: test.nums, Strings: test.strs}
			var buf bytes.Buffer
			n, err := ti.Puts(&buf, test.s, test.lines, test.baud)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if got := buf.String(); got != test.exp {
				t.Errorf("expected %q, got: %q", test.exp, got)
			}
			if n != buf.Len() {
				t.Errorf("expected %d bytes written, got: %d", buf.Len(), n)
			}
		})
	}
}
