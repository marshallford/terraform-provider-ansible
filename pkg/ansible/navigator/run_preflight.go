package navigator

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

func (r *Run) Preflight(ctx context.Context) error {
	var errs []error

	if err := r.checkWorkingDir(); err != nil {
		errs = append(errs, err)
	}

	if r.config.Settings.EEEnabled {
		if err := r.checkContainerEngine(ctx); err != nil {
			errs = append(errs, err)
		}
	} else {
		if err := r.checkPlaybookBinary(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	binary, err := r.resolveNavigatorBinary()
	if err != nil {
		errs = append(errs, err)
	} else {
		r.binary = binary
		if err := r.checkNavigatorBinary(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (r *Run) checkWorkingDir() error {
	if err := ansible.CheckDirectory(r.fs, r.config.WorkingDir); err != nil {
		return &PreflightError{Check: CheckWorkingDir, Message: "working directory is not valid", Err: err}
	}

	return nil
}

func (r *Run) checkContainerEngine(ctx context.Context) error {
	engine := r.config.Settings.ContainerEngine

	if !slices.Contains(ContainerEngineOptions(true), engine) {
		return &PreflightError{
			Check:   CheckContainerEngine,
			Message: fmt.Sprintf("container engine %s is not a valid option", engine),
		}
	}

	if engine != ContainerEngineAuto && r.programExistsOnPath(engine) != nil {
		return &PreflightError{
			Check:   CheckContainerEngine,
			Message: fmt.Sprintf("container engine %s not found in PATH", engine),
		}
	}

	if engine == ContainerEngineAuto {
		for _, option := range ContainerEngineOptions(false) {
			if r.programExistsOnPath(option) == nil {
				engine = option

				break
			}
		}
	}

	if engine == ContainerEngineAuto {
		return &PreflightError{
			Check:   CheckContainerEngine,
			Message: "no container engine found in PATH",
		}
	}

	command := r.exec.CommandContext(ctx, engine, "info")
	if _, err := command.Run(); err != nil {
		return &PreflightError{
			Check:   CheckContainerEngine,
			Message: fmt.Sprintf("container engine is not running or usable, '%s info' command failed", engine),
			Err:     err,
		}
	}

	return nil
}

func (r *Run) checkPlaybookBinary(ctx context.Context) error {
	if err := r.programExistsOnPath(ansible.PlaybookProgram); err != nil {
		return &PreflightError{
			Check:   CheckPlaybook,
			Message: fmt.Sprintf("%s not found in PATH, required when not using an execution environment", ansible.PlaybookProgram),
		}
	}

	command := r.exec.CommandContext(ctx, ansible.PlaybookProgram, "--version")
	stdoutStderr, err := command.Run()
	if err != nil {
		return &PreflightError{
			Check:   CheckPlaybook,
			Message: fmt.Sprintf("'%s --version' command failed", ansible.PlaybookProgram),
			Err:     err,
		}
	}

	if !strings.HasPrefix(string(stdoutStderr), ansible.PlaybookProgram) {
		return &PreflightError{
			Check:   CheckPlaybook,
			Message: fmt.Sprintf("'%s --version' command output not expected", ansible.PlaybookProgram),
		}
	}

	return nil
}

func (r *Run) resolveNavigatorBinary() (string, error) {
	path := r.config.Binary
	if path == "" {
		path, err := r.exec.LookPath(Program)
		if err != nil {
			return "", &PreflightError{
				Check:   CheckNavigatorResolve,
				Message: fmt.Sprintf("%s not found in PATH", Program),
			}
		}

		return path, nil
	}

	path, err := r.exec.Abs(path)
	if err != nil {
		return "", &PreflightError{
			Check:   CheckNavigatorResolve,
			Message: fmt.Sprintf("absolute path of %s cannot be determined", Program),
			Err:     err,
		}
	}

	return path, nil
}

func (r *Run) checkNavigatorBinary(ctx context.Context) error {
	command := r.exec.CommandContext(ctx, r.binary, "--version")
	stdoutStderr, err := command.Run()
	if err != nil {
		return &PreflightError{
			Check:   CheckNavigatorBinary,
			Message: fmt.Sprintf("'%s --version' command failed", r.binary),
			Err:     err,
		}
	}

	if !strings.HasPrefix(string(stdoutStderr), Program) {
		return &PreflightError{
			Check:   CheckNavigatorBinary,
			Message: fmt.Sprintf("'%s --version' command output not expected", r.binary),
		}
	}

	return nil
}

func (r *Run) programExistsOnPath(program string) error {
	_, err := r.exec.LookPath(program)

	return err
}
