package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

var (
	_ function.Function = (*SSHArgsFunction)(nil)
)

func NewSSHArgsFunction() function.Function { //nolint:ireturn
	return &SSHArgsFunction{}
}

type SSHArgsFunction struct{}

func (f *SSHArgsFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "ssh_args"
}

func (f *SSHArgsFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             "SSH args for configuring Ansible to integrate with provider managed known hosts.",
		Description:         "SSH command line arguments for configuring Ansible to integrate with provider managed known hosts. Set or append to the 'ansible_ssh_common_args' Ansible variable or environment variable.",
		MarkdownDescription: "SSH command line arguments for configuring Ansible to integrate with provider managed known hosts. Set or append to the `ansible_ssh_common_args` Ansible variable or environment variable.",

		Parameters: []function.Parameter{
			function.BoolParameter{
				Name:                "accept_new",
				Description:         "Accept and add new host keys ('StrictHostKeyChecking=accept_new') or only allow connections to hosts whose key(s) are already present ('StrictHostKeyChecking=yes').",
				MarkdownDescription: "Accept and add new host keys (`StrictHostKeyChecking=accept_new`) or only allow connections to hosts whose key(s) are already present (`StrictHostKeyChecking=yes`).",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *SSHArgsFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var acceptNew bool

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &acceptNew))

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, ansible.SSHArgs(acceptNew)))
}
