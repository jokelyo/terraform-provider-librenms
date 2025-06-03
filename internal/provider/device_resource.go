package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/jokelyo/go-librenms"
)

const (
	snmpV1  = "v1"
	snmpV2C = "v2c"
	snmpV3  = "v3"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &deviceResource{}
	_ resource.ResourceWithConfigure   = &deviceResource{}
	_ resource.ResourceWithImportState = &deviceResource{}
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
	//
	// Display, Location, and LocationID are commented out because they
	// are still possibly null after creation, as discovery may not have completed yet,
	// which causes TF state errors; and there are no reliable defaults for them.
	deviceResourceModel struct {
		ID types.Int32 `tfsdk:"id"`
		//Display             types.String         `tfsdk:"display"`
		//Location            types.String         `tfsdk:"location"`
		//LocationID          types.Int32          `tfsdk:"location_id"`
		ForceAdd            types.Bool           `tfsdk:"force_add"`
		Hostname            types.String         `tfsdk:"hostname"`
		OverrideSysLocation types.Bool           `tfsdk:"override_syslocation"`
		PollerGroup         types.Int32          `tfsdk:"poller_group"`
		Port                types.Int32          `tfsdk:"port"`
		PortAssociationMode types.Int32          `tfsdk:"port_association_mode"`
		Transport           types.String         `tfsdk:"transport"`
		SnmpV1              *deviceSNMPV1Model   `tfsdk:"snmp_v1"`
		SnmpV2C             *deviceSNMPV2CModel  `tfsdk:"snmp_v2c"`
		SnmpV3              *deviceSNMPV3Model   `tfsdk:"snmp_v3"`
		ICMPOnly            *deviceICMPOnlyModel `tfsdk:"icmp_only"`
	}

	// deviceSNMPV1Model maps SNMP v1 configuration data to a Go type.
	deviceSNMPV1Model struct {
		Community types.String `tfsdk:"community"`
	}

	// deviceSNMPV1V2CModel maps SNMP v2c configuration data to a Go type.
	deviceSNMPV2CModel struct {
		Community types.String `tfsdk:"community"`
	}

	// deviceSNMPV3Model maps SNMP v3 configuration data to a Go type.
	deviceSNMPV3Model struct {
		AuthAlgorithm   types.String `tfsdk:"auth_algorithm"`
		AuthLevel       types.String `tfsdk:"auth_level"`
		AuthName        types.String `tfsdk:"auth_name"`
		AuthPass        types.String `tfsdk:"auth_pass"`
		CryptoAlgorithm types.String `tfsdk:"crypto_algorithm"`
		CryptoPass      types.String `tfsdk:"crypto_pass"`
	}

	// deviceICMPOnlyModel maps ICMP-only configuration data to a Go type.
	deviceICMPOnlyModel struct {
		Hardware types.String `tfsdk:"hardware"`
		OS       types.String `tfsdk:"os"`
		SysName  types.String `tfsdk:"sys_name"`
	}
)

