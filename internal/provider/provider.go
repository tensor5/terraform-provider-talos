package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ provider.Provider = &TalosProvider{}
var _ provider.ProviderWithMetadata = &TalosProvider{}

type TalosProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type TalosProviderModel struct{}

func (p *TalosProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "talos"
	resp.Version = p.version
}

func (p *TalosProvider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{},
	}, nil
}

func (p *TalosProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data TalosProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (p *TalosProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewBootstrapResource,
		NewGenConfigResource,
	}
}

func (p *TalosProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewKubeconfigDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TalosProvider{
			version: version,
		}
	}
}
