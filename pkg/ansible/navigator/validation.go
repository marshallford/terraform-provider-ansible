package navigator

import (
	"fmt"
	"time"
	_ "time/tzdata" // embedded copy of the timezone database

	"github.com/containers/image/v5/docker/reference"
	jq "github.com/itchyny/gojq"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

func ValidateIANATimezone(timezone string) error {
	if len(timezone) == 0 {
		return fmt.Errorf("%w, IANA time zone must not be empty", ansible.ErrValidation)
	}

	if timezone == "local" {
		return nil
	}

	if _, err := time.LoadLocation(timezone); err != nil {
		return fmt.Errorf("%w, IANA time zone not found, %w", ansible.ErrValidation, err)
	}

	return nil
}

func ValidateJQFilter(filter string) error {
	if len(filter) == 0 {
		return fmt.Errorf("%w, JQ filter must not be empty", ansible.ErrValidation)
	}

	if _, err := jq.Parse(filter); err != nil {
		return fmt.Errorf("%w, failed to parse JQ filter, %w", ansible.ErrValidation, err)
	}

	return nil
}

func ValidateContainerImageName(image string) error {
	if len(image) == 0 {
		return fmt.Errorf("%w, container image name must not be empty", ansible.ErrValidation)
	}

	if _, err := reference.ParseNormalizedNamed(image); err != nil {
		return fmt.Errorf("%w, failed to parse container image name, %w", ansible.ErrValidation, err)
	}

	return nil
}
