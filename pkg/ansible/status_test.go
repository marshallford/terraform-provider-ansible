package ansible_test

import (
	"testing"

	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

func TestParseStatus(t *testing.T) {
	t.Parallel()

	for _, status := range []string{"successful", "failed", "timeout", "canceled"} {
		if got := ansible.ParseStatus(status); got != ansible.Status(status) {
			t.Errorf("ParseStatus(%q): want %q, got %q", status, status, got)
		}
	}

	for _, status := range []string{"unstarted", "starting", "running", "error", "", "cancelled", "Successful"} {
		if got := ansible.ParseStatus(status); got != ansible.StatusUnknown {
			t.Errorf("ParseStatus(%q): want %q, got %q", status, ansible.StatusUnknown, got)
		}
	}
}
