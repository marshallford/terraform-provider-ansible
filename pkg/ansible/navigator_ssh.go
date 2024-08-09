package ansible

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	privateKeysDir = "private-keys"
	knownHostsDir  = "known-hosts"
	knownHostsFile = "known_hosts"
	// TODO assumes EE is unix-like with a /tmp dir
	eePrivateKeysDir = "/tmp/private-keys"
	eeKnownHostsDir  = "/tmp/known-hosts"
)

type PrivateKey struct {
	Name string
	Data string
}

type KnownHost = string

func privateKeyPath(dir string, key string, eeEnabled bool) string {
	if eeEnabled {
		return strings.Join([]string{eePrivateKeysDir, key}, "/") // assume EE is unix-like
	}

	return filepath.Join(dir, privateKeysDir, key)
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

func knownHostsPath(dir string, eeEnabled bool) string {
	if eeEnabled {
		return strings.Join([]string{eeKnownHostsDir, knownHostsFile}, "/") // assume EE is unix-like
	}

	return filepath.Join(dir, knownHostsDir, knownHostsFile)
}

func CreateKnownHosts(dir string, knownHosts []KnownHost, settings *NavigatorSettings) error {
	err := writeFile(filepath.Join(dir, knownHostsDir, knownHostsFile), strings.Join(knownHosts, "\n"))
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
