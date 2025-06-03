package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"

	"github.com/hashicorp/terraform-plugin-framework/types"

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
	_ resource.Resource                = &deviceGroupResource{}
	_ resource.ResourceWithConfigure   = &deviceGroupResource{}
	_ resource.ResourceWithImportState = &deviceGroupResource{}
)

// NewDeviceGroupResource is a helper function to simplify the provider implementation.
func NewDeviceGroupResource() resource.Resource {
	return &deviceGroupResource{}
}

type (
	// deviceGroupResource is the resource implementation.
	deviceGroupResource struct {
		client *librenms.Client
	}

	// deviceGroupModel maps resource schema data to a Go type.
	deviceGroupModel struct {
		ID          types.Int32          `tfsdk:"id"`
		Name        types.String         `tfsdk:"name"`
		Description types.String         `tfsdk:"description"`
		Devices     types.List           `tfsdk:"devices"`
		Rules       jsontypes.Normalized `tfsdk:"rules"`
		Type        types.String         `tfsdk:"type"`
	}
)

// Metadata returns the resource type name.
func (r *deviceGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_devicegroup"
}

// Schema defines the schema for the resource.
func (r *deviceGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int32Attribute{
				Computed:    true,
				Description: "The unique numeric identifier of the LibreNMS device group.",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The device group name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The device group description.",
				Optional:    true,
			},
			// appears to be unused in LibreNMS
			//"pattern": schema.StringAttribute{
			//	Description: "The device group pattern.",
			//	Optional:    true,
			//}
			"type": schema.StringAttribute{
				Description: "The device group type, [`dynamic`, `static`].",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("dynamic", "static"),
				},
			},

			"devices": schema.ListAttribute{
				Description: "The list of device IDs in the group. This is only applicable for static device groups.",
				Optional:    true,
				ElementType: types.Int32Type,
			},

			"rules": schema.StringAttribute{
				Description: "The rules for dynamic device groups, in serialized JSON format. This is only applicable for dynamic device groups." +
					" Using an encoded string supports the arbitrarily-deep nested structure of the LibreNMS rulesets.",
				Optional:   true,
				CustomType: jsontypes.NormalizedType{},
			},
		},
	}
}

// ConfigValidators defines validation rules for the resource configuration.
func (r *deviceGroupResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("devices"),
			path.MatchRoot("rules"),
		),
	}
}

// ValidateConfig validates the resource configuration.
func (r *deviceGroupResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data deviceGroupModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the device group type is dynamic, ensure that rules are provided.
	if data.Type.ValueString() == "dynamic" {
		if data.Rules.IsNull() || data.Rules.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("type"),
				"Missing Dynamic Device Group Rules",
				"The device group type is set to 'dynamic', but no rules are provided. Please define the `rules` attribute with an encoded json.",
			)
			return
		}
	}

	// If the device group type is static, ensure that devices are provided.
	if data.Type.ValueString() == "static" {
		if data.Devices.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("type"),
				"Missing Static Device Group Devices",
				"The device group type is set to 'static', but no devices are provided. "+
					"Please define the `devices` attribute with a list of device IDs.",
			)
			return
		}
	}
}

// Configure sets the provider client for the resource.
func (r *deviceGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *deviceGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan deviceGroupModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the device group using the LibreNMS client.
	payload := &librenms.DeviceGroupCreateRequest{
		Name: plan.Name.ValueString(),
		Description: func() *string {
			if plan.Description.IsNull() {
				return nil
			}
			v := plan.Description.ValueString()
			return &v
		}(),
		Type: plan.Type.ValueString(),
	}

	if payload.Type == "static" {
		payload.Devices = make([]int, 0)
		diags = plan.Devices.ElementsAs(ctx, &payload.Devices, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		// if Type is dynamic, then rules must be provided.
		v := plan.Rules.ValueString()
		payload.Rules = &v
	}

	deviceGroupResp, err := r.client.CreateDeviceGroup(payload)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Device Group",
			fmt.Sprintf("Could not create devicegroup: %s", err),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.Int32Value(int32(deviceGroupResp.ID))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *deviceGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state deviceGroupModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed value from LibreNMS API
	groupResp, err := r.client.GetDeviceGroup(strconv.Itoa(int(state.ID.ValueInt32())))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Device Groups",
			fmt.Sprintf("Could not read LibreNMS devicegroup ID %d: %s", state.ID.ValueInt32(), err.Error()),
		)
		return
	}

	if groupResp == nil {
		resp.Diagnostics.AddError(
			"Error Reading Device Groups",
			"Received nil response when reading device group. Please check the LibreNMS API.",
		)
		return
	}

	if len(groupResp.Groups) != 1 {
		resp.Diagnostics.AddError(
			"Unexpected Device Group Get Response",
			fmt.Sprintf("Expected one device group to be retrieved, got %d device groups. Please check the LibreNMS API.", len(groupResp.Groups)),
		)
		return
	}

	// Overwrite items with refreshed state
	group := groupResp.Groups[0]
	state.ID = types.Int32Value(int32(group.ID))
	state.Name = types.StringValue(group.Name)
	state.Type = types.StringValue(group.Type)

	// possibly null values
	state.Description = types.StringNull()
	if group.Description != nil {
		state.Description = types.StringValue(*group.Description)
	}

	if group.Type == "static" {
		// pull group members list from LibreNMS API
		members, err := r.client.GetDeviceGroupMembers(strconv.Itoa(group.ID))
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Device Group Members",
				fmt.Sprintf("Could not read members for device group ID %d: %s", state.ID.ValueInt32(), err.Error()),
			)
			return
		}

		if members == nil {
			resp.Diagnostics.AddError(
				"Error Reading Device Group Members",
				"Received nil response when reading members for device group. Please check the LibreNMS API.",
			)
			return
		}

		// update devices list
		var devices []types.Int32
		for _, device := range members.Devices {
			devices = append(devices, types.Int32Value(int32(device.ID)))
		}

		var devicesDiags diag.Diagnostics
		state.Devices, devicesDiags = types.ListValueFrom(ctx, types.Int32Type, devices)
		resp.Diagnostics.Append(devicesDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		rules, err := group.Rules.JSON()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Serializing Device Group Rules",
				fmt.Sprintf("Could not serialize rules for device group ID %d: %s", state.ID.ValueInt32(), err.Error()),
			)
			return
		}
		state.Rules = jsontypes.NewNormalizedValue(rules)
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *deviceGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan deviceGroupModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state deviceGroupModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := &librenms.DeviceGroupUpdateRequest{
		Type: plan.Type.ValueString(),
		Description: func() *string {
			if plan.Description.IsNull() {
				return nil
			}
			v := plan.Description.ValueString()
			return &v
		}(),
	}

	// other fields can be sent with every patch, but only update the name if it has changed
	if !plan.Name.Equal(state.Name) {
		payload.Name = plan.Name.ValueString()
	}

	if payload.Type == "static" {
		payload.Devices = make([]int, 0)
		diags = plan.Devices.ElementsAs(ctx, &payload.Devices, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		// if Type is dynamic, then rules must be provided.
		v := plan.Rules.ValueString()
		payload.Rules = &v
	}

	_, err := r.client.UpdateDeviceGroup(strconv.Itoa(int(state.ID.ValueInt32())), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating LibreNMS Device Group",
			"Could not update device group, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *deviceGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deviceGroupModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing group
	_, err := r.client.DeleteDeviceGroup(strconv.Itoa(int(state.ID.ValueInt32())))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting LibreNMS Device Group",
			"Could not delete group, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *deviceGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
