package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const (
	TestingVersion = "test"
)

// Ensure K0sctlProvider satisfies various provider interfaces.
var _ provider.Provider = &K0sctlProvider{}

// K0sctlProvider defines the provider implementation.
type K0sctlProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// K0sctlProviderModel describes the provider data model.
type K0sctlProviderModel struct {
	testingMode bool
}

func (p *K0sctlProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "k0sctl"
	resp.Version = p.version
}

func (p *K0sctlProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{},
	}
}

func (p *K0sctlProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data K0sctlProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if p.version == TestingVersion {
		data.testingMode = true
	}

	resp.ResourceData = &data
	resp.DataSourceData = &data

	AllLoggingToTFLog(ctx)
}

func (p *K0sctlProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewK0sctlConfigResource,
	}
}

func (p *K0sctlProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &K0sctlProvider{
			version: version,
		}
	}
}
