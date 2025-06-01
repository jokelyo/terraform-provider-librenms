package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/jokelyo/go-librenms"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &deviceResource{}
	_ resource.ResourceWithConfigure = &deviceResource{}
)

// NewDeviceResource is a helper function to simplify the provider implementation.
func NewDeviceResource() resource.Resource {
	return &deviceResource{}
}

type (
	// deviceResource is the resource implementation.
	deviceResource struct {
		client *librenms.Client
	}

	// deviceResourceModel maps resource schema data to a Go type.
	deviceResourceModel struct {
		ID                  string                `tfsdk:"id"`
		Display             string                `tfsdk:"display"`
		Hostname            string                `tfsdk:"hostname"`
		Location            string                `tfsdk:"location"`
		LocationID          int64                 `tfsdk:"location_id"`
		OverrideSysLocation bool                  `tfsdk:"override_sysLocation"`
		PollerGroup         int32                 `tfsdk:"poller_group"`
		Port                int64                 `tfsdk:"port"`
		PortAssociationMode string                `tfsdk:"port_association_mode"`
		Transport           string                `tfsdk:"transport"`
		SnmpV1V2C           *deviceSNMPV1V2CModel `tfsdk:"snmp_v1v2c"`
		SnmpV3              *deviceSNMPV3Model    `tfsdk:"snmp_v3"`
		IcmpOnly            *deviceICMPOnlyModel  `tfsdk:"icmp_only"`
	}

	// deviceSNMPV1V2CModel maps SNMP v1/v2c configuration data to a Go type.
	deviceSNMPV1V2CModel struct {
		Community string `tfsdk:"community"`
	}

	// deviceSNMPV3Model maps SNMP v3 configuration data to a Go type.
	deviceSNMPV3Model struct {
		AuthAlgorithm   string `tfsdk:"auth_algorithm"`
		AuthLevel       string `tfsdk:"auth_level"`
		AuthName        string `tfsdk:"auth_name"`
		AuthPass        string `tfsdk:"auth_pass"`
		CryptoAlgorithm string `tfsdk:"crypto_algorithm"`
		CryptoPass      string `tfsdk:"crypto_pass"`
	}

	// deviceICMPOnlyModel maps ICMP (Ping-only) configuration data to a Go type.
	deviceICMPOnlyModel struct {
		Hardware string `tfsdk:"hardware"`
		OS       string `tfsdk:"os"`
		SysName  string `tfsdk:"sys_name"`
	}
)

// Metadata returns the resource type name.
func (r *deviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_order"
}