// Metadata returns the resource type name.
func (r *deviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

// Schema defines the schema for the resource.
func (r *deviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int32Attribute{
				Computed:    true,
				Description: "The unique numeric identifier of the LibreNMS device.",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			//"display": schema.StringAttribute{
			//	Computed:    true,
			//	Description: "A string to display as the name of this device, defaults to hostname.",
			//	Optional:    true,
			//},
			"hostname": schema.StringAttribute{
				Description: "The device hostname or IP address. If hostname, it must have a valid DNS entry.",
				Required:    true,
			},
			//"location": schema.StringAttribute{
			//	Computed:    true,
			//	Description: "The name of the device's location.",
			//	Optional:    true,
			//},
			//"location_id": schema.Int32Attribute{
			//	Computed:    true,
			//	Description: "The ID of the device's location.",
			//	Optional:    true,
			//},
			"force_add": schema.BoolAttribute{
				Optional:    true,
				Description: "If true, the SNMP/ICMP checks will be skipped, and the device will be added immediately.",
			},
			"override_syslocation": schema.BoolAttribute{
				Computed:    true,
				Description: "If true, the device will override the sysLocation value with the one set in LibreNMS.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"poller_group": schema.Int32Attribute{
				Computed:    true,
				Description: "The ID of the poller group to assign this device to. If not set, the default poller group will be used (typically 0).",
				Optional:    true,
				Validators: []validator.Int32{
					int32validator.Between(0, 65535),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"port": schema.Int32Attribute{
				Computed:    true,
				Description: "The SNMP port to use for this device. If not set, the default SNMP port defined in your LibreNMS config will be used.",
				Optional:    true,
				Validators: []validator.Int32{
					int32validator.Between(1, 65535),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"port_association_mode": schema.Int32Attribute{
				Computed: true,
				Description: "The int code of the port association mode to use for this device." +
					" Options are `1 (ifIndex)`, `2 (ifName)`, `3 (ifDesc)`, or `4 (ifAlias)`. If not set, the LibreNMS default is ifIndex `1`.",
				Optional: true,
				Validators: []validator.Int32{
					int32validator.OneOf(1, 2, 3, 4),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			//"snmp_disable": schema.BoolAttribute{
			//	Description: "If true, the device will be added as an ICMP-only device.",
			//	Required:    true,
			//},
			"transport": schema.StringAttribute{
				Computed: true,
				Description: "The transport protocol to use for SNMP communication [`udp`, `tcp`, `udp6`, `tcp6`]." +
					" If not set, the default transport protocol defined in your LibreNMS config will be used.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf("udp", "tcp", "udp6", "tcp6"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"icmp_only": schema.SingleNestedAttribute{
				Description: "Configuration for ICMP-only devices. Disables SNMP polling for the device. Mutually exclusive with other `snmp_` attributes.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"hardware": schema.StringAttribute{
						Computed:    true,
						Description: "The hardware type of the ICMP-only device.",
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"os": schema.StringAttribute{
						Computed:    true,
						Description: "The operating system of the ICMP-only device.",
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"sys_name": schema.StringAttribute{
						Computed:    true,
						Description: "The system name of the ICMP-only device.",
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},

			"snmp_v1": schema.SingleNestedAttribute{
				Description: "Configuration for SNMP v1. Mutually exclusive with other `snmp_` and `icmp_` attributes.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"community": schema.StringAttribute{
						Description: "The SNMP community string for v1",
						Required:    true,
						Sensitive:   true,
					},
				},
			},

			"snmp_v2c": schema.SingleNestedAttribute{
				Description: "Configuration for SNMP v2c. Mutually exclusive with other `snmp_`  and `icmp_` attributes.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"community": schema.StringAttribute{
						Description: "The SNMP community string for v2c.",
						Required:    true,
						Sensitive:   true,
					},
				},
			},

			"snmp_v3": schema.SingleNestedAttribute{
				Description: "Configuration for SNMPv3. Mutually exclusive with other `snmp_`  and `icmp_` attributes.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"auth_algorithm": schema.StringAttribute{
						Description: "The SNMPv3 authentication algorithm [`MD5`, `SHA`, `SHA-224`, `SHA-256`, `SHA-384`, `SHA-512`].",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("MD5", "SHA", "SHA-224", "SHA-256", "SHA-384", "SHA-512"),
						},
					},
					"auth_level": schema.StringAttribute{
						Description: "The SNMPv3 authentication level [`noAuthNoPriv`, `authNoPriv`, `authPriv`].",
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
						Description: "The SNMPv3 encryption algorithm [`DES`, `AES`, `AES-192`, `AES-256`, `AES-256-C`].",
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
		},
	}
}

// ConfigValidators defines validation rules for the resource configuration.
func (r *deviceResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("icmp_only"),
			path.MatchRoot("snmp_v1"),
			path.MatchRoot("snmp_v2c"),
			path.MatchRoot("snmp_v3"),
		),
	}
}

// ValidateConfig validates the resource configuration.
func (r *deviceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data deviceResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// if forceAdd is true, then warn the user that ICMP/SNNP checks will be skipped on the host
	if data.ForceAdd.ValueBool() {
		resp.Diagnostics.AddWarning(
			"Force Add Warning",
			"Force add is set to true, which means the device will be added without SNMP/ICMP checks. "+
				"This may lead to incomplete device information in LibreNMS. Use with caution.",
		)
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

	// Create the device using the LibreNMS client.
	payload := &librenms.DeviceCreateRequest{
		Hostname: plan.Hostname.ValueString(),
	}

	// Set optional fields
	//if !plan.Display.IsNull() {
	//	payload.Display = plan.Display.ValueString()
	//}
	//if !plan.Location.IsNull() {
	//	payload.Location = plan.Location.ValueString()
	//}
	//if !plan.LocationID.IsNull() {
	//	payload.LocationID = int(plan.LocationID.ValueInt32())
	//}
	if !plan.OverrideSysLocation.IsNull() {
		payload.OverrideSysLocation = plan.OverrideSysLocation.ValueBool()
	}
	if !plan.PollerGroup.IsNull() {
		payload.PollerGroup = int(plan.PollerGroup.ValueInt32())
	}
	if !plan.Port.IsNull() {
		payload.Port = int(plan.Port.ValueInt32())
	}
	if !plan.PortAssociationMode.IsNull() {
		payload.PortAssocMode = int(plan.PortAssociationMode.ValueInt32())
	}
	if !plan.Transport.IsNull() {
		payload.Transport = plan.Transport.ValueString()
	}

	if plan.ForceAdd.ValueBool() {
		payload.ForceAdd = true
	}

	if plan.ICMPOnly != nil {
		payload.SNMPDisable = true
		if !plan.ICMPOnly.Hardware.IsNull() {
			payload.Hardware = plan.ICMPOnly.Hardware.ValueString()
		}
		if !plan.ICMPOnly.OS.IsNull() {
			payload.OS = plan.ICMPOnly.OS.ValueString()
		}
		if !plan.ICMPOnly.SysName.IsNull() {
			payload.SysName = plan.ICMPOnly.SysName.ValueString()
		}
	}

	if plan.SnmpV1 != nil {
		payload.SNMPVersion = snmpV1
		payload.SNMPCommunity = plan.SnmpV1.Community.ValueString()
	}

	if plan.SnmpV2C != nil {
		payload.SNMPVersion = snmpV2C
		payload.SNMPCommunity = plan.SnmpV2C.Community.ValueString()
	}

	if plan.SnmpV3 != nil {
		payload.SNMPVersion = snmpV3
		payload.SNMPAuthAlgo = plan.SnmpV3.AuthAlgorithm.ValueString()
		payload.SNMPAuthLevel = plan.SnmpV3.AuthLevel.ValueString()
		payload.SNMPAuthName = plan.SnmpV3.AuthName.ValueString()
		payload.SNMPAuthPass = plan.SnmpV3.AuthPass.ValueString()
		payload.SNMPCrytoAlgo = plan.SnmpV3.CryptoAlgorithm.ValueString()
		payload.SNMPCryptoPass = plan.SnmpV3.CryptoPass.ValueString()
	}

	if _, err := r.client.CreateDevice(payload); err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Device",
			fmt.Sprintf("Could not create device: %s", err),
		)
		return
	}

	// We need to GET the device to get all the fields, as the create response does not return all of them.
	deviceResp, err := r.client.GetDevice(payload.Hostname)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Device",
			fmt.Sprintf("Could not get device: %s", err),
		)
		return
	}

	if deviceResp == nil {
		resp.Diagnostics.AddError(
			"Error Getting Device",
			"Received nil response when getting device. Please check the LibreNMS API.",
		)
		return
	}

	if len(deviceResp.Devices) != 1 {
		resp.Diagnostics.AddError(
			"Unexpected LibreNMS API Response",
			fmt.Sprintf("Expected one device to be retrieved, got %d devices. Please check the LibreNMS API.", len(deviceResp.Devices)),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.Int32Value(int32(deviceResp.Devices[0].DeviceID))
	plan.OverrideSysLocation = types.BoolValue(bool(deviceResp.Devices[0].OverrideSysLocation))
	plan.PollerGroup = types.Int32Value(int32(deviceResp.Devices[0].PollerGroup))
	plan.Port = types.Int32Value(int32(deviceResp.Devices[0].Port))
	plan.PortAssociationMode = types.Int32Value(int32(deviceResp.Devices[0].PortAssociationMode))
	plan.Transport = types.StringValue(deviceResp.Devices[0].Transport)

	// Check optionally null fields that may have been updated by SNMP
	//if deviceResp.Devices[0].Display != nil {
	//	plan.Display = types.StringValue(*deviceResp.Devices[0].Display)
	//}
	//if deviceResp.Devices[0].Location != nil {
	//	plan.Location = types.StringValue(*deviceResp.Devices[0].Location)
	//}
	//if deviceResp.Devices[0].LocationID != nil {
	//	plan.LocationID = types.Int32Value(int32(*deviceResp.Devices[0].LocationID))
	//}
	if plan.ICMPOnly != nil {
		plan.ICMPOnly.Hardware = types.StringValue(deviceResp.Devices[0].Hardware)
		plan.ICMPOnly.OS = types.StringValue(deviceResp.Devices[0].OS)
		plan.ICMPOnly.SysName = types.StringValue(deviceResp.Devices[0].SysName)
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *deviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state deviceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed value from LibreNMS API
	deviceResp, err := r.client.GetDevice(strconv.Itoa(int(state.ID.ValueInt32())))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Device",
			fmt.Sprintf("Could not read LibreNMS device ID %d: %s", state.ID.ValueInt32(), err.Error()),
		)
		return
	}

	if deviceResp == nil {
		resp.Diagnostics.AddError(
			"Error Reading Device",
			"Received nil response when creating device. Please check the LibreNMS API.",
		)
		return
	}

	if len(deviceResp.Devices) != 1 {
		resp.Diagnostics.AddError(
			"Unexpected Device Get Response",
			fmt.Sprintf("Expected one device to be retrieved, got %d devices. Please check the LibreNMS API.", len(deviceResp.Devices)),
		)
		return
	}

	// Overwrite items with refreshed state
	state.Hostname = types.StringValue(deviceResp.Devices[0].Hostname)
	state.OverrideSysLocation = types.BoolValue(bool(deviceResp.Devices[0].OverrideSysLocation))
	state.PollerGroup = types.Int32Value(int32(deviceResp.Devices[0].PollerGroup))
	state.Port = types.Int32Value(int32(deviceResp.Devices[0].Port))
	state.PortAssociationMode = types.Int32Value(int32(deviceResp.Devices[0].PortAssociationMode))
	state.Transport = types.StringValue(deviceResp.Devices[0].Transport)

	// possibly null fields
	//state.Display = types.StringNull()
	//if deviceResp.Devices[0].Display != nil {
	//	state.Display = types.StringValue(*deviceResp.Devices[0].Display)
	//}
	//
	//state.Location = types.StringNull()
	//if deviceResp.Devices[0].Location != nil {
	//	state.Display = types.StringValue(*deviceResp.Devices[0].Location)
	//}
	//
	//state.LocationID = types.Int32Null()
	//if deviceResp.Devices[0].LocationID != nil {
	//	state.LocationID = types.Int32Value(int32(*deviceResp.Devices[0].LocationID))
	//}

	state.ICMPOnly = nil
	state.SnmpV1 = nil
	state.SnmpV2C = nil
	state.SnmpV3 = nil
	if deviceResp.Devices[0].SNMPDisable {
		state.ICMPOnly = &deviceICMPOnlyModel{
			Hardware: types.StringValue(deviceResp.Devices[0].Hardware),
			OS:       types.StringValue(deviceResp.Devices[0].OS),
			SysName:  types.StringValue(deviceResp.Devices[0].SysName),
		}
	} else {
		switch deviceResp.Devices[0].SNMPVersion {
		case snmpV1:
			state.SnmpV1 = stateSNMPV1(deviceResp.Devices[0])
		case snmpV2C:
			state.SnmpV2C = stateSNMPV2C(deviceResp.Devices[0])
		case snmpV3:
			state.SnmpV3 = stateSNMPV3(deviceResp.Devices[0])
		}
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *deviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan deviceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state deviceResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := new(librenms.DeviceUpdateRequest)

	// Build a payload of fields that have changed; LibreNMS API only supports partial updates.
	if !plan.OverrideSysLocation.IsUnknown() && !plan.OverrideSysLocation.Equal(state.OverrideSysLocation) {
		payload.Field = append(payload.Field, "override_sysLocation")
		payload.Data = append(payload.Data, librenms.Bool(plan.OverrideSysLocation.ValueBool()))
	}
	if !plan.PollerGroup.IsUnknown() && !plan.PollerGroup.Equal(state.PollerGroup) {
		payload.Field = append(payload.Field, "poller_group")
		payload.Data = append(payload.Data, int(plan.PollerGroup.ValueInt32()))
	}
	if !plan.Port.IsUnknown() && !plan.Port.Equal(state.Port) {
		payload.Field = append(payload.Field, "port")
		payload.Data = append(payload.Data, int(plan.Port.ValueInt32()))
	}
	if !plan.PortAssociationMode.IsUnknown() && !plan.PortAssociationMode.Equal(state.PortAssociationMode) {
		payload.Field = append(payload.Field, "port_association_mode")
		payload.Data = append(payload.Data, int(plan.PortAssociationMode.ValueInt32()))
	}
	if !plan.Transport.IsUnknown() && !plan.Transport.Equal(state.Transport) {
		payload.Field = append(payload.Field, "transport")
		payload.Data = append(payload.Data, plan.Transport.ValueString())
	}

	if plan.ICMPOnly != nil && state.ICMPOnly == nil {
		payload.Field = append(payload.Field, "snmp_disable")
		payload.Data = append(payload.Data, librenms.Bool(true))
		if !plan.ICMPOnly.Hardware.IsNull() {
			payload.Field = append(payload.Field, "hardware")
			payload.Data = append(payload.Data, plan.ICMPOnly.Hardware.ValueString())
		}
		if !plan.ICMPOnly.OS.IsNull() {
			payload.Field = append(payload.Field, "os")
			payload.Data = append(payload.Data, plan.ICMPOnly.OS.ValueString())
		}
		if !plan.ICMPOnly.SysName.IsNull() {
			payload.Field = append(payload.Field, "sys_name")
			payload.Data = append(payload.Data, plan.ICMPOnly.SysName.ValueString())
		}
	} else if plan.ICMPOnly == nil && state.ICMPOnly != nil {
		payload.Field = append(payload.Field, "snmp_disable")
		payload.Data = append(payload.Data, librenms.Bool(false))
	}

	if plan.SnmpV1 != nil && state.SnmpV1 == nil {
		payload.Field = append(payload.Field, "snmpver")
		payload.Data = append(payload.Data, snmpV1)
		payload.Field = append(payload.Field, "community")
		payload.Data = append(payload.Data, plan.SnmpV1.Community.ValueString())
	}

	if plan.SnmpV2C != nil && state.SnmpV2C == nil {
		payload.Field = append(payload.Field, "snmpver")
		payload.Data = append(payload.Data, snmpV2C)
		payload.Field = append(payload.Field, "community")
		payload.Data = append(payload.Data, plan.SnmpV2C.Community.ValueString())
	}

	if plan.SnmpV3 != nil && state.SnmpV3 == nil {
		payload.Field = append(payload.Field, "snmpver")
		payload.Data = append(payload.Data, snmpV3)
		payload.Field = append(payload.Field, "authalgo")
		payload.Data = append(payload.Data, plan.SnmpV3.AuthAlgorithm.ValueString())
		payload.Field = append(payload.Field, "authlevel")
		payload.Data = append(payload.Data, plan.SnmpV3.AuthLevel.ValueString())
		payload.Field = append(payload.Field, "authname")
		payload.Data = append(payload.Data, plan.SnmpV3.AuthName.ValueString())
		payload.Field = append(payload.Field, "authpass")
		payload.Data = append(payload.Data, plan.SnmpV3.AuthPass.ValueString())
		payload.Field = append(payload.Field, "cryptoalgo")
		payload.Data = append(payload.Data, plan.SnmpV3.CryptoAlgorithm.ValueString())
		payload.Field = append(payload.Field, "cryptopass")
		payload.Data = append(payload.Data, plan.SnmpV3.CryptoPass.ValueString())
	}

	// If no relevant fields have changed, treat it as a no-op update.
	if len(payload.Field) > 0 {
		_, err := r.client.UpdateDevice(plan.Hostname.ValueString(), payload)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating LibreNMS Device",
				"Could not update device, unexpected error: "+err.Error(),
			)
			return
		}
	}

	// Get updated device record from LibreNMS API
	deviceResp, err := r.client.GetDevice(plan.Hostname.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Device",
			fmt.Sprintf("Could not read LibreNMS device ID %d: %s", state.ID.ValueInt32(), err.Error()),
		)
		return
	}

	if deviceResp == nil {
		resp.Diagnostics.AddError(
			"Error Reading Device",
			"Received nil response when creating device. Please check the LibreNMS API.",
		)
		return
	}

	if len(deviceResp.Devices) != 1 {
		resp.Diagnostics.AddError(
			"Unexpected Device Get Response",
			fmt.Sprintf("Expected one device to be retrieved, got %d devices. Please check the LibreNMS API.", len(deviceResp.Devices)),
		)
		return
	}

	// Update resource state
	plan.OverrideSysLocation = types.BoolValue(bool(deviceResp.Devices[0].OverrideSysLocation))
	plan.PollerGroup = types.Int32Value(int32(deviceResp.Devices[0].PollerGroup))
	plan.Port = types.Int32Value(int32(deviceResp.Devices[0].Port))
	plan.PortAssociationMode = types.Int32Value(int32(deviceResp.Devices[0].PortAssociationMode))
	plan.Transport = types.StringValue(deviceResp.Devices[0].Transport)

	// Check optionally null fields that may have been updated by SNMP
	//if deviceResp.Devices[0].Display != nil {
	//	plan.Display = types.StringValue(*deviceResp.Devices[0].Display)
	//}
	//if deviceResp.Devices[0].Location != nil {
	//	plan.Location = types.StringValue(*deviceResp.Devices[0].Location)
	//}
	//if deviceResp.Devices[0].LocationID != nil {
	//	plan.LocationID = types.Int32Value(int32(*deviceResp.Devices[0].LocationID))
	//}
	if plan.ICMPOnly != nil {
		plan.ICMPOnly.Hardware = types.StringValue(deviceResp.Devices[0].Hardware)
		plan.ICMPOnly.OS = types.StringValue(deviceResp.Devices[0].OS)
		plan.ICMPOnly.SysName = types.StringValue(deviceResp.Devices[0].SysName)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *deviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deviceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing device
	_, err := r.client.DeleteDevice(state.Hostname.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting LibreNMS Device",
			"Could not delete device, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *deviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 32)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing ID for Import",
			fmt.Sprintf("Expected a numeric ID for import, but got %q: %s", req.ID, err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int32Value(int32(id)))...)
}

