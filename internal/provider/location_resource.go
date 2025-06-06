package provider

import (
	"context"
	"fmt"
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
	_ resource.Resource                = &locationResource{}
	_ resource.ResourceWithConfigure   = &locationResource{}
	_ resource.ResourceWithImportState = &locationResource{}
)

// NewLocationResource is a helper function to simplify the provider implementation.
func NewLocationResource() resource.Resource {
	return &locationResource{}
}

type (
	// locationResource is the resource implementation.
	locationResource struct {
		client *librenms.Client
	}

	// locationResourceModel maps resource schema data to a Go type.
	locationResourceModel struct {
		ID               types.Int32   `tfsdk:"id"`
		FixedCoordinates types.Bool    `tfsdk:"fixed_coordinates"`
		Latitude         types.Float64 `tfsdk:"latitude"`
		Longitude        types.Float64 `tfsdk:"longitude"`
		Name             types.String  `tfsdk:"name"`
		Timestamp        types.String  `tfsdk:"timestamp"`
	}
)

// Metadata returns the resource type name.
func (r *locationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_location"
}

// Schema defines the schema for the resource.
func (r *locationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int32Attribute{
				Computed:    true,
				Description: "The unique numeric identifier of the LibreNMS location.",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"fixed_coordinates": schema.BoolAttribute{
				Description: "If true, the location will use fixed coordinates instead of discovered ones.",
				Optional:    true,
			},
			"latitude": schema.Float64Attribute{
				Description: "The latitude of the location.",
				Required:    true,
			},
			"longitude": schema.Float64Attribute{
				Description: "The longitude of the location.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The location name.",
				Required:    true,
			},
			"timestamp": schema.StringAttribute{
				Description: "The timestamp of the location creation.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure sets the provider client for the resource.
func (r *locationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *locationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan locationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the location using the LibreNMS client.
	payload := &librenms.LocationCreateRequest{
		FixedCoordinates: librenms.Bool(plan.FixedCoordinates.ValueBool()),
		Latitude:         plan.Latitude.ValueFloat64(),
		Longitude:        plan.Longitude.ValueFloat64(),
		Name:             plan.Name.ValueString(),
	}

	_, err := r.client.CreateLocation(payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Location",
			fmt.Sprintf("Could not create location: %s", err),
		)
		return
	}

	// GetLocations to get the computed values.
	// Name is unique, so we can use it to find the created location.
	locationsResp, err := r.client.GetLocations()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Locations",
			fmt.Sprintf("Could not get locations: %s", err),
		)
		return
	}

	if locationsResp == nil {
		resp.Diagnostics.AddError(
			"Error Getting Locations",
			"Received nil response when getting locations. Please check the LibreNMS API.",
		)
		return
	}

	if len(locationsResp.Locations) == 0 {
		resp.Diagnostics.AddError(
			"Unexpected LibreNMS API Response",
			"Expected > 0 locations to be retrieved, got 0 locations. Please check LibreNMS.",
		)
		return
	}

	var location librenms.Location
	for _, loc := range locationsResp.Locations {
		if loc.Name == plan.Name.ValueString() {
			location = loc
		}
	}

	if location.ID == 0 {
		resp.Diagnostics.AddError(
			"Unexpected LibreNMS API Response",
			fmt.Sprintf("Unable to find location with name '%s'. Please check LibreNMS.", plan.Name.ValueString()),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.Int32Value(int32(location.ID))
	plan.Timestamp = types.StringValue(location.Timestamp)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *locationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state locationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed value from LibreNMS API
	locationResp, err := r.client.GetLocation(int(state.ID.ValueInt32()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Location",
			fmt.Sprintf("Could not retrieve LibreNMS location ID %d: %s", state.ID.ValueInt32(), err.Error()),
		)
		return
	}

	if locationResp == nil {
		resp.Diagnostics.AddError(
			"Error Reading Location",
			"Received nil response when getting location. Please check the LibreNMS API.",
		)
		return
	}

	// Overwrite items with refreshed state
	state.FixedCoordinates = types.BoolValue(bool(locationResp.Location.FixedCoordinates))
	state.Name = types.StringValue(locationResp.Location.Name)
	state.Latitude = types.Float64Value(float64(locationResp.Location.Latitude))
	state.Longitude = types.Float64Value(float64(locationResp.Location.Longitude))
	state.Timestamp = types.StringValue(locationResp.Location.Timestamp)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *locationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan locationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state locationResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build a payload of fields that have changed; LibreNMS API only supports partial updates for locations.
	updateLocationReq := librenms.NewLocationUpdateRequest()
	hasChanges := false
	if !plan.FixedCoordinates.Equal(state.FixedCoordinates) {
		updateLocationReq.SetFixedCoordinates(plan.FixedCoordinates.ValueBool())
		hasChanges = true
	}
	if !plan.Latitude.Equal(state.Latitude) {
		updateLocationReq.SetLatitude(plan.Latitude.ValueFloat64())
		hasChanges = true
	}
	if !plan.Longitude.Equal(state.Longitude) {
		updateLocationReq.SetLongitude(plan.Longitude.ValueFloat64())
		hasChanges = true
	}
	if !plan.Name.Equal(state.Name) {
		updateLocationReq.SetName(plan.Name.ValueString())
		hasChanges = true
	}

	// Only call the API if there are actual changes to apply
	if hasChanges {
		_, err := r.client.UpdateLocation(int(plan.ID.ValueInt32()), updateLocationReq)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating LibreNMS Location",
				"Could not update location, unexpected error: "+err.Error(),
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
func (r *locationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state locationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing location
	_, err := r.client.DeleteLocation(int(state.ID.ValueInt32()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting LibreNMS Location",
			"Could not delete location, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *locationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
