package ansible

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gliderlabs/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const (
	privateKeysDir = "private-keys"
	knownHostsDir  = "known-hosts"
	knownHostsFile = "known_hosts"
	// TODO assumes EE is unix-like with a /tmp dir.
	eePrivateKeysDir = "/tmp/private-keys"
	eeKnownHostsDir  = "/tmp/known-hosts"
)

type PrivateKey struct {
	Name string
	Data string
}

type KnownHost = string

func PrivateKeyPath(dir string, name string, eeEnabled bool) string {
	if eeEnabled {
		return strings.Join([]string{eePrivateKeysDir, name}, "/") // assume EE is unix-like
	}

	return filepath.Join(dir, privateKeysDir, name)
}

func CreatePrivateKeys(dir string, keys []PrivateKey, settings *NavigatorSettings) error {
	for _, key := range keys {
		err := writeFile(filepath.Join(dir, privateKeysDir, key.Name), key.Data)
		if err != nil {
			return fmt.Errorf("failed to create private key file for run, %w", err)
		}
	}

	if !settings.EEEnabled {
		return nil
	}

	// TODO better option?
	if settings.VolumeMounts == nil {
		settings.VolumeMounts = map[string]string{}
	}

	settings.VolumeMounts[filepath.Join(dir, privateKeysDir)] = eePrivateKeysDir

	return nil
}

func KnownHostsPath(dir string, eeEnabled bool) string {
	if eeEnabled {
		return strings.Join([]string{eeKnownHostsDir, knownHostsFile}, "/") // assume EE is unix-like
	}

	return filepath.Join(dir, knownHostsDir, knownHostsFile)
}

func CreateKnownHosts(dir string, knownHosts []KnownHost, settings *NavigatorSettings) error {
	path := filepath.Join(dir, knownHostsDir, knownHostsFile)

	err := writeFile(path, strings.Join(knownHosts, "\n"))
	if err != nil {
		return fmt.Errorf("failed to create known hosts file for run, %w", err)
	}

	if !settings.EEEnabled {
		return nil
	}

	// TODO better option?
	if settings.VolumeMounts == nil {
		settings.VolumeMounts = map[string]string{}
	}

	settings.VolumeMounts[filepath.Join(dir, knownHostsDir)] = eeKnownHostsDir

	return nil
}

func GetKnownHosts(dir string) ([]KnownHost, error) {
	path := filepath.Join(dir, knownHostsDir, knownHostsFile)

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
