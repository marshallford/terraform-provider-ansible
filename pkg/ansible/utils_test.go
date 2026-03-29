package ansible_test

import (
	"testing"

	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
	"github.com/spf13/afero"
)

func TestCheckDirectory(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input     string
		expectErr bool
		setup     func(*testing.T, afero.Fs)
	}{
		"valid": {
			input: "/mydir",
			setup: func(t *testing.T, fs afero.Fs) {
				t.Helper()

				if err := fs.MkdirAll("/mydir", 0o755); err != nil {
					t.Fatal(err)
				}
			},
		},
		"not_a_directory": {
			input: "/myfile",
			setup: func(t *testing.T, fs afero.Fs) {
				t.Helper()

				if err := afero.WriteFile(fs, "/myfile", []byte(""), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			expectErr: true,
		},
		"not_found": {
			input:     "/nonexistent",
			setup:     func(_ *testing.T, _ afero.Fs) {},
			expectErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			test.setup(t, fs)

			err := ansible.CheckDirectory(fs, test.input)

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