// Schema defines the schema for the resource.
func (r *deviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"display": schema.StringAttribute{
				Description: "A string to display as the name of this device, defaults to hostname.",
				Optional:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "The device hostname or IP address.",
				Required:    true,
			},
			"location": schema.StringAttribute{
				Description: "The name of the device's location.",
				Optional:    true,
			},
			"location_id": schema.Int64Attribute{
				Description: "The ID of the device's location.",
				Optional:    true,
			},
			"override_sysLocation": schema.BoolAttribute{
				Optional: true,
			},
			"poller_group": schema.Int32Attribute{
				Description: "The ID of the poller group to assign this device to. Defaults to 0.",
				Optional:    true,
			},
			"port": schema.Int64Attribute{
				Description: "The SNMP port to use for this device. Defaults to port defined in config.",
				Optional:    true,
			},
			"port_association_mode": schema.StringAttribute{
				Description: "The port association mode to use for this device. Defaults to ifIndex.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("ifIndex", "ifName", "ifAlias", "ifDescr"),
				},
			},
			"transport": schema.StringAttribute{
				Description: "The transport protocol to use for SNMP communication. Defaults to transport defined in config.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("udp", "tcp", "udp6", "tcp6"),
				},
			},

			"snmp_v1v2c": schema.SingleNestedAttribute{
				Description: "Configuration for SNMP v1/v2c. Mutually exclusive with `snmp_v3` and `icmp` blocks.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"community": schema.StringAttribute{
						Description: "The SNMP community string for v1/v2c.",
						Required:    true,
					},
				},
			},

			"snmp_v3": schema.SingleNestedAttribute{
				Description: "Configuration for SNMPv3. Mutually exclusive with `snmp_v1v2c` and `icmp` blocks.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"auth_algorithm": schema.StringAttribute{
						Description: "The SNMPv3 authentication algorithm.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("MD5", "SHA", "SHA-224", "SHA-256", "SHA-384", "SHA-512"),
						},
					},
					"auth_level": schema.StringAttribute{
						Description: "The SNMPv3 authentication level.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("noAuthNoPriv", "authNoPriv", "authPriv"),
						},
					},
					"auth_name": schema.StringAttribute{
						Description: "The SNMPv3 authentication username.",
						Required:    true,
						Sensitive:   true,
					},
					"auth_pass": schema.StringAttribute{
						Description: "The SNMPv3 authentication password.",
						Required:    true,
						Sensitive:   true,
					},
					"crypto_algorithm": schema.StringAttribute{
						Description: "The SNMPv3 encryption algorithm.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("DES", "AES", "AES-192", "AES-256", "AES-256-C"),
						},
					},
					"crypto_pass": schema.StringAttribute{
						Description: "The SNMPv3 encryption password.",
						Required:    true,
						Sensitive:   true,
					},
				},
			},

			"icmp_only": schema.SingleNestedAttribute{
				Description: "Configuration for ICMP (Ping-only). Mutually exclusive with `snmp_v1v2c` and `snmp_v3` blocks.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"hardware": schema.StringAttribute{
						Description: "The user-defined hardware type.",
						Optional:    true,
					},
					"os": schema.StringAttribute{
						Description: "The user-defined OS. Defaults to 'ping'.",
						Optional:    true,
					},
					"sys_name": schema.StringAttribute{
						Description: "The user-defined value for sysName.",
						Optional:    true,
					},
				},
			},
		},
	}
}

func (r *deviceResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("snmp_v1v2c"),
			path.MatchRoot("snmp_v3"),
			path.MatchRoot("icmp"),
		),
	}
}

// Configure sets the provider client for the resource.
func (r *deviceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*librenms.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *librenms.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *deviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan deviceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var items []hashicups.OrderItem
	for _, item := range plan.Items {
		items = append(items, hashicups.OrderItem{
			Coffee: hashicups.Coffee{
				ID: int(item.Coffee.ID.ValueInt64()),
			},
			Quantity: int(item.Quantity.ValueInt64()),
		})
	}

	// Create the device using the LibreNMS client.
	device, err := r.client.CreateDevice(librenms.CreateDeviceOptions{
		Display:             plan.Display,
		Hostname:            plan.Hostname,
		Location:            plan.Location,
		LocationID:          plan.LocationID,
		OverrideSysLocation: plan.OverrideSysLocation,
		PollerGroup:         plan.PollerGroup,
		Port:                plan.Port,
		PortAssociationMode: plan.PortAssociationMode,
		Transport:           plan.Transport,
		SnmpV1V2C:           plan.SnmpV1V2C,
		SnmpV3:              plan.SnmpV3,
		IcmpOnly:            plan.IcmpOnly,
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Device",
			fmt.Sprintf("Could not create device: %s", err),
		)
		return
	}

	// Set the ID and other attributes in the state.
	state := deviceResourceModel{
		ID:          fmt.Sprintf("%d", device.ID),
		Display:     device.Display,
		Hostname:    device.Hostname,
		Location:    device.Location,
		LocationID:  device.LocationID,
		PollerGroup: device.PollerGroup,
		SnmpV1V2C:   device.SnmpV1V2C,
		SnmpV3:      device.SnmpV3,
		IcmpOnly:    device.IcmpOnly,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *deviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *deviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *deviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
