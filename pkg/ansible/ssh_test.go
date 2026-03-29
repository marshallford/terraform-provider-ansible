package ansible

import (
	"errors"
	"io"
	"slices"
	"strings"
	"testing"
	"testing/iotest"
)

func TestParseKnownHosts(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input     io.Reader
		expected  []KnownHost
		expectErr bool
	}{
		"empty": {
			input:    strings.NewReader(""),
			expected: []KnownHost{},
		},
		"multiple_lines": {
			input:    strings.NewReader("host1 ssh-rsa AAAA...\nhost2 ssh-ed25519 BBBB..."),
			expected: []KnownHost{"host1 ssh-rsa AAAA...", "host2 ssh-ed25519 BBBB..."},
		},
		"scanner_error": {
			input:     iotest.ErrReader(errors.New("read failure")),
			expectErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseKnownHosts(test.input)

			if test.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !slices.Equal(got, test.expected) {
				t.Errorf("expected %v, got %v", test.expected, got)
			}
		})
	}
}

func TestKnownHostsLine(t *testing.T) {
	t.Parallel()

	// Valid ed25519 public key for testing
	validPublicKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl"

	tests := map[string]struct {
		inputAddresses []string
		inputPublicKey string
		expected       string
		expectErr      bool
	}{
		"valid": {
			inputAddresses: []string{"host1:22"},
			inputPublicKey: validPublicKey,
			expected:       "host1 " + validPublicKey,
		},
		"no_addresses": {
			inputAddresses: []string{},
			inputPublicKey: validPublicKey,
			expectErr:      true,
		},
		"invalid_public_key": {
			inputAddresses: []string{"host1:22"},
			inputPublicKey: "not-a-key",
			expectErr:      true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := KnownHostsLine(test.inputAddresses, test.inputPublicKey)

			if test.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != test.expected {
				t.Errorf("expected %q, got %q", test.expected, got)
			}
		})
	}
}

func TestSSHArgs(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input     bool
		expected  string
		expectErr bool
	}{
		"strict": {
			input:    false,
			expected: "-o StrictHostKeyChecking=yes -o UserKnownHostsFile={{ " + SSHKnownHostsFileVar + " }}",
		},
		"accept_new": {
			input:    true,
			expected: "-o StrictHostKeyChecking=accept-new -o UserKnownHostsFile={{ " + SSHKnownHostsFileVar + " }}",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := SSHArgs(test.input); got != test.expected {
				t.Errorf("expected %q, got %q", test.expected, got)
			}
		})
	}
}
