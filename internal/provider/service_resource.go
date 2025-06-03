package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/jokelyo/go-librenms"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_           resource.Resource                = &serviceResource{}
	_           resource.ResourceWithConfigure   = &serviceResource{}
	_           resource.ResourceWithImportState = &serviceResource{}
	_           resource.ResourceWithModifyPlan  = &serviceResource{}
	reServiceID                                  = regexp.MustCompile(`\(#(\d+)\)$`)
)

// NewServiceResource is a helper function to simplify the provider implementation.
func NewServiceResource() resource.Resource {
	return &serviceResource{}
}

type (
	// serviceResource is the resource implementation.
	serviceResource struct {
		client *librenms.Client
	}

	// serviceResourceModel maps resource schema data to a Go type.
	//
	// Display, Location, and LocationID are commented out because they
	// are still possibly null after creation, as discovery may not have completed yet,
	// which causes TF state errors; and there are no reliable defaults for them.
	serviceResourceModel struct {
		ID          types.Int32  `tfsdk:"id"`
		Description types.String `tfsdk:"description"`
		DeviceID    types.Int32  `tfsdk:"device_id"`
		Ignore      types.Bool   `tfsdk:"ignore"`
		Name        types.String `tfsdk:"name"`
		Parameters  types.String `tfsdk:"parameters"`
		Target      types.String `tfsdk:"target"`
		Type        types.String `tfsdk:"type"`
	}
)

// Metadata returns the resource type name.
func (r *serviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

// Schema defines the schema for the resource.
func (r *serviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int32Attribute{
				Computed:    true,
				Description: "The unique numeric identifier of the LibreNMS service.",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "A description of the service.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"device_id": schema.Int32Attribute{
				Description: "The device ID this service is associated to.",
				Required:    true,
			},
			"ignore": schema.BoolAttribute{
				Description: "If true, the service will be ignored by LibreNMS.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The service name.",
				Required:    true,
			},
			"parameters": schema.StringAttribute{
				Description: "The parameters for the service plugin, such as `-C 30,14` for `http`.",
				Required:    true,
			},
			"target": schema.StringAttribute{
				Description: "The target for the service, either an IP address or hostname.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of service to create. Must be a supported plugin type in LibreNMS, such as `http`, `ping`, etc.",
				Required:    true,
			},
		},
	}
}

// ModifyPlan is called when the plan is created or updated. It allows us to prevent changes
// to certain attributes that can't be modified after creation, like device_id.
func (r *serviceResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// If there's no plan or state, this must be a creation or deletion, not an update
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var plan, state serviceResourceModel

	// Retrieve the resource from plan
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve the resource from state
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if device_id is being changed
	if !plan.DeviceID.Equal(state.DeviceID) {
		resp.Diagnostics.AddError(
			"Device ID Change Not Allowed",
			"Changing the device ID of an existing service is not supported. Please delete and recreate the service with the new device ID.",
		)
		return
	}
}

