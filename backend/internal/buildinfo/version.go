// Package buildinfo exposes release metadata injected at link time.
package buildinfo

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// Version is set by release builds via:
//
//	-ldflags "-X github.com/Rain-kl/Foam/backend/internal/buildinfo.Version=v1.2.3"
//
// Tag releases use the git tag (e.g. v1.0.0, v1.0.0-beta).
// Untagged CI builds inject a canary version (e.g. canary-a1b2c3d).
var Version string

// BuildTime is the UTC build timestamp (RFC3339), injected as:
//
//	-X github.com/Rain-kl/Foam/backend/internal/buildinfo.BuildTime=...
var BuildTime string

// CanaryPrefix is the version prefix used when no release tag is available.
const CanaryPrefix = "canary"

// CurrentVersion returns the running binary's version.
// Order: ldflags Version → FOAM_VERSION → VERSION file → module build info → "dev".
func CurrentVersion() string {
	if value := cleanVersion(Version); value != "" {
		return value
	}
	if value := cleanVersion(os.Getenv("FOAM_VERSION")); value != "" {
		return value
	}
	candidates := []string{"VERSION", filepath.Join("..", "VERSION")}
	if executable, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(executable), "VERSION"))
	}
	for _, candidate := range candidates {
		if data, err := os.ReadFile(candidate); err == nil { //nolint:gosec // fixed local VERSION candidates
			if value := cleanVersion(string(data)); value != "" {
				return value
			}
		}
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

// CurrentBuildTime returns the injected build time, or empty when unset.
func CurrentBuildTime() string {
	return strings.TrimSpace(BuildTime)
}

// IsCanary reports whether the effective version is a canary build.
func IsCanary() bool {
	v := strings.ToLower(CurrentVersion())
	return v == CanaryPrefix || strings.HasPrefix(v, CanaryPrefix+"-") || strings.HasPrefix(v, CanaryPrefix+".")
}

func cleanVersion(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 128 || strings.IndexFunc(value, func(r rune) bool { return r < 0x20 || r == 0x7f }) >= 0 {
		return ""
	}
	return value
}
