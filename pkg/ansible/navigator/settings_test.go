package navigator

import (
	"slices"
	"strings"
	"testing"
)

func TestVolumeMountOptionsMatchNavigator(t *testing.T) {
	t.Parallel()

	want := VolumeMountOptions{"O", "ro", "rw", "z", "Z"}

	got := slices.Clone(allVolumeMountOptions())
	slices.Sort(got)
	slices.Sort(want)

	if !slices.Equal(got, want) {
		t.Errorf("VolumeMountOptions drifted from navigator\nwant: %v\ngot:  %v", want, got)
	}
}

func TestVolumeMountOptionsString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		options VolumeMountOptions
		want    string
	}{
		"none":     {want: ""},
		"single":   {options: VolumeMountOptions{VolumeMountRelabelUnshared}, want: "Z"},
		"multiple": {options: VolumeMountOptions{VolumeMountReadOnly, VolumeMountRelabelShared}, want: "ro,z"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := test.options.String(); got != test.want {
				t.Errorf("want %q, got %q", test.want, got)
			}
		})
	}
}

// An empty options string is what navigator treats as "no options"; a stray
// separator would be parsed as an unrecognised option.
func TestVolumeMountOptionsStringHasNoStraySeparator(t *testing.T) {
	t.Parallel()

	for _, option := range allVolumeMountOptions() {
		got := VolumeMountOptions{option}.String()
		if strings.Contains(got, ",") {
			t.Errorf("option %q rendered with a separator: %q", option, got)
		}
	}
}
