// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/jokelyo/go-librenms"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &librenmsProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &librenmsProvider{
			version: version,
		}
	}
}

// librenmsProvider is the provider implementation.
type librenmsProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// librenmsProviderModel maps provider schema data to a Go type.
type librenmsProviderModel struct {
	Host  types.String `tfsdk:"host"`
	Token types.String `tfsdk:"token"`
}

// Metadata returns the provider type name.
func (p *librenmsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "librenms"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *librenmsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "The LibreNMS API base URL, supported format `http[s]://hostname[:port]/`." +
					" May also be set using the `LIBRENMS_HOST` environment variable.",
				Optional: true,
			},
			"token": schema.StringAttribute{
				Description: "The LibreNMS API token. May also be set using the `LIBRENMS_TOKEN` environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure prepares a LibreNMS API client for data sources and resources.
func (p *librenmsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring LibreNMS client")

	var config librenmsProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown LibreNMS API Host",
			"The provider cannot create the LibreNMS API client as there is an unknown configuration value for the LibreNMS API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the LIBRENMS_HOST environment variable.",
		)
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown LibreNMS API token",
			"The provider cannot create the LibreNMS API client as there is an unknown configuration value for the LibreNMS API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the LIBRENMS_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("LIBRENMS_HOST")
	token := os.Getenv("LIBRENMS_TOKEN")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing LibreNMS API Host",
			"The provider cannot create the LibreNMS API client as there is a missing or empty value for the LibreNMS API host. "+
				"Set the host value in the configuration or use the LIBRENMS_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing LibreNMS API Token",
			"The provider cannot create the LibreNMS API client as there is a missing or empty value for the LibreNMS API password. "+
				"Set the password value in the configuration or use the LIBRENMS_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "librenms_host", host)
	ctx = tflog.SetField(ctx, "librenms_token", token)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "librenms_token")
	tflog.Debug(ctx, "Creating LibreNMS client")

	// Create a new LibreNMS client using the configuration values
	client, err := librenms.New(host, token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create LibreNMS API Client",
			"An unexpected error occurred when creating the LibreNMS API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"LibreNMS Client Error: "+err.Error(),
		)
		return
	}

	// Make the LibreNMS client available during DataSource and Resource type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured LibreNMS client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *librenmsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

// Resources defines the resources implemented in the provider.
func (p *librenmsProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDeviceResource,
		NewDeviceGroupResource,
		NewAlertRuleResource,
		NewServiceResource,
	}
}
