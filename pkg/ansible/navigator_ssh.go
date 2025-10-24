package ansible

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const (
	privateKeysDir = "private-keys"
	knownHostsDir  = "known-hosts"
	knownHostsFile = "known_hosts"
)

type PrivateKey struct {
	Name string
	Data string
}

type KnownHost = string

func CreatePrivateKeys(runDir *RunDir, keys []PrivateKey) error {
	for _, key := range keys {
		err := writeFile(runDir.HostJoin(privateKeysDir, key.Name), key.Data)
		if err != nil {
			return fmt.Errorf("failed to create private key file for run, %w", err)
		}
	}

	return nil
}

func CreateKnownHosts(runDir *RunDir, knownHosts []KnownHost) error {
	path := runDir.HostJoin(knownHostsDir, knownHostsFile)

	err := writeFile(path, strings.Join(knownHosts, "\n"))
	if err != nil {
		return fmt.Errorf("failed to create known hosts file for run, %w", err)
	}

	return nil
}

func GetKnownHosts(runDir *RunDir) ([]KnownHost, error) {
	path := runDir.HostJoin(knownHostsDir, knownHostsFile)

	file, err := os.Open(path) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("failed to read %s, %w", knownHostsFile, err)
	}

	defer file.Close() //nolint:errcheck

	knownHosts := make([]KnownHost, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		knownHosts = append(knownHosts, scanner.Text())
	}

	if scanner.Err() != nil {
		return nil, fmt.Errorf("failed to read %s, %w", knownHostsFile, scanner.Err())
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
