package terminfo

import (
	"bytes"
	"errors"
	"slices"
	"testing"
)

func TestCanonicalizeAscChars(t *testing.T) {
	for _, test := range []struct {
		name        string
		input, want []byte
	}{
		{"mapped", []byte("q-xl"), []byte("q-xl")},
		{"sort by source", []byte("x|q-"), []byte("q-x|")},
		{"duplicate source", []byte("q-q=x|"), []byte("q-x|")},
		{"shared destination", []byte("x-q-"), []byte("q-x-")},
		{"identity", []byte("xxqq"), []byte("qqxx")},
		{"empty", nil, nil},
		{"odd length", []byte("q-x"), []byte("q-")},
		{"single byte", []byte("q"), nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := canonicalizeAscChars(test.input); !bytes.Equal(got, test.want) {
				t.Fatalf("canonicalizeAscChars(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestDecodeExtendedCancelledString(t *testing.T) {
	// An extended header's offset field counts only the offsets that are used
	// (ie, that point into the string data table), so it is less than the
	// number of offsets actually stored whenever an extended string cap is
	// absent or cancelled.
	buf := buildExtended(t,
		"t|test",
		[]int{-2, 0, 0, 3}, // cancelled cap, one cap, two names
		"\x1b[1m\x00XA\x00XB\x00",
		3, // offsets used: the one cap and the two names
	)
	ti, err := Decode(buf)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if exp, got := []string{"t", "test"}, ti.Names; !slices.Equal(got, exp) {
		t.Errorf("expected names %q, got: %q", exp, got)
	}
	if _, ok := ti.ExtStrings[0]; ok {
		t.Errorf("expected extended string 0 to be absent, got: %q", ti.ExtStrings[0])
	}
	if !ti.ExtStringsM[0] {
		t.Errorf("expected extended string 0 to be missing")
	}
	if ti.ExtStringsM[1] {
		t.Errorf("expected extended string 1 to not be missing")
	}
	if exp, got := "\x1b[1m", string(ti.ExtStrings[1]); got != exp {
		t.Errorf("expected extended string 1 to be %q, got: %q", exp, got)
	}
	for i, exp := range []string{"XA", "XB"} {
		if got := string(ti.ExtStringNames[i]); got != exp {
			t.Errorf("expected extended string name %d to be %q, got: %q", i, exp, got)
		}
	}
}

// buildExtended builds a terminfo file that declares no regular caps, and only
// extended string caps described by idx, data and used.
func buildExtended(t *testing.T, names string, idx []int, data string, used int) []byte {
	t.Helper()
	addShort := func(buf []byte, v int) []byte {
		return append(buf, byte(v), byte(v>>8)) //nolint:gosec // the low 16 bits, little endian
	}
	var buf []byte
	for _, v := range []int{magic, len(names) + 1, 0, 0, 0, 0} {
		buf = addShort(buf, v)
	}
	buf = append(buf, names...)
	buf = append(buf, 0)
	if len(buf)%2 != 0 {
		buf = append(buf, 0)
	}
	// extended header: bools, nums, strings, offsets used, table size
	for _, v := range []int{0, 0, len(idx) / 2, used, len(data)} {
		buf = addShort(buf, v)
	}
	for _, v := range idx {
		buf = addShort(buf, v)
	}
	return append(buf, data...)
}

// buildMain builds a terminfo file from the five counts of the main header and
// the bytes that follow the header.
func buildMain(counts [5]int, body []byte) []byte {
	var buf []byte
	for _, v := range append([]int{magic}, counts[:]...) {
		buf = append(buf, byte(v), byte(v>>8)) //nolint:gosec // the low 16 bits, little endian
	}
	return append(buf, body...)
}

func TestDecodeOddAscChars(t *testing.T) {
	// The string capability at AcsChars holds three bytes, so the last byte
	// has no partner. Decode must not panic.
	idx := make([]byte, 2*(AcsChars+1))
	for i := range AcsChars {
		idx[2*i], idx[2*i+1] = 0xff, 0xff // absent
	}
	table := []byte("q-x\x00")
	names := []byte("t\x00")
	body := append(append(names, idx...), table...)
	buf := buildMain([5]int{len(names), 0, 0, AcsChars + 1, len(table)}, body)
	ti, err := Decode(buf)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if exp, got := "q-", string(ti.Strings[AcsChars]); got != exp {
		t.Errorf("expected acsc %q, got: %q", exp, got)
	}
}

func TestDecodeNegativeHeaderCount(t *testing.T) {
	for _, test := range []struct {
		name   string
		counts [5]int
	}{
		{"name size", [5]int{-1, 0, 0, 0, 0}},
		{"bool count", [5]int{2, -1, 0, 0, 0}},
		{"num count", [5]int{2, 0, -25300, 0, 0}},
		{"string count", [5]int{2, 0, 0, -1, 0}},
		{"table size", [5]int{2, 0, 0, 0, -1}},
	} {
		t.Run(test.name, func(t *testing.T) {
			buf := buildMain(test.counts, make([]byte, 64))
			if _, err := Decode(buf); !errors.Is(err, ErrInvalidHeader) {
				t.Fatalf("expected %v, got: %v", ErrInvalidHeader, err)
			}
		})
	}
}

func TestReadBytesNegative(t *testing.T) {
	d := &decoder{buf: make([]byte, 8), n: 8, pos: 4}
	if _, err := d.readBytes(-1); !errors.Is(err, ErrUnexpectedFileEnd) {
		t.Fatalf("expected %v, got: %v", ErrUnexpectedFileEnd, err)
	}
}
