package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

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
	_ resource.Resource                = &alertRuleResource{}
	_ resource.ResourceWithConfigure   = &alertRuleResource{}
	_ resource.ResourceWithImportState = &alertRuleResource{}
)

// NewAlertRuleResource is a helper function to simplify the provider implementation.
func NewAlertRuleResource() resource.Resource {
	return &alertRuleResource{}
}

type (
	// alertRuleResource is the resource implementation.
	alertRuleResource struct {
		client *librenms.Client
	}

	// alertRuleModel maps resource schema data to a Go type.
	alertRuleModel struct {
		ID           types.Int32          `tfsdk:"id"`
		Builder      jsontypes.Normalized `tfsdk:"builder"`
		Delay        types.String         `tfsdk:"delay"`
		Devices      types.List           `tfsdk:"devices"`
		Disabled     types.Bool           `tfsdk:"disabled"`
		Extra        jsontypes.Normalized `tfsdk:"extra"`
		Groups       types.List           `tfsdk:"groups"`
		Interval     types.String         `tfsdk:"interval"`
		Locations    types.List           `tfsdk:"locations"`
		MaxAlerts    types.Int32          `tfsdk:"max_alerts"` // `count` is a reserved root attribute
		Mute         types.Bool           `tfsdk:"mute"`
		Name         types.String         `tfsdk:"name"`
		Notes        types.String         `tfsdk:"notes"`
		ProcedureURL types.String         `tfsdk:"procedure_url"`
		Query        types.String         `tfsdk:"query"`
		Severity     types.String         `tfsdk:"severity"`
	}
)

// Metadata returns the resource type name.
func (r *alertRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alertrule"
}

// Schema defines the schema for the resource.
func (r *alertRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int32Attribute{
				Computed:    true,
				Description: "The unique numeric identifier of the LibreNMS alert rule.",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"builder": schema.StringAttribute{
				Description: "The alert rule builder field defines the rule logic in serialized JSON format.",
				Required:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"delay": schema.StringAttribute{
				Description: "The delay before the alert rule is triggered, in a format like `5m` or `1h`.",
				Optional:    true,
			},
			"devices": schema.ListAttribute{
				Description: "The list of device IDs attached to the alert rule. If not set, the rule applies to all devices.",
				Optional:    true,
				ElementType: types.Int32Type,
			},
			"disabled": schema.BoolAttribute{
				Description: "Whether the alert rule is disabled.",
				Required:    true,
			},
			"extra": schema.StringAttribute{
				Description: "Extra information stored in serialized JSON format. This is set by LibreNMS.",
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"groups": schema.ListAttribute{
				Description: "The list of group IDs attached to the alert rule. This can be defined alongside `devices` and `locations`.",
				Optional:    true,
				ElementType: types.Int32Type,
			},
			"interval": schema.StringAttribute{
				Description: "The interval at which the alert rule is checked, in a format like `5m` or `1h`.",
				Optional:    true,
			},
			"locations": schema.ListAttribute{
				Description: "The list of location IDs attached to the alert rule. This can be defined alongside `devices` and `groups`.",
				Optional:    true,
				ElementType: types.Int32Type,
			},
			"max_alerts": schema.Int32Attribute{
				Description: "The number of times the alert rule will send an alert.",
				Optional:    true,
				Validators: []validator.Int32{
					int32validator.AtLeast(0),
				},
			},
			"mute": schema.BoolAttribute{
				Description: "Whether the alert rule is muted. Muted rules do not trigger alerts.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "The alert rule name.",
				Required:    true,
			},
			"notes": schema.StringAttribute{
				Description: "The alert rule notes.",
				Optional:    true,
			},
			"procedure_url": schema.StringAttribute{
				Description: "A procedure URL (runbook) related to the alert.",
				Optional:    true,
			},
			"query": schema.StringAttribute{
				Description: "The SQL query rendered from the builder rules. This is set by LibreNMS.",
				Computed:    true,
			},
			"severity": schema.StringAttribute{
				Description: "The severity of the alert rule [`ok`, `warning`, `critical`].",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("ok", "warning", "critical"),
				},
			},
		},
	}
}

func (r *alertRuleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data alertRuleModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// verify the builder is valid JSON
	var generic map[string]interface{}
	err := json.Unmarshal([]byte(data.Builder.ValueString()), &generic)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("builder"),
			"Invalid Builder Value",
			"The builder field must be a valid JSON string representing the alert rule logic. Unmarshal Error: "+err.Error(),
		)
		return
	}
}

