package ansible

import (
	"bufio"
	"fmt"
	"io"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type PrivateKey struct {
	Name string
	Data string
}

type KnownHost = string

func ParseKnownHosts(r io.Reader) ([]KnownHost, error) {
	knownHosts := make([]KnownHost, 0)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		knownHosts = append(knownHosts, scanner.Text())
	}

	if scanner.Err() != nil {
		return nil, fmt.Errorf("failed to parse known hosts, %w", scanner.Err())
	}

	return knownHosts, nil
}

func KnownHostsLine(addresses []string, publicKey string) (string, error) {
	if len(addresses) == 0 {
		return "", fmt.Errorf("%w, no addresses provided", ErrValidation)
	}

	entry, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey)) //nolint:dogsled
	if err != nil {
		return "", fmt.Errorf("%w, failed to parse public key, %w", ErrValidation, err)
	}

	return knownhosts.Line(addresses, entry), nil
}

func SSHArgs(acceptNew bool) string {
	strictHostKeyChecking := "yes"

	if acceptNew {
		strictHostKeyChecking = "accept-new"
	}

	return fmt.Sprintf("-o StrictHostKeyChecking=%s -o UserKnownHostsFile={{ %s }}", strictHostKeyChecking, SSHKnownHostsFileVar)
}
