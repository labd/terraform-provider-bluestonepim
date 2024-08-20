package provider

import (
	"context"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-bluestonepim/internal/resources/attribute_definition"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/category"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/category_attribute"
	"github.com/labd/terraform-provider-bluestonepim/internal/sdk/pim"
	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

// Ensure BluestonePimProvider satisfies various provider interfaces.
var _ provider.Provider = &BluestonePimProvider{}
var _ provider.ProviderWithFunctions = &BluestonePimProvider{}

// BluestonePimProvider defines the provider implementation.
type BluestonePimProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// BluestonePimProviderModel describes the provider data model.
type BluestonePimProviderModel struct {
	ClientSecret types.String `tfsdk:"client_secret"`
}

func (p *BluestonePimProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "bluestonepim"
	resp.Version = p.version
}

func (p *BluestonePimProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_secret": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The client secret for Bluestone PIM Management API (MAPI)",
				Sensitive:           true,
			},
		},
	}
}

func (p *BluestonePimProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data BluestonePimProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	httpClient := &http.Client{
		Transport: utils.DebugTransport,
	}

	apiKeyProvider, apiKeyProviderErr := securityprovider.NewSecurityProviderApiKey("header", "api-key", data.ClientSecret.ValueString())
	if apiKeyProviderErr != nil {
		panic(apiKeyProviderErr)
	}
	client, err := pim.NewClientWithResponses(
		"https://mapi.test.bluestonepim.com/pim",
		pim.WithRequestEditorFn(apiKeyProvider.Intercept),
		pim.WithHTTPClient(httpClient))
	if err != nil {
		panic(err)
	}

	container := &utils.ProviderData{
		Client: client,
	}

	// Example client configuration for data sources and resources
	resp.DataSourceData = container
	resp.ResourceData = container
}

func (p *BluestonePimProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		category.NewResource,
		attribute_definition.NewResource,
		category_attribute.NewResource,
	}
}

func (p *BluestonePimProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		category.NewDataSource,
	}
}

func (p *BluestonePimProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &BluestonePimProvider{
			version: version,
		}
	}
}
