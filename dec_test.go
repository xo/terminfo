package terminfo

import (
	"bytes"
	"testing"
)

func TestCanonicalizeAscChars(t *testing.T) {
	for _, tt := range []struct {
		name        string
		input, want []byte
	}{
		{"mapped", []byte("q-xl"), []byte("q-xl")},
		{"sort by source", []byte("x|q-"), []byte("q-x|")},
		{"duplicate source", []byte("q-q=x|"), []byte("q-x|")},
		{"shared destination", []byte("x-q-"), []byte("q-x-")},
		{"identity", []byte("xxqq"), []byte("qqxx")},
		{"empty", nil, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := canonicalizeAscChars(tt.input); !bytes.Equal(got, tt.want) {
				t.Fatalf("canonicalizeAscChars(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
