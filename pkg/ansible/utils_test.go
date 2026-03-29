package ansible

import (
	"testing"

	"github.com/spf13/afero"
)

func TestCheckDirectory(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input     string
		expectErr bool
		setup     func(afero.Fs)
	}{
		"valid": {
			input: "/mydir",
			setup: func(fs afero.Fs) {
				fs.MkdirAll("/mydir", 0o755) //nolint:errcheck
			},
		},
		"not_a_directory": {
			input: "/myfile",
			setup: func(fs afero.Fs) {
				afero.WriteFile(fs, "/myfile", []byte(""), 0o644) //nolint:errcheck
			},
			expectErr: true,
		},
		"not_found": {
			input:     "/nonexistent",
			setup:     func(fs afero.Fs) {},
			expectErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			test.setup(fs)

			err := CheckDirectory(fs, test.input)

			if test.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
