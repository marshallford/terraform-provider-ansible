package navigator

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

type fakeExecutor struct {
	lookPath  map[string]string
	absErr    error
	abs       func(string) (string, error)
	environ   []string
	responses []fakeResponse

	commands []*fakeCmd
}

type fakeResponse struct {
	match  string
	output string
	err    error
}

func newFakeExecutor() *fakeExecutor {
	return &fakeExecutor{
		lookPath: map[string]string{},
		environ:  []string{"PATH=/usr/bin", "HOME=/home/test"},
	}
}

func (e *fakeExecutor) withProgram(programs ...string) *fakeExecutor {
	for _, program := range programs {
		e.lookPath[program] = "/usr/bin/" + program
	}

	return e
}

func (e *fakeExecutor) withResponse(match string, output string, err error) *fakeExecutor {
	e.responses = append(e.responses, fakeResponse{match: match, output: output, err: err})

	return e
}

func (e *fakeExecutor) LookPath(file string) (string, error) {
	if path, ok := e.lookPath[file]; ok {
		return path, nil
	}

	return "", fmt.Errorf("%s: %w", file, exec.ErrNotFound)
}

func (e *fakeExecutor) CommandContext(_ context.Context, name string, args ...string) ansible.Cmd { //nolint:ireturn
	cmd := &fakeCmd{exec: e, argv: append([]string{name}, args...)}
	e.commands = append(e.commands, cmd)

	return cmd
}

func (e *fakeExecutor) Environ() []string {
	return append([]string(nil), e.environ...)
}

func (e *fakeExecutor) Abs(path string) (string, error) {
	if e.absErr != nil {
		return "", e.absErr
	}

	if e.abs != nil {
		return e.abs(path)
	}

	if strings.HasPrefix(path, "/") {
		return path, nil
	}

	return "/abs/" + strings.TrimPrefix(path, "./"), nil
}

func (e *fakeExecutor) commandStrings() []string {
	strs := make([]string, 0, len(e.commands))
	for _, cmd := range e.commands {
		strs = append(strs, cmd.String())
	}

	return strs
}

func (e *fakeExecutor) lastCommand() *fakeCmd {
	if len(e.commands) == 0 {
		return nil
	}

	return e.commands[len(e.commands)-1]
}

// argv[0] is the program, matching osCmd: exec.Cmd.String() renders Path
// followed by Args[1:], and the binary is always resolved to an absolute path
// before CommandContext, so Path equals Args[0].
type fakeCmd struct {
	exec *fakeExecutor
	argv []string
	dir  string
	env  []string
	runs int
}

func (c *fakeCmd) Run() ([]byte, error) {
	c.runs++

	command := c.String()
	for _, response := range c.exec.responses {
		if strings.Contains(command, response.match) {
			return []byte(response.output), response.err
		}
	}

	return nil, nil
}

func (c *fakeCmd) SetDir(dir string) {
	c.dir = dir
}

func (c *fakeCmd) SetEnv(env []string) {
	c.env = append([]string(nil), env...)
}

func (c *fakeCmd) AppendEnv(key, value string) {
	c.env = append(c.env, key+"="+value)
}

func (c *fakeCmd) AppendArgs(args ...string) {
	c.argv = append(c.argv, args...)
}

func (c *fakeCmd) String() string {
	return strings.Join(c.argv, " ")
}

func (c *fakeCmd) envDelta() []string {
	host := map[string]struct{}{}
	for _, entry := range c.exec.environ {
		host[entry] = struct{}{}
	}

	var delta []string

	for _, entry := range c.env {
		if _, ok := host[entry]; !ok {
			delta = append(delta, entry)
		}
	}

	return delta
}
