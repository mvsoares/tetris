package version

import (
	"strings"
	"testing"
)

func TestVersionFull(t *testing.T) {
	if Version != "1.0.0" {
		t.Fatalf("expected Version to be 1.0.0, got %s", Version)
	}

	full := Full()
	if !strings.HasPrefix(full, "v1.0.0") {
		t.Fatalf("expected Full() to start with 'v1.0.0', got: %s", full)
	}

	Commit = "abc1234"
	Date = "2026-09-29T12:00:00Z"
	fullWithMeta := Full()
	if !strings.Contains(fullWithMeta, "abc1234") || !strings.Contains(fullWithMeta, "2026-09-29T12:00:00Z") {
		t.Fatalf("expected Full() to contain commit and date metadata, got: %s", fullWithMeta)
	}
}
