package navigator

import (
	"errors"
	"strings"
)

func (r *Run) Setup() error {
	if err := r.createDirs(); err != nil {
		return err
	}

	var errs []error

	if err := r.writePlaybook(); err != nil {
		errs = append(errs, err)
	}

	if err := r.writeInventories(); err != nil {
		errs = append(errs, err)
	}

	if len(r.config.ExtraVars) > 0 {
		if err := r.writeExtraVars(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(r.config.PrivateKeys) > 0 {
		if err := r.writePrivateKeys(); err != nil {
			errs = append(errs, err)
		}
	}

	if r.config.UseKnownHosts {
		if err := r.writeKnownHosts(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (r *Run) createDirs() error {
	if err := r.fs.Mkdir(r.hostDir, dirPermissions); err != nil {
		return &SetupError{Component: SetupDir, Message: "failed to create directory for run", Err: err}
	}

	if err := r.fs.Mkdir(r.hostJoin(inventoriesDir), dirPermissions); err != nil {
		return &SetupError{Component: SetupDir, Message: "failed to create inventories directory for run", Err: err}
	}

	if err := r.fs.Mkdir(r.hostJoin(extraVarsDir), dirPermissions); err != nil {
		return &SetupError{Component: SetupDir, Message: "failed to create extra vars directory for run", Err: err}
	}

	if err := r.fs.Mkdir(r.hostJoin(privateKeysDir), dirPermissions); err != nil {
		return &SetupError{Component: SetupDir, Message: "failed to create private keys directory for run", Err: err}
	}

	if err := r.fs.Mkdir(r.hostJoin(knownHostsDir), dirPermissions); err != nil {
		return &SetupError{Component: SetupDir, Message: "failed to create known hosts directory for run", Err: err}
	}

	if !r.config.Settings.EEEnabled {
		return nil
	}

	r.resolvedDir = containerRunDir
	r.resolvedPathSeparator = containerPathSeparator

	if r.config.Settings.VolumeMounts == nil {
		r.config.Settings.VolumeMounts = map[string]string{}
	}

	r.config.Settings.VolumeMounts[r.hostDir] = r.resolvedDir

	return nil
}

func (r *Run) writePlaybook() error {
	if err := writeFile(r.fs, r.hostJoin(playbookFilename), r.config.Playbook); err != nil {
		return &SetupError{Component: SetupPlaybook, Message: "failed to create playbook file for run", Err: err}
	}

	return nil
}

func (r *Run) writeInventories() error {
	for _, inventory := range r.config.Inventories {
		err := writeFile(r.fs, r.hostJoin(inventoriesDir, inventory.Name), inventory.Contents)
		if err != nil {
			return &SetupError{Component: SetupInventories, Message: "failed to create ansible inventory file for run", Err: err}
		}
	}

	return nil
}

func (r *Run) writeExtraVars() error {
	for _, f := range r.config.ExtraVars {
		err := writeFile(r.fs, r.hostJoin(extraVarsDir, f.Name), f.Contents)
		if err != nil {
			return &SetupError{Component: SetupExtraVars, Message: "failed to create extra vars file for run", Err: err}
		}
	}

	return nil
}

func (r *Run) writePrivateKeys() error {
	for _, key := range r.config.PrivateKeys {
		err := writeFile(r.fs, r.hostJoin(privateKeysDir, key.Name), key.Data)
		if err != nil {
			return &SetupError{Component: SetupPrivateKeys, Message: "failed to create private key file for run", Err: err}
		}
	}

	return nil
}

func (r *Run) writeKnownHosts() error {
	path := r.hostJoin(knownHostsDir, knownHostsFile)
	err := writeFile(r.fs, path, strings.Join(r.config.KnownHosts, "\n"))
	if err != nil {
		return &SetupError{Component: SetupKnownHosts, Message: "failed to create known hosts file for run", Err: err}
	}

	return nil
}
