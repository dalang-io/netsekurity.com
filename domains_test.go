package main

import (
	"strings"
	"testing"
)

func TestIsValidDomain(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"example.com", true},
		{"sub.example.co.id", true},
		{"a.b", true},
		{"example", false},                // no dot
		{"exa", false},                    // too short
		{"-bad.com", false},               // leading hyphen
		{"bad-.com", false},               // trailing hyphen
		{"exa..com", false},               // empty label
		{"", false},                       // empty
		{strings.Repeat("a", 254), false}, // too long
	}
	for _, c := range cases {
		if got := isValidDomain(c.in); got != c.want {
			t.Errorf("isValidDomain(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
