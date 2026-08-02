package navigator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/afero"
)

func (r *Run) Setup() error {
	if err := r.createDirs(); err != nil {
		return err
	}

	writes := []struct {
		needed bool
		write  func() error
	}{
		{true, r.writePlaybook},
		{true, r.writeInventories},
		{len(r.config.ExtraVars) > 0, r.writeExtraVars},
		{len(r.config.PrivateKeys) > 0, r.writePrivateKeys},
		{r.config.UseKnownHosts, r.writeKnownHosts},
		{true, r.writeSettings},
	}

	var errs []error

	for _, w := range writes {
		if !w.needed {
			continue
		}

		if err := w.write(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (r *Run) createDirs() error {
	if err := r.fs.Mkdir(r.HostDir(), dirPermissions); err != nil {
		return newSetupError(SetupDir, "failed to create directory for run", err)
	}

	if err := r.fs.Mkdir(r.hostJoin(inventoriesDir), dirPermissions); err != nil {
		return newSetupError(SetupDir, "failed to create inventories directory for run", err)
	}

	if err := r.fs.Mkdir(r.hostJoin(extraVarsDir), dirPermissions); err != nil {
		return newSetupError(SetupDir, "failed to create extra vars directory for run", err)
	}

	if err := r.fs.Mkdir(r.hostJoin(privateKeysDir), dirPermissions); err != nil {
		return newSetupError(SetupDir, "failed to create private keys directory for run", err)
	}

	if err := r.fs.Mkdir(r.hostJoin(knownHostsDir), dirPermissions); err != nil {
		return newSetupError(SetupDir, "failed to create known hosts directory for run", err)
	}

	return nil
}

func (r *Run) writeSettings() error {
	contents, err := r.settings().generate()
	if err != nil {
		return newSetupError(SetupSettings, "failed to generate navigator settings for run", err)
	}

	if err := r.writeFile(r.hostJoin(navigatorSettingsFilename), contents); err != nil {
		return newSetupError(SetupSettings, "failed to create navigator settings file for run", err)
	}

	return nil
}

func (r *Run) writePlaybook() error {
	if err := r.writeFile(r.hostJoin(playbookFilename), r.config.Playbook); err != nil {
		return newSetupError(SetupPlaybook, "failed to create playbook file for run", err)
	}

	return nil
}

func (r *Run) writeInventories() error {
	for _, inventory := range r.config.Inventories {
		err := r.writeFile(r.hostJoin(inventoriesDir, inventory.Name), inventory.Contents)
		if err != nil {
			return newSetupError(SetupInventories, "failed to create ansible inventory file for run", err)
		}
	}

	return nil
}

func (r *Run) writeExtraVars() error {
	for _, f := range r.config.ExtraVars {
		err := r.writeFile(r.hostJoin(extraVarsDir, f.Name), f.Contents)
		if err != nil {
			return newSetupError(SetupExtraVars, "failed to create extra vars file for run", err)
		}
	}

	return nil
}

func (r *Run) writePrivateKeys() error {
	for _, key := range r.config.PrivateKeys {
		err := r.writeFile(r.hostJoin(privateKeysDir, key.Name), key.Data)
		if err != nil {
			return newSetupError(SetupPrivateKeys, "failed to create private key file for run", err)
		}
	}

	return nil
}

func (r *Run) writeKnownHosts() error {
	path := r.hostJoin(knownHostsDir, knownHostsFile)
	err := r.writeFile(path, strings.Join(r.config.KnownHosts, "\n"))
	if err != nil {
		return newSetupError(SetupKnownHosts, "failed to create known hosts file for run", err)
	}

	return nil
}

func (r *Run) writeFile(path string, contents string) error {
	if err := afero.WriteFile(r.fs, path, []byte(contents), filePermissions); err != nil {
		return fmt.Errorf("failed to write file, %w", err)
	}

	return nil
}