// Configure sets the provider client for the resource.
func (r *alertRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *alertRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan alertRuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the alert rule using the LibreNMS client.
	payload := &librenms.AlertRuleCreateRequest{
		Builder:      plan.Builder.ValueString(),
		Count:        int(plan.MaxAlerts.ValueInt32()),
		Delay:        plan.Delay.ValueString(),
		Disabled:     librenms.Bool(plan.Disabled.ValueBool()),
		Interval:     plan.Interval.ValueString(),
		Mute:         plan.Mute.ValueBool(),
		Name:         plan.Name.ValueString(),
		Notes:        plan.Notes.ValueString(),
		ProcedureURL: plan.ProcedureURL.ValueString(),
		Severity:     plan.Severity.ValueString(),
		Devices:      make([]int, 0),
		Groups:       make([]int, 0),
		Locations:    make([]int, 0),
	}

	diags = plan.Devices.ElementsAs(ctx, &payload.Devices, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = plan.Groups.ElementsAs(ctx, &payload.Groups, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = plan.Locations.ElementsAs(ctx, &payload.Locations, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.CreateAlertRule(payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Alert Rule",
			fmt.Sprintf("Could not create alert rule: %s", err),
		)
		return
	}

	// have to get all the alert rules, so we can match by name to get computed values
	rulesResp, err := r.client.GetAlertRules()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Alert Rules",
			fmt.Sprintf("Could not get alert rules after creation: %s", err),
		)
		return
	}

	// Find the created alert rule by name
	var createdRule *librenms.AlertRule
	for _, rule := range rulesResp.Rules {
		if rule.Name == payload.Name {
			createdRule = &rule
			break
		}
	}
	if createdRule == nil {
		resp.Diagnostics.AddError(
			"Error Finding Created Alert Rule",
			fmt.Sprintf("Could not find the created alert rule with name %s in the response: %s", payload.Name, err),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.Int32Value(int32(createdRule.ID))
	plan.Extra = jsontypes.NewNormalizedValue(createdRule.Extra)
	plan.Query = types.StringValue(createdRule.Query)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *alertRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state alertRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed value from LibreNMS API
	alertResp, err := r.client.GetAlertRule(int(state.ID.ValueInt32()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Alert Rules",
			fmt.Sprintf("Could not read LibreNMS alert rule ID %d: %s", state.ID.ValueInt32(), err.Error()),
		)
		return
	}

	if alertResp == nil {
		resp.Diagnostics.AddError(
			"Error Reading Alert Rules",
			"Received nil response when creating alert rule. Please check the LibreNMS API.",
		)
		return
	}

	if len(alertResp.Rules) != 1 {
		resp.Diagnostics.AddError(
			"Unexpected Alert Rule Get Response",
			fmt.Sprintf("Expected one alert rule to be retrieved, got %d alert rules. Please check the LibreNMS API.", len(alertResp.Rules)),
		)
		return
	}

	// Overwrite items with refreshed state
	alertRule := alertResp.Rules[0]
	state.ID = types.Int32Value(int32(alertRule.ID))
	state.Builder = jsontypes.NewNormalizedValue(alertRule.Builder)
	state.Disabled = types.BoolValue(bool(alertRule.Disabled))
	state.Extra = jsontypes.NewNormalizedValue(alertRule.Extra)
	state.Name = types.StringValue(alertRule.Name)
	state.Query = types.StringValue(alertRule.Query)
	state.Severity = types.StringValue(alertRule.Severity)
	state.Devices = types.ListNull(types.Int32Type)
	state.Groups = types.ListNull(types.Int32Type)
	state.Locations = types.ListNull(types.Int32Type)

	// count, delay, interval, mute are all stored in extra as serialized JSON.
	// The delay and interval are represented in seconds, even though their input is in a format like `5m` or `1h`.
	// I've also seen inconsistent responses from the API where not all fields are returned in extra.
	//
	// So, these fields will not be updated in Read(). This will cause apply changes after an initial import, but
	// it shouldn't cause errors on apply.

	// check possibly null fields
	state.Notes = types.StringNull()
	if alertRule.Notes != nil {
		state.Notes = types.StringValue(*alertRule.Notes)
	}

	state.ProcedureURL = types.StringNull()
	if alertRule.ProcedureURL != nil {
		state.ProcedureURL = types.StringValue(*alertRule.ProcedureURL)
	}

	// Populate Devices list
	if len(alertRule.Devices) > 0 {
		deviceElements := make([]types.Int32, len(alertRule.Devices))
		for i, deviceID := range alertRule.Devices {
			deviceElements[i] = types.Int32Value(int32(deviceID))
		}

		var devicesDiags diag.Diagnostics
		state.Devices, devicesDiags = types.ListValueFrom(ctx, types.Int32Type, deviceElements)
		resp.Diagnostics.Append(devicesDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		state.Devices = types.ListNull(types.Int32Type)
	}

	// Populate Groups list
	if len(alertRule.Groups) > 0 {
		groupElements := make([]types.Int32, len(alertRule.Groups))
		for i, groupID := range alertRule.Groups {
			groupElements[i] = types.Int32Value(int32(groupID))
		}
		var groupsDiags diag.Diagnostics
		state.Groups, groupsDiags = types.ListValueFrom(ctx, types.Int32Type, groupElements)
		resp.Diagnostics.Append(groupsDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		state.Groups = types.ListNull(types.Int32Type)
	}

	// Populate Locations list
	if len(alertRule.Locations) > 0 {
		locationElements := make([]types.Int32, len(alertRule.Locations))
		for i, locationID := range alertRule.Locations {
			locationElements[i] = types.Int32Value(int32(locationID))
		}
		var locationsDiags diag.Diagnostics
		state.Locations, locationsDiags = types.ListValueFrom(ctx, types.Int32Type, locationElements)
		resp.Diagnostics.Append(locationsDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		state.Locations = types.ListNull(types.Int32Type)
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *alertRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan alertRuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the alert rule using the LibreNMS client.
	payload := &librenms.AlertRuleUpdateRequest{
		ID: int(plan.ID.ValueInt32()),
		AlertRuleCreateRequest: librenms.AlertRuleCreateRequest{
			Builder:      plan.Builder.ValueString(),
			Count:        int(plan.MaxAlerts.ValueInt32()),
			Delay:        plan.Delay.ValueString(),
			Disabled:     librenms.Bool(plan.Disabled.ValueBool()),
			Interval:     plan.Interval.ValueString(),
			Mute:         plan.Mute.ValueBool(),
			Name:         plan.Name.ValueString(),
			Notes:        plan.Notes.ValueString(),
			ProcedureURL: plan.ProcedureURL.ValueString(),
			Severity:     plan.Severity.ValueString(),
			Devices:      make([]int, 0),
			Groups:       make([]int, 0),
			Locations:    make([]int, 0),
		},
	}

	diags = plan.Devices.ElementsAs(ctx, &payload.Devices, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = plan.Groups.ElementsAs(ctx, &payload.Groups, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = plan.Locations.ElementsAs(ctx, &payload.Locations, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UpdateAlertRule(payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Alert Rule",
			fmt.Sprintf("Could not update alert rule: %s", err),
		)
		return
	}

	// have to get all the alert rules, so we can match by name to get computed values
	ruleResp, err := r.client.GetAlertRule(payload.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Alert Rule",
			fmt.Sprintf("Could not get alert rule after creation: %s", err),
		)
		return
	}

	if ruleResp == nil {
		resp.Diagnostics.AddError(
			"Error Reading Alert Rule",
			"Received nil response when getting alert rule. Please check the LibreNMS API.",
		)
		return
	}

	if len(ruleResp.Rules) != 1 {
		resp.Diagnostics.AddError(
			"Unexpected Alert Rule Get Response",
			fmt.Sprintf("Expected one alert rule to be retrieved, got %d alert rules. Please check the LibreNMS API.", len(ruleResp.Rules)),
		)
		return
	}

	// Map response body to schema and populate updated attribute values
	alertRule := ruleResp.Rules[0]
	plan.Extra = jsontypes.NewNormalizedValue(alertRule.Extra)
	plan.Query = types.StringValue(alertRule.Query)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *alertRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alertRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing group
	_, err := r.client.DeleteAlertRule(int(state.ID.ValueInt32()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting LibreNMS Alert Rule",
			"Could not delete alert rule, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *alertRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
