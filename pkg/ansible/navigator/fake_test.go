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
	environ   []string
	responses []fakeResponse

	commands []ansible.Command
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

func (e *fakeExecutor) Environ() []string {
	return append([]string(nil), e.environ...)
}

func (e *fakeExecutor) Abs(path string) (string, error) {
	if strings.HasPrefix(path, "/") {
		return path, nil
	}

	return "/abs/" + strings.TrimPrefix(path, "./"), nil
}

func (e *fakeExecutor) Run(_ context.Context, command ansible.Command) ([]byte, error) {
	e.commands = append(e.commands, command)

	for _, response := range e.responses {
		if strings.Contains(command.String(), response.match) {
			return []byte(response.output), response.err
		}
	}

	return nil, nil
}

func (e *fakeExecutor) commandStrings() []string {
	strs := make([]string, 0, len(e.commands))
	for _, command := range e.commands {
		strs = append(strs, command.String())
	}

	return strs
}

func (e *fakeExecutor) envDelta(command ansible.Command) []string {
	host := map[string]struct{}{}
	for _, entry := range e.environ {
		host[entry] = struct{}{}
	}

	var delta []string

	for _, entry := range command.Env {
		if _, ok := host[entry]; !ok {
			delta = append(delta, entry)
		}
	}

	return delta
}
