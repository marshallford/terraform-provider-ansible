package ansible

import (
	"crypto"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"testing"

	gossh "golang.org/x/crypto/ssh"
)

func testSSHKeygen(t *testing.T) (string, string) {
	t.Helper()

	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}

	privateKey, err := gossh.MarshalPrivateKey(crypto.PrivateKey(priv), "")
	if err != nil {
		t.Fatal(err)
	}

	publicKey, err := gossh.NewPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}

	return fmt.Sprintf("ssh-ed25519 %s", base64.StdEncoding.EncodeToString(publicKey.Marshal())), string(pem.EncodeToMemory(privateKey))
}

func TestValidateSSHPrivateKey(t *testing.T) {
	t.Parallel()

	_, validKey := testSSHKeygen(t)

	tests := map[string]struct {
		input     string
		expectErr bool
	}{
		"valid": {
			input: validKey,
		},
		"empty": {
			input:     "",
			expectErr: true,
		},
		"invalid": {
			input:     "not-a-key",
			expectErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := ValidateSSHPrivateKey(test.input)

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

func TestValidateSSHPrivateKeyName(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input     string
		expectErr bool
	}{
		"valid": {
			input: "id-ed25519",
		},
		"empty": {
			input:     "",
			expectErr: true,
		},
		"leading_dash": {
			input:     "-bad",
			expectErr: true,
		},
		"trailing_dash": {
			input:     "bad-",
			expectErr: true,
		},
		"invalid_char": {
			input:     "bad_name",
			expectErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := ValidateSSHPrivateKeyName(test.input)

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

func TestValidateSSHKnownHost(t *testing.T) {
	t.Parallel()

	publicKey, _ := testSSHKeygen(t)

	tests := map[string]struct {
		input     string
		expectErr bool
	}{
		"valid": {
			input: fmt.Sprintf("host1 %s", publicKey),
		},
		"empty": {
			input:     "",
			expectErr: true,
		},
		"invalid": {
			input:     "not-a-known-host",
			expectErr: true,
		},
		"multiple_entries": {
			input:     fmt.Sprintf("host1 %s\nhost2 %s", publicKey, publicKey),
			expectErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := ValidateSSHKnownHost(test.input)

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

func TestValidateEnvVarName(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input     string
		expectErr bool
	}{
		"valid": {
			input: "MY_VAR",
		},
		"empty": {
			input:     "",
			expectErr: true,
		},
		"equals": {
			input:     "NOT=VALID",
			expectErr: true,
		},
		"non_ascii": {
			input:     "VAR\x00",
			expectErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := ValidateEnvVarName(test.input)

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

func TestValidateYAML(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input     string
		expectErr bool
	}{
		"valid": {
			input: "key: value",
		},
		"invalid": {
			input:     "key: [",
			expectErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := ValidateYAML(test.input)

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
