package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
)

func privateKeyTestCases() []TestCase {
	return []TestCase{
		{
			name: "ee_enabled",
			variables: func(t *testing.T) config.Variables { //nolint:thelper
				return config.Variables{
					"ee_enabled": config.BoolVariable(true),
				}
			},
			setup: func(t *testing.T) { //nolint:thelper
				t.Parallel()
			},
		},
		{
			name: "ee_disabled",
			variables: func(t *testing.T) config.Variables { //nolint:thelper
				return config.Variables{
					"ee_enabled": config.BoolVariable(false),
				}
			},
			setup: func(t *testing.T) { //nolint:thelper
				testPrependPlaybookToPath(t)
			},
		},
	}
}