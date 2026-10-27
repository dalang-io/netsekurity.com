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

// defaultSentryDSN points at the Dalang platform's internal Sentry (Mata). A DSN
// is a public client key, not a secret, but every deployment may override it via
// SENTRY_DSN; an empty SENTRY_DSN disables reporting entirely.
const defaultSentryDSN = "https://bd3ff6f325012d29eacd074457aba1ea@mata.dalang.io/8"

// sentryDSN returns the DSN to report to. An explicitly empty .env value or
// environment variable disables reporting (getenv cannot express that: it treats
// empty as unset); with neither set, the project default above is used.
func sentryDSN() string {
	if v, ok := env["SENTRY_DSN"]; ok {
		return v
	}
	if v, ok := os.LookupEnv("SENTRY_DSN"); ok {
		return v
	}
	return defaultSentryDSN
}

// sentryRelease is the release reported to Sentry (set it to the release tag,
// e.g. v2026.10.27.1). Empty is fine: the SDK then omits the release.
func sentryRelease() string { return getenv("SENTRY_RELEASE", "") }

// sentryEnvironment mirrors the ENVIRONMENT setting already used by
// checkSecrets, so events from production are not mixed with local ones.
func sentryEnvironment() string { return getenv("ENVIRONMENT", "development") }

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
