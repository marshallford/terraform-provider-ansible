package navigator

import (
	"fmt"

	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
	"github.com/spf13/afero"
)

func (r *Run) Query(queries map[string]ansible.PlaybookArtifactQuery) error {
	path := r.hostJoin(playbookArtifactFilename)

	contents, err := afero.ReadFile(r.fs, path)
	if err != nil {
		return fmt.Errorf("failed to read playbook artifact, %w", err)
	}

	for name, query := range queries {
		results, err := jqQuery(contents, query.JQFilter, query.Raw)
		if err != nil {
			return fmt.Errorf("failed to query playbook artifact, %w", err)
		}

		query.Results = results
		queries[name] = query
	}

	return nil
}

func (r *Run) ReadKnownHosts() ([]ansible.KnownHost, error) {
	file, err := r.fs.Open(r.hostJoin(knownHostsDir, knownHostsFile))
	if err != nil {
		return nil, fmt.Errorf("failed to open known hosts file, %w", err)
	}

	defer file.Close() //nolint:errcheck

	return ansible.ParseKnownHosts(file)
}
