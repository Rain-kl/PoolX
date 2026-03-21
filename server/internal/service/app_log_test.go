package service

import (
	"ginnexttemplate/internal/model"
	"testing"
)

func TestNormalizeLogLevelDefaultsToInfo(t *testing.T) {
	if got := normalizeLogLevel(""); got != model.AppLogLevelInfo {
		t.Fatalf("normalizeLogLevel empty = %q, want %q", got, model.AppLogLevelInfo)
	}
	if got := normalizeLogLevel("invalid"); got != model.AppLogLevelInfo {
		t.Fatalf("normalizeLogLevel invalid = %q, want %q", got, model.AppLogLevelInfo)
	}
}

func TestNormalizeLogLevelSupportsKnownValues(t *testing.T) {
	cases := []string{
		model.AppLogLevelDebug,
		model.AppLogLevelInfo,
		model.AppLogLevelWarn,
		model.AppLogLevelError,
	}

	for _, input := range cases {
		if got := normalizeLogLevel(input); got != input {
			t.Fatalf("normalizeLogLevel(%q) = %q, want %q", input, got, input)
		}
	}
}
