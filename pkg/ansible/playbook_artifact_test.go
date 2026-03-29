package ansible_test

import (
	"slices"
	"testing"

	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

func TestPlaybookStdoutString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    ansible.PlaybookStdout
		expected string
	}{
		"empty": {
			input:    ansible.PlaybookStdout{},
			expected: "",
		},
		"multiple_lines": {
			input:    ansible.PlaybookStdout{"line1", "line2", "line3"},
			expected: "line1\nline2\nline3",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := test.input.String(); got != test.expected {
				t.Errorf("expected %q, got %q", test.expected, got)
			}
		})
	}
}

func TestParsePlaybookArtifact(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input     []byte
		expected  *ansible.PlaybookArtifact
		expectErr bool
	}{
		"valid": {
			input: []byte(`{"status":"successful","stdout":["line1","line2"]}`),
			expected: &ansible.PlaybookArtifact{
				Status: "successful",
				Stdout: ansible.PlaybookStdout{"line1", "line2"},
			},
		},
		"invalid": {
			input:     []byte(`{invalid`),
			expectErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := ansible.ParsePlaybookArtifact(test.input)

			if test.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Status != test.expected.Status {
				t.Errorf("status: expected %q, got %q", test.expected.Status, got.Status)
			}

			if !slices.Equal(got.Stdout, test.expected.Stdout) {
				t.Errorf("stdout: expected %v, got %v", test.expected.Stdout, got.Stdout)
			}
		})
	}
}
