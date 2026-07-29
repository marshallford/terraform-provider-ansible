package ansible

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const commandWaitDelay = 10 * time.Second

// Command describes a process to run. Zero values follow os/exec: an empty Dir
// means the current process's working directory, and a nil Env means the child
// inherits the current process's environment. Note that a non-nil but empty Env
// means the opposite, an environment with nothing in it.
type Command struct {
	Name string
	Args []string
	Dir  string
	Env  []string
}

func (c Command) String() string {
	return strings.Join(append([]string{c.Name}, c.Args...), " ")
}

func (c Command) AppendArgs(args ...string) Command {
	c.Args = append(slices.Clone(c.Args), args...)

	return c
}

func (c Command) AppendEnv(key, value string) Command {
	c.Env = append(slices.Clone(c.Env), key+"="+value)

	return c
}

type Executor interface {
	LookPath(file string) (string, error)
	Abs(path string) (string, error)
	Environ() []string
	Run(ctx context.Context, command Command) ([]byte, error)
}

type osExecutor struct{}

var _ Executor = (*osExecutor)(nil)

func OSExecutor() Executor { //nolint:ireturn
	return osExecutor{}
}

func (osExecutor) LookPath(file string) (string, error) {
	return exec.LookPath(file) //nolint:wrapcheck
}

func (osExecutor) Environ() []string {
	return os.Environ()
}

func (osExecutor) Abs(path string) (string, error) {
	return filepath.Abs(path) //nolint:wrapcheck
}

func (osExecutor) Run(ctx context.Context, command Command) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command.Name, command.Args...) //nolint:gosec
	cmd.WaitDelay = commandWaitDelay
	cmd.Dir = command.Dir
	cmd.Env = command.Env

	return cmd.CombinedOutput() //nolint:wrapcheck
}
