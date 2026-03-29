package ansible

import (
	"slices"
	"testing"
)

func TestPlaybookOptionsArgs(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    PlaybookOptions
		expected []string
	}{
		"empty": {
			input:    PlaybookOptions{},
			expected: nil,
		},
		"simple": {
			input: PlaybookOptions{
				SkipTags: []string{"tag1"},
			},
			expected: []string{"--skip-tags", "tag1"},
		},
		"all": {
			input: PlaybookOptions{
				ForceHandlers: true,
				SkipTags:      []string{"tag1", "tag2"},
				StartAtTask:   "task name",
				Limit:         []string{"host1", "host2"},
				Tags:          []string{"tag3", "tag4"},
			},
			expected: []string{
				"--force-handlers",
				"--skip-tags", "tag1,tag2",
				"--start-at-task", "task name",
				"--limit", "host1,host2",
				"--tags", "tag3,tag4",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.input.Args()

			if !slices.Equal(got, test.expected) {
				t.Errorf("expected %v, got %v", test.expected, got)
			}
		})
	}
}
