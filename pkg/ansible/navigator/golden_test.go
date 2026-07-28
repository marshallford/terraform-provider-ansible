package navigator

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//nolint:gochecknoglobals
var update = flag.Bool("update", false, "update golden files")

func assertGolden(t *testing.T, name string, got string) {
	t.Helper()

	path := filepath.Join("testdata", name)

	if !strings.HasSuffix(got, "\n") {
		got += "\n"
	}

	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatalf("failed to create golden directory: %v", err)
		}

		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}

		return
	}

	want, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		t.Fatalf("failed to read golden file %s (run with -update to create it): %v", path, err)
	}

	if string(want) != got {
		t.Errorf("golden file %s mismatch\n--- want ---\n%s\n--- got ---\n%s", path, want, got)
	}
}

func assertGoldenLines(t *testing.T, name string, got []string) {
	t.Helper()

	assertGolden(t, name, strings.Join(got, "\n"))
}
