package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

var (
	_ function.Function = (*SSHKnownHostFunction)(nil)
)

func NewSSHKnownHostFunction() function.Function { //nolint:ireturn
	return &SSHKnownHostFunction{}
}

type SSHKnownHostFunction struct{}

func (f *SSHKnownHostFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "ssh_known_host"
}

func (f *SSHKnownHostFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Format a public key and addresses into a known hosts entry.",
		Description: "Format a public key and addresses into a known hosts entry/line suitable for use in an SSH known hosts file.",

		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "public_key",
				Description: "Public key data in the authorized keys format.",
			},
		},
		VariadicParameter: function.StringParameter{
			Name:        "addresses",
			Description: "Addresses to associate with the public key. Can be one or more hostnames or IP addresses with an optional port.",
		},
		Return: function.StringReturn{},
	}
}

func (f *SSHKnownHostFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var publicKey string
	var addresses []string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &publicKey, &addresses))

	entry, err := ansible.KnownHostsLine(addresses, publicKey)
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(err.Error()))

		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, entry))
}
