package main

import "strconv"

// parseFloat parses a string as float64, returning def on failure.
func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
