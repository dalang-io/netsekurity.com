package main

import "testing"

func TestParseFloat(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"16500", 16500},
		{"0", 0},
		{"3.14", 3.14},
		{"", 0},
		{"abc", 0},
	}
	for _, c := range cases {
		if got := parseFloat(c.in); got != c.want {
			t.Errorf("parseFloat(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
