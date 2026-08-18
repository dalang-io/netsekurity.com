package main

import (
	"crypto/subtle"
	"strconv"
)

// parseFloat parses a string as float64, returning def on failure.
func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// secureCompare is a constant-time string comparison for bearer secrets
// (the BOT_AUTH_TOKEN). Avoids timing side channels.
func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