// Configure sets the provider client for the resource.
func (r *serviceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *serviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan serviceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the service using the LibreNMS client.
	payload := &librenms.ServiceCreateRequest{
		Description: plan.Description.ValueString(),
		IP:          plan.Target.ValueString(),
		Ignore:      librenms.Bool(plan.Ignore.ValueBool()),
		Name:        plan.Name.ValueString(),
		Param:       plan.Parameters.ValueString(),
		Type:        plan.Type.ValueString(),
	}

	deviceIdentifier := strconv.Itoa(int(plan.DeviceID.ValueInt32()))
	createResp, err := r.client.CreateService(deviceIdentifier, payload)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Service",
			fmt.Sprintf("Could not create service: %s", err),
		)
		return
	}

	if createResp == nil {
		resp.Diagnostics.AddError(
			"Error Getting Service",
			"Received nil response when getting service. Please check the LibreNMS API.",
		)
		return
	}

	// This is unfortunate, but we have to parse the new ID out of the message response. LibreNMS will allow
	// services that are exact duplicates to be created, so we cannot do a GetServices and check for name.
	//
	// current message format: "Service ping has been added to device 2 (#5)"
	m := reServiceID.FindStringSubmatch(createResp.Message)
	if len(m) != 2 {
		resp.Diagnostics.AddError(
			"Unexpected LibreNMS API Response",
			fmt.Sprintf("Could not parse service ID from message: `%s`. Please check the LibreNMS API.", createResp.Message),
		)
		return
	}
	serviceID, err := strconv.Atoi(m[1])
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Service ID",
			fmt.Sprintf("Could not parse service ID from message: `%s`: %s", createResp.Message, err),
		)
		return
	}

	serviceResp, err := r.client.GetService(serviceID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Service",
			fmt.Sprintf("Could not get service: %s", err),
		)
		return
	}

	if serviceResp == nil {
		resp.Diagnostics.AddError(
			"Error Getting Service",
			"Received nil response when getting service. Please check the LibreNMS API.",
		)
		return
	}

	if len(serviceResp.Services) != 1 {
		resp.Diagnostics.AddError(
			"Unexpected LibreNMS API Response",
			fmt.Sprintf("Expected one service to be retrieved, got %d services. Please check the LibreNMS API.", len(serviceResp.Services)),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.Int32Value(int32(serviceResp.Services[0].ID))
	plan.Description = types.StringValue(serviceResp.Services[0].Description) // set default value if not provided

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *serviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state serviceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed value from LibreNMS API
	serviceResp, err := r.client.GetService(int(state.ID.ValueInt32()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Service",
			fmt.Sprintf("Could not read LibreNMS service ID %d: %s", state.ID.ValueInt32(), err.Error()),
		)
		return
	}

	if serviceResp == nil {
		resp.Diagnostics.AddError(
			"Error Reading Service",
			"Received nil response when creating service. Please check the LibreNMS API.",
		)
		return
	}

	if len(serviceResp.Services) != 1 {
		resp.Diagnostics.AddError(
			"Unexpected Service Get Response",
			fmt.Sprintf("Expected one service to be retrieved, got %d services. Please check the LibreNMS API.", len(serviceResp.Services)),
		)
		return
	}

	// Overwrite items with refreshed state
	state.Description = types.StringValue(serviceResp.Services[0].Description)
	state.DeviceID = types.Int32Value(int32(serviceResp.Services[0].DeviceID))
	state.Ignore = types.BoolValue(bool(serviceResp.Services[0].Ignore))
	state.Name = types.StringValue(serviceResp.Services[0].Name)
	state.Parameters = types.StringValue(serviceResp.Services[0].Param)
	state.Target = types.StringValue(serviceResp.Services[0].IP)
	state.Type = types.StringValue(serviceResp.Services[0].Type)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *serviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan serviceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state serviceResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build a payload of fields that have changed; LibreNMS API only supports partial updates for services.
	updateServiceReq := librenms.NewServiceUpdateRequest()
	hasChanges := false
	if !plan.Description.Equal(state.Description) {
		updateServiceReq.SetDescription(plan.Description.ValueString())
		hasChanges = true
	}
	if !plan.Ignore.Equal(state.Ignore) {
		updateServiceReq.SetIgnore(plan.Ignore.ValueBool())
		hasChanges = true
	}
	if !plan.Name.Equal(state.Name) {
		updateServiceReq.SetName(plan.Name.ValueString())
		hasChanges = true
	}
	if !plan.Parameters.Equal(state.Parameters) {
		updateServiceReq.SetParam(plan.Parameters.ValueString())
		hasChanges = true
	}
	if !plan.Target.Equal(state.Target) {
		updateServiceReq.SetIP(plan.Target.ValueString())
		hasChanges = true
	}

	// Only call the API if there are actual changes to apply
	if hasChanges {
		_, err := r.client.UpdateService(int(plan.ID.ValueInt32()), updateServiceReq)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating LibreNMS Service",
				"Could not update service, unexpected error: "+err.Error(),
			)
			return
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *serviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing service
	_, err := r.client.DeleteService(int(state.ID.ValueInt32()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting LibreNMS Service",
			"Could not delete service, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *serviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
