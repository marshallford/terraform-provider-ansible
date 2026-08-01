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

	if r.config.mode().UsesEE() {
		if err := r.checkContainerEngine(ctx); err != nil {
			errs = append(errs, err)
		}
	} else {
		if err := r.checkPlaybookBinary(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if err := r.checkNavigatorBinary(ctx); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (r *Run) checkWorkingDir() error {
	if err := ansible.CheckDirectory(r.fs, r.config.WorkingDir); err != nil {
		return &PreflightError{Check: CheckWorkingDir, Message: "working directory is not valid", Err: err}
	}

	workingDir, err := r.exec.Abs(r.config.WorkingDir)
	if err != nil {
		return &PreflightError{Check: CheckWorkingDir, Message: "absolute path of working directory cannot be determined", Err: err}
	}

	r.resolved.workingDir = workingDir

	return nil
}

func (r *Run) checkContainerEngine(ctx context.Context) error {
	engine := r.config.Settings.ExecutionEnvironment.ContainerEngine

	if !slices.Contains(ContainerEngineOptions(), engine) {
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
		for _, option := range ContainerEngines() {
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

	if _, err := r.exec.Run(ctx, ansible.Command{Name: engine, Args: []string{"info"}}); err != nil {
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

	stdoutStderr, err := r.exec.Run(ctx, ansible.Command{Name: ansible.PlaybookProgram, Args: []string{"--version"}})
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

func (r *Run) resolveNavigatorBinary() error {
	if r.config.Binary == "" {
		path, err := r.exec.LookPath(Program)
		if err != nil {
			return &PreflightError{
				Check:   CheckNavigatorResolve,
				Message: fmt.Sprintf("%s not found in PATH", Program),
			}
		}

		r.resolved.navigatorBinary = path

		return nil
	}

	path, err := r.exec.Abs(r.config.Binary)
	if err != nil {
		return &PreflightError{
			Check:   CheckNavigatorResolve,
			Message: fmt.Sprintf("absolute path of %s cannot be determined", Program),
			Err:     err,
		}
	}

	r.resolved.navigatorBinary = path

	return nil
}

func (r *Run) checkNavigatorBinary(ctx context.Context) error {
	if err := r.resolveNavigatorBinary(); err != nil {
		return err
	}

	stdoutStderr, err := r.exec.Run(ctx, ansible.Command{Name: r.resolved.navigatorBinary, Args: []string{"--version"}})
	if err != nil {
		return &PreflightError{
			Check:   CheckNavigatorBinary,
			Message: fmt.Sprintf("'%s --version' command failed", r.resolved.navigatorBinary),
			Err:     err,
		}
	}

	if !strings.HasPrefix(string(stdoutStderr), Program) {
		return &PreflightError{
			Check:   CheckNavigatorBinary,
			Message: fmt.Sprintf("'%s --version' command output not expected", r.resolved.navigatorBinary),
		}
	}

	return nil
}

func (r *Run) programExistsOnPath(program string) error {
	_, err := r.exec.LookPath(program)

	return err
}