func stateSNMPV1(device librenms.Device) *deviceSNMPV1Model {
	ret := &deviceSNMPV1Model{
		Community: types.StringNull(),
	}
	if device.Community != nil {
		ret.Community = types.StringValue(*device.Community)
	}
	return ret
}

func stateSNMPV2C(device librenms.Device) *deviceSNMPV2CModel {
	ret := &deviceSNMPV2CModel{
		Community: types.StringNull(),
	}
	if device.Community != nil {
		ret.Community = types.StringValue(*device.Community)
	}
	return ret
}

func stateSNMPV3(device librenms.Device) *deviceSNMPV3Model {
	ret := &deviceSNMPV3Model{
		AuthAlgorithm:   types.StringNull(),
		AuthLevel:       types.StringNull(),
		AuthName:        types.StringNull(),
		AuthPass:        types.StringNull(),
		CryptoAlgorithm: types.StringNull(),
		CryptoPass:      types.StringNull(),
	}

	if device.AuthAlgorithm != nil {
		ret.AuthAlgorithm = types.StringValue(*device.AuthAlgorithm)
	}
	if device.AuthLevel != nil {
		ret.AuthLevel = types.StringValue(*device.AuthLevel)
	}
	if device.AuthName != nil {
		ret.AuthName = types.StringValue(*device.AuthName)
	}
	if device.AuthPass != nil {
		ret.AuthPass = types.StringValue(*device.AuthPass)
	}
	if device.CryptoAlgorithm != nil {
		ret.CryptoAlgorithm = types.StringValue(*device.CryptoAlgorithm)
	}
	if device.CryptoPass != nil {
		ret.CryptoPass = types.StringValue(*device.CryptoPass)
	}
	return ret
}
