package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/bluestonepim-go-sdk/global_settings"
	"github.com/labd/bluestonepim-go-sdk/notification_external"
	"github.com/labd/bluestonepim-go-sdk/pim"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/attribute_definition"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/attribute_definition_context"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/attribute_group"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/attribute_group_context"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/attribute_group_definition_order"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/attribute_group_order"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/category"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/category_attribute"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/category_column_attribute"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/category_compound_attribute"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/category_context"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/category_dictionary_attribute"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/category_matrix_attribute"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/column_attribute_definition"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/compound_attribute_definition"
	bpcontext "github.com/labd/terraform-provider-bluestonepim/internal/resources/context"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/dictionary_attribute_definition"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/dictionary_attribute_value"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/matrix_attribute_definition"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/relation"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/webhook"
	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// Ensure BluestonePimProvider satisfies various provider interfaces.
var _ provider.Provider = &BluestonePimProvider{}

// BluestonePimProvider defines the provider implementation.
type BluestonePimProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
	debug   bool
}

// BluestonePimProviderModel describes the provider data model.
type BluestonePimProviderModel struct {
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	AuthURL      types.String `tfsdk:"auth_url"`
	ApiURL       types.String `tfsdk:"api_url"`
}

func (p *BluestonePimProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "bluestonepim"
	resp.Version = p.version
}

func (p *BluestonePimProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				MarkdownDescription: "The client id for Bluestone Platform API",
				Optional:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "The client secret for Bluestone Platform API",
				Optional:            true,
				Sensitive:           true,
			},
			"auth_url": schema.StringAttribute{
				MarkdownDescription: "The authentication URL of the Bluestone Platform API",
				Optional:            true,
			},
			"api_url": schema.StringAttribute{
				MarkdownDescription: "The api URL of the Bluestone Platform API",
				Optional:            true,
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

	var clientID = utils.GetEnv("BP_CLIENT_ID", "")
	var clientSecret = utils.GetEnv("BP_CLIENT_SECRET", "")
	var apiURL = utils.GetEnv("BP_API_URL", "https://api.bluestonepim.com")
	var authURL = utils.GetEnv("BP_AUTH_URL", "https://idp.bluestonepim.com/op/token")

	if data.ClientID.ValueString() != "" {
		clientID = data.ClientID.ValueString()
	}

	if data.ClientSecret.ValueString() != "" {
		clientSecret = data.ClientSecret.ValueString()
	}

	if data.ApiURL.ValueString() != "" {
		apiURL = data.ApiURL.ValueString()
	}

	if data.AuthURL.ValueString() != "" {
		authURL = data.AuthURL.ValueString()
	}

	if clientID == "" {
		resp.Diagnostics.AddError(
			"Missing Client ID Configuration",
			"While configuring the provider, the API token was not found in "+
				"the BP_CLIENT_ID environment variable or provider "+
				"configuration block client_id attribute.",
		)
	}

	if clientSecret == "" {
		resp.Diagnostics.AddError(
			"Missing Client Secret Configuration",
			"While configuring the provider, the API token was not found in "+
				"the BP_CLIENT_SECRET environment variable or provider "+
				"configuration block client_secret attribute.",
		)
	}

	retryableClient := newRetryableClient(retryMax, retryWaitMin, retryWaitMax)

	oauth2Config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     authURL,
	}

	httpClient := oauth2Config.Client(
		context.WithValue(context.Background(), oauth2.HTTPClient, retryableClient.StandardClient()),
	)

	pimClient, err := pim.NewClientWithResponses(
		fmt.Sprintf("%s/pim", apiURL),
		pim.WithHTTPClient(httpClient),
	)
	if err != nil {
		panic(err)
	}

	notificationsClient, err := notification_external.NewClientWithResponses(
		fmt.Sprintf("%s/notification-external", apiURL),
		notification_external.WithHTTPClient(httpClient),
	)
	if err != nil {
		panic(err)
	}

	globalSettingsClient, err := global_settings.NewClientWithResponses(
		fmt.Sprintf("%s/global-settings", apiURL),
		global_settings.WithHTTPClient(httpClient),
	)
	if err != nil {
		panic(err)
	}

	container := &utils.ProviderData{
		PimClient:            pimClient,
		NotificationClient:   notificationsClient,
		GlobalSettingsClient: globalSettingsClient,
	}

	// Example client configuration for data sources and resources
	resp.DataSourceData = container
	resp.ResourceData = container
}

func (p *BluestonePimProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		category.NewResource,
		category_context.NewResource,
		attribute_definition.NewResource,
		attribute_definition_context.NewResource,
		attribute_group.NewResource,
		attribute_group_context.NewResource,
		attribute_group_order.NewResource,
		attribute_group_definition_order.NewResource,
		column_attribute_definition.NewResource,
		matrix_attribute_definition.NewResource,
		dictionary_attribute_definition.NewResource,
		dictionary_attribute_value.NewResource,
		compound_attribute_definition.NewResource,
		category_attribute.NewResource,
		category_column_attribute.NewResource,
		category_matrix_attribute.NewResource,
		category_dictionary_attribute.NewResource,
		category_compound_attribute.NewResource,
		webhook.NewResource,
		relation.NewResource,
		bpcontext.NewResource,
	}
}

func (p *BluestonePimProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		category.NewDataSource,
	}
}

func New(version string, debug bool) func() provider.Provider {
	return func() provider.Provider {
		return &BluestonePimProvider{
			version: version,
			debug:   debug,
		}
	}
}
