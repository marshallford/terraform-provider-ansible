package provider_test

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/marshallford/terraform-provider-ansible/internal/provider"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){ //nolint:gochecknoglobals
	"ansible": providerserver.NewProtocol6WithError(provider.New("test")()),
}
