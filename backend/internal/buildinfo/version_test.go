package buildinfo

import "testing"

func TestCurrentVersionPrefersLdflags(t *testing.T) {
	prev := Version
	t.Cleanup(func() { Version = prev })
	Version = "v1.2.3-beta.1"
	if got := CurrentVersion(); got != "v1.2.3-beta.1" {
		t.Fatalf("CurrentVersion() = %q", got)
	}
}

func TestIsCanary(t *testing.T) {
	prev := Version
	t.Cleanup(func() { Version = prev })

	cases := []struct {
		version string
		want    bool
	}{
		{"v1.0.0", false},
		{"v1.0.0-beta", false},
		{"canary", true},
		{"canary-abc1234", true},
		{"canary.1", true},
		{"dev", false},
	}
	for _, tc := range cases {
		Version = tc.version
		if got := IsCanary(); got != tc.want {
			t.Fatalf("IsCanary(%q) = %v, want %v", tc.version, got, tc.want)
		}
	}
}
