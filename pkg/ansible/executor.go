package ansible

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const commandWaitDelay = 10 * time.Second

type Cmd interface {
	Run() ([]byte, error)
	SetDir(dir string)
	SetEnv(env []string)
	AppendEnv(key, value string)
	AppendArgs(args ...string)
	String() string
}

type Executor interface {
	LookPath(file string) (string, error)
	CommandContext(ctx context.Context, name string, args ...string) Cmd
	Environ() []string
	Abs(path string) (string, error)
}

type osExecutor struct{}

var (
	_ Executor = (*osExecutor)(nil)
	_ Cmd      = (*osCmd)(nil)
)

func OSExecutor() Executor { //nolint:ireturn
	return osExecutor{}
}

func (osExecutor) LookPath(file string) (string, error) {
	return exec.LookPath(file) //nolint:wrapcheck
}

func (osExecutor) CommandContext(ctx context.Context, name string, args ...string) Cmd { //nolint:ireturn
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec
	cmd.WaitDelay = commandWaitDelay

	return &osCmd{cmd: cmd}
}

func (osExecutor) Environ() []string {
	return os.Environ()
}

func (osExecutor) Abs(path string) (string, error) {
	return filepath.Abs(path) //nolint:wrapcheck
}

type osCmd struct {
	cmd *exec.Cmd
}

func (c *osCmd) Run() ([]byte, error) {
	return c.cmd.CombinedOutput() //nolint:wrapcheck
}

func (c *osCmd) SetDir(dir string) {
	c.cmd.Dir = dir
}

func (c *osCmd) SetEnv(env []string) {
	c.cmd.Env = env
}

func (c *osCmd) AppendEnv(key, value string) {
	c.cmd.Env = append(c.cmd.Env, key+"="+value)
}

func (c *osCmd) AppendArgs(args ...string) {
	c.cmd.Args = append(c.cmd.Args, args...)
}

func (c *osCmd) String() string {
	return c.cmd.String()
}
