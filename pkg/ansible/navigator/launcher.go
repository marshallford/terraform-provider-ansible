package navigator

import (
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

type Launcher interface {
	PrepareCommand(cmd ansible.Cmd, config *RunConfig) ansible.Cmd
	Cleanup() error
}

type nativeLauncher struct{}

var _ Launcher = (*nativeLauncher)(nil)

func NativeLauncher() Launcher { //nolint:ireturn
	return nativeLauncher{}
}

func (nativeLauncher) PrepareCommand(cmd ansible.Cmd, _ *RunConfig) ansible.Cmd { //nolint:ireturn
	return cmd
}

func (nativeLauncher) Cleanup() error {
	return nil
}
