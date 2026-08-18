package main

import (
	"bufio"
	"os"
	"strings"
)

// env holds variables loaded from the .env file (if present). Credentials can
// be reused from the existing dalang.io / api.dalang.io .env files.
var env = map[string]string{}

// loadEnv parses a simple KEY=VALUE .env file into the env map. Real env vars
// always take precedence over the file.
func loadEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		i := strings.Index(line, "=")
		if i < 0 {
			continue
		}
		k := strings.TrimSpace(line[:i])
		v := strings.Trim(strings.TrimSpace(line[i+1:]), `"'`)
		env[k] = v
	}
}

// getenv returns the .env value, falling back to an actual environment variable,
// then to a default.
func getenv(key, def string) string {
	if v, ok := env[key]; ok && v != "" {
		return v
	}
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
