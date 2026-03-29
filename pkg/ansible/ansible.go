package ansible

import "strings"

const (
	PlaybookProgram              = "ansible-playbook"
	RunnerDefaultHostKeyChecking = false
	SSHKnownHostsFileVar         = "ansible_ssh_known_hosts_file"
)

type PlaybookOptions struct {
	ForceHandlers bool
	SkipTags      []string
	StartAtTask   string
	Limit         []string
	Tags          []string
}

func (o *PlaybookOptions) Args() []string {
	var args []string

	if o.ForceHandlers {
		args = append(args, "--force-handlers")
	}

	if len(o.SkipTags) > 0 {
		args = append(args, "--skip-tags", strings.Join(o.SkipTags, ","))
	}

	if o.StartAtTask != "" {
		args = append(args, "--start-at-task", o.StartAtTask)
	}

	if len(o.Limit) > 0 {
		args = append(args, "--limit", strings.Join(o.Limit, ","))
	}

	if len(o.Tags) > 0 {
		args = append(args, "--tags", strings.Join(o.Tags, ","))
	}

	return args
}

type Inventory struct {
	Name     string
	Contents string
	Exclude  bool
}

type ExtraVarsFile struct {
	Name     string
	Contents string
}
