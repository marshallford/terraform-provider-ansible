package navigator

import (
	"errors"
	"fmt"

	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

func (r *Run) Query(queries map[string]ansible.PlaybookArtifactQuery) error {
	contents, err := r.readPlaybookArtifact()
	if err != nil {
		return err
	}

	var errs []error

	for name, query := range queries {
		results, err := ansible.QueryPlaybookArtifact(contents, query)
		if err != nil {
			errs = append(errs, newQueryError(name, "failed to query playbook artifact", err))

			continue
		}

		query.Results = results
		queries[name] = query
	}

	return errors.Join(errs...)
}

func (r *Run) ReadKnownHosts() ([]ansible.KnownHost, error) {
	file, err := r.fs.Open(r.hostJoin(knownHostsDir, knownHostsFile))
	if err != nil {
		return nil, fmt.Errorf("failed to open known hosts file, %w", err)
	}

	defer file.Close() //nolint:errcheck

	return ansible.ParseKnownHosts(file)
}
