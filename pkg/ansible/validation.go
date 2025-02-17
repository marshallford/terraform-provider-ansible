package ansible

import (
	"errors"
	"fmt"
	"strings"
	"time"
	_ "time/tzdata" // embedded copy of the timezone database
	"unicode"

	"github.com/containers/image/v5/docker/reference"
	jq "github.com/itchyny/gojq"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
)

func ValidateSSHPrivateKey(privateKey string) error {
	if len(privateKey) == 0 {
		return fmt.Errorf("%w, SSH private key must not be empty", ErrValidation)
	}

	_, err := ssh.ParseRawPrivateKey([]byte(privateKey))

	var passphraseErr *ssh.PassphraseMissingError
	if errors.As(err, &passphraseErr) {
		return fmt.Errorf("%w, SSH private key must be unencrypted (no passphrase), %w", ErrValidation, err)
	}

	if err != nil {
		return fmt.Errorf("%w, SSH private key must be a RSA, DSA, ECDSA, or Ed25519 private key formatted as PKCS#1, PKCS#8, OpenSSL, or OpenSSH", ErrValidation)
	}

	return nil
}

func ValidateSSHPrivateKeyName(privateKeyName string) error {
	if len(privateKeyName) == 0 {
		return fmt.Errorf("%w, SSH private key name cannot be empty", ErrValidation)
	}

	if strings.HasPrefix(privateKeyName, "-") || strings.HasSuffix(privateKeyName, "-") {
		return fmt.Errorf("%w, SSH private key name cannot start or end with a dash", ErrValidation)
	}

	for _, character := range privateKeyName {
		if !unicode.IsLetter(character) && !unicode.IsDigit(character) && character != '-' {
			return fmt.Errorf("%w, SSH private key name can only contain letters (A-Z, a-z), numbers (0-9), and dashes (-)", ErrValidation)
		}
	}

	return nil
}

func ValidateSSHKnownHost(knownHost string) error {
	if len(knownHost) == 0 {
		return fmt.Errorf("%w, SSH known host must not be empty", ErrValidation)
	}

	_, _, _, _, rest, err := ssh.ParseKnownHosts([]byte(knownHost)) //nolint:dogsled
	if err != nil {
		return fmt.Errorf("%w, failed to parse SSH known host, %w", ErrValidation, err)
	}

	if len(rest) > 0 {
		return fmt.Errorf("%w, must not include multiple SSH known host entries or additional data", ErrValidation)
	}

	return nil
}

func ValidateEnvVarName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("%w, environment variable name must not be empty", ErrValidation)
	}

	for _, r := range name {
		if r > unicode.MaxASCII || !unicode.IsPrint(r) || r == '=' {
			return fmt.Errorf("%w, environment variable name must consist only of printable ASCII characters other than '='", ErrValidation)
		}
	}

	return nil
}

func ValidateYAML(value string) error {
	var output any
	if err := yaml.Unmarshal([]byte(value), &output); err != nil {
		return fmt.Errorf("%w, failed to deserialize YAML, %w", ErrValidation, err)
	}

	if _, err := yaml.Marshal(output); err != nil {
		return fmt.Errorf("%w, failed to serialize YAML, %w", ErrValidation, err)
	}

	return nil
}

func ValidateIANATimezone(timezone string) error {
	if len(timezone) == 0 {
		return fmt.Errorf("%w, IANA time zone must not be empty", ErrValidation)
	}

	if timezone == "local" {
		return nil
	}

	if _, err := time.LoadLocation(timezone); err != nil {
		return fmt.Errorf("%w, IANA time zone not found, %w", ErrValidation, err)
	}

	return nil
}

func ValidateJQFilter(filter string) error {
	if _, err := jq.Parse(filter); err != nil {
		return fmt.Errorf("%w, failed to parse JQ filter, %w", ErrValidation, err)
	}

	return nil
}

func ValidateContainerImageName(image string) error {
	if len(image) == 0 {
		return fmt.Errorf("%w, container image name must not be empty", ErrValidation)
	}

	if _, err := reference.ParseNormalizedNamed(image); err != nil {
		return fmt.Errorf("%w, failed to parse container image name, %w", ErrValidation, err)
	}

	return nil
}
