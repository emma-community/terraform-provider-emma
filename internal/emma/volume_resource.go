package emma

import (
	"context"
	"fmt"
	emmaSdk "github.com/emma-community/emma-go-sdk"
	"github.com/emma-community/terraform-provider-emma/internal/emma/common/async"
	"github.com/emma-community/terraform-provider-emma/internal/emma/common/convert"
	"github.com/emma-community/terraform-provider-emma/internal/emma/common/errors"
	"github.com/emma-community/terraform-provider-emma/internal/emma/common/retry"
	"github.com/emma-community/terraform-provider-emma/internal/emma/common/state"
	emma "github.com/emma-community/terraform-provider-emma/internal/emma/validation"
	"github.com/emma-community/terraform-provider-emma/tools"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"net/http"
)

var _ resource.Resource = &volumeResource{}

func NewVolumeResource() resource.Resource {
	return &volumeResource{}
}

// volumeResource defines the resource implementation.
type volumeResource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

// volumeResourceModel describes the resource data model.
type volumeResourceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	DataCenterId   types.String `tfsdk:"data_center_id"`
	VolumeGb       types.Int64  `tfsdk:"volume_gb"`
	VolumeType     types.String `tfsdk:"volume_type"`
	AttachedToId   types.Int64  `tfsdk:"attached_to_id"`
	IsSystem       types.Bool   `tfsdk:"is_system"`
	Status         types.String `tfsdk:"status"`
	ProjectId      types.Int64  `tfsdk:"project_id"`
	CloudProvider  types.Object `tfsdk:"cloud_provider"`
	Location       types.Object `tfsdk:"location"`
	DataCenter     types.Object `tfsdk:"data_center"`
	CreatedAt      types.String `tfsdk:"created_at"`
	CreatedByName  types.String `tfsdk:"created_by_name"`
	CreatedById    types.Int64  `tfsdk:"created_by_id"`
	ModifiedAt     types.String `tfsdk:"modified_at"`
	ModifiedByName types.String `tfsdk:"modified_by_name"`
	ModifiedById   types.Int64  `tfsdk:"modified_by_id"`
	Cost           types.Object `tfsdk:"cost"`
}

type volumeResourceProviderModel struct {
	Id   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type volumeResourceLocationModel struct {
	Id        types.Int64   `tfsdk:"id"`
	Name      types.String  `tfsdk:"name"`
	Continent types.String  `tfsdk:"continent"`
	Region    types.String  `tfsdk:"region"`
	Latitude  types.Float64 `tfsdk:"latitude"`
	Longitude types.Float64 `tfsdk:"longitude"`
}

type volumeResourceDataCenterModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ProviderId   types.Int64  `tfsdk:"provider_id"`
	ProviderName types.String `tfsdk:"provider_name"`
	LocationId   types.Int64  `tfsdk:"location_id"`
	LocationName types.String `tfsdk:"location_name"`
}

type volumeResourceCostModel struct {
	Unit     types.String  `tfsdk:"unit"`
	Currency types.String  `tfsdk:"currency"`
	Price    types.Float64 `tfsdk:"price"`
}

func (r *volumeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (r *volumeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource creates and manages storage volumes in the Emma platform.\n\n" +
			"Volumes are block storage devices that can be attached to compute instances (VMs) for persistent data storage. " +
			"To create a volume, you need to specify the data center, size in gigabytes, and volume type (e.g., ssd, hdd).\n\n" +
			"Volumes can be created independently or attached to a compute instance during creation. " +
			"You can also resize volumes (increase size only) and change attachments after creation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "ID of the volume",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "Name of the volume (assigned by the platform)",
				Computed:    true,
			},
			"data_center_id": schema.StringAttribute{
				Description:   "Data center ID where the volume will be created, volume will be recreated after changing this value",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{emma.ValidDataCenterId{}},
			},
			"volume_gb": schema.Int64Attribute{
				Description: "Volume size in gigabytes, can only be increased",
				Required:    true,
				Validators:  []validator.Int64{emma.MinimumVolumeSize{}, emma.PositiveInt64{}},
			},
			"volume_type": schema.StringAttribute{
				Description:   "Volume type (e.g., ssd, hdd), volume will be recreated after changing this value",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{emma.NonEmptyVolumeType{}},
			},
			"attached_to_id": schema.Int64Attribute{
				Description: "ID of the compute instance to attach the volume to",
				Optional:    true,
				Validators:  []validator.Int64{emma.PositiveInt64{}},
			},
			"is_system": schema.BoolAttribute{
				Description: "Indicates whether the volume contains the operating system",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Current status of the volume",
				Computed:    true,
			},
			"project_id": schema.Int64Attribute{
				Description: "Project ID owning the volume",
				Computed:    true,
			},
			"cloud_provider": schema.SingleNestedAttribute{
				Description: "Cloud provider information",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.Int64Attribute{
						Description: "Provider ID",
						Computed:    true,
					},
					"name": schema.StringAttribute{
						Description: "Provider name",
						Computed:    true,
					},
				},
			},
			"location": schema.SingleNestedAttribute{
				Description: "Geographic location information",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.Int64Attribute{
						Description: "Location ID",
						Computed:    true,
					},
					"name": schema.StringAttribute{
						Description: "Location name",
						Computed:    true,
					},
					"continent": schema.StringAttribute{
						Description: "Continent",
						Computed:    true,
					},
					"region": schema.StringAttribute{
						Description: "Region",
						Computed:    true,
					},
					"latitude": schema.Float64Attribute{
						Description: "Approximate latitude of the geographical location",
						Computed:    true,
					},
					"longitude": schema.Float64Attribute{
						Description: "Approximate longitude of the geographical location",
						Computed:    true,
					},
				},
			},
			"data_center": schema.SingleNestedAttribute{
				Description: "Data center details",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "Data center ID",
						Computed:    true,
					},
					"name": schema.StringAttribute{
						Description: "Data center name",
						Computed:    true,
					},
					"provider_id": schema.Int64Attribute{
						Description: "ID of the cloud provider that owns the data center",
						Computed:    true,
					},
					"provider_name": schema.StringAttribute{
						Description: "Name of the cloud provider that owns the data center",
						Computed:    true,
					},
					"location_id": schema.Int64Attribute{
						Description: "ID of the data center location",
						Computed:    true,
					},
					"location_name": schema.StringAttribute{
						Description: "Name of the data center location",
						Computed:    true,
					},
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp",
				Computed:    true,
			},
			"created_by_name": schema.StringAttribute{
				Description: "Name of the user who created the volume",
				Computed:    true,
			},
			"created_by_id": schema.Int64Attribute{
				Description: "ID of the user who created the volume",
				Computed:    true,
			},
			"modified_at": schema.StringAttribute{
				Description: "Date and time when the volume was last modified",
				Computed:    true,
			},
			"modified_by_name": schema.StringAttribute{
				Description: "Name of the user who last modified the volume",
				Computed:    true,
			},
			"modified_by_id": schema.Int64Attribute{
				Description: "ID of the user who last modified the volume",
				Computed:    true,
			},
			"cost": schema.SingleNestedAttribute{
				Description: "Cost information for the volume",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"unit": schema.StringAttribute{
						Description: "Cost period unit",
						Computed:    true,
					},
					"currency": schema.StringAttribute{
						Description: "Currency of cost",
						Computed:    true,
					},
					"price": schema.Float64Attribute{
						Description: "Cost of the volume for the period",
						Computed:    true,
					},
				},
			},
		},
	}
}

func (r *volumeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData))
		return
	}
	r.apiClient = client.apiClient
	r.token = client.token
}

func (r *volumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data volumeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build VolumeCreate request from resource model using helper function
	volumeCreateRequest := convertResourceToVolumeCreateRequest(&data)

	// Call Emma API to create volume
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	volume, response, err := r.apiClient.VolumesAPI.VolumeCreate(auth).VolumeCreate(*volumeCreateRequest).Execute()

	if err != nil {
		statusCode := 0
		apiError := ""
		if response != nil {
			statusCode = response.StatusCode
			apiError = tools.ExtractErrorMessage(response)
		}
		
		resourceErr := errors.NewError("emma_volume", "Create").
			WithStatusCode(statusCode).
			WithAPIError(apiError).
			WithMessage(errors.MapHTTPError(statusCode, apiError)).
			Build()
		
		resp.Diagnostics.AddError("Client Error", resourceErr.Error())
		return
	}

	// Convert API response to resource model
	convertVolumeResponseToResource(ctx, &data, volume, resp.Diagnostics)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *volumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data volumeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Extract volume ID from state
	volumeId, err := convert.StringToInt32(data.Id)
	if err != nil {
		resourceErr := errors.NewError("emma_volume", "Read").
			WithID(data.Id.ValueString()).
			WithMessage(fmt.Sprintf("Invalid volume ID: %v", err)).
			Build()
		
		resp.Diagnostics.AddError("Validation Error", resourceErr.Error())
		return
	}

	// Call Emma API to get volume
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	volume, response, err := r.apiClient.VolumesAPI.GetVolume(auth, volumeId).Execute()

	if err != nil {
		// Handle 404 errors by removing from state using StateManager
		if response != nil && response.StatusCode == 404 {
			stateManager := state.NewStateManager(ctx)
			stateManager.RemoveFromState(resp)
			return
		}

		statusCode := 0
		apiError := ""
		if response != nil {
			statusCode = response.StatusCode
			apiError = tools.ExtractErrorMessage(response)
		}
		
		resourceErr := errors.NewError("emma_volume", "Read").
			WithID(data.Id.ValueString()).
			WithStatusCode(statusCode).
			WithAPIError(apiError).
			WithMessage(errors.MapHTTPError(statusCode, apiError)).
			Build()
		
		resp.Diagnostics.AddError("Client Error", resourceErr.Error())
		return
	}

	// Update all computed attributes
	convertVolumeResponseToResource(ctx, &data, volume, resp.Diagnostics)

	// Save updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// waitForStableAndRetry waits for a resource to reach a stable state, then executes an API operation with retry.
func (r *volumeResource) waitForStableAndRetry(ctx context.Context, auth context.Context, config state.StateTransitionConfig, operation func() (*http.Response, error), operationName string, volumeIdStr string) error {
	if err := state.NewStateTransitionManager(config).WaitForStableState(auth); err != nil {
		return fmt.Errorf("resource did not reach stable state before %s: %w", operationName, err)
	}

	retryConfig := retry.StateConflictRetryConfig()
	var lastResponse *http.Response
	var lastAPIError string

	err := retry.Retry(auth, retryConfig, func() error {
		response, err := operation()
		lastResponse = response
		if err != nil {
			lastAPIError = tools.ExtractErrorMessage(response)
			statusCode := 0
			if response != nil {
				statusCode = response.StatusCode
			}
			if retry.IsStateConflictError(err, statusCode, lastAPIError) {
				return err
			}
			return fmt.Errorf("non-retryable error: %w", err)
		}
		return nil
	})

	if err != nil {
		statusCode := 0
		apiError := ""
		if lastResponse != nil {
			statusCode = lastResponse.StatusCode
			apiError = lastAPIError
		}
		return errors.NewError("emma_volume", "Update").
			WithID(volumeIdStr).
			WithStatusCode(statusCode).
			WithAPIError(apiError).
			WithMessage(fmt.Sprintf("%s failed: %s", operationName, errors.MapHTTPError(statusCode, apiError))).
			Build()
	}
	return nil
}

func (r *volumeResource) volumeStateConfig(auth context.Context, volumeId int32, volumeIdStr string) state.StateTransitionConfig {
	return state.StateTransitionConfig{
		ResourceType: "volume",
		ResourceID:   volumeIdStr,
		StatusChecker: func(ctx context.Context) (string, error) {
			vol, _, err := r.apiClient.VolumesAPI.GetVolume(auth, volumeId).Execute()
			if err != nil {
				return "", err
			}
			if vol.Status == nil {
				return "", fmt.Errorf("volume status is nil")
			}
			return *vol.Status, nil
		},
		TargetStates:       state.VolumeStableStates,
		TransitionalStates: state.VolumeTransitionalStates,
		FailureStates:      state.VolumeFailureStates,
		Timeout:            async.DefaultTimeout,
		PollInterval:       async.DefaultPollInterval,
	}
}

func (r *volumeResource) vmStateConfig(auth context.Context, vmId int32) state.StateTransitionConfig {
	return state.StateTransitionConfig{
		ResourceType: "vm",
		ResourceID:   fmt.Sprintf("%d", vmId),
		StatusChecker: func(ctx context.Context) (string, error) {
			vm, _, err := r.apiClient.VirtualMachinesAPI.GetVm(auth, vmId).Execute()
			if err != nil {
				return "", err
			}
			if vm.Status == nil {
				return "", fmt.Errorf("VM status is nil")
			}
			return *vm.Status, nil
		},
		TargetStates:       state.VMStableStates,
		TransitionalStates: state.VMTransitionalStates,
		FailureStates:      state.VMFailureStates,
		Timeout:            async.DefaultTimeout,
		PollInterval:       async.DefaultPollInterval,
	}
}

func (r *volumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan volumeResourceModel
	var stateData volumeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	volumeId, err := convert.StringToInt32(stateData.Id)
	if err != nil {
		resp.Diagnostics.AddError("Validation Error",
			errors.NewError("emma_volume", "Update").WithID(stateData.Id.ValueString()).
				WithMessage(fmt.Sprintf("Invalid volume ID: %v", err)).Build().Error())
		return
	}
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	volumeIdStr := stateData.Id.ValueString()

	planAttachedToId := plan.AttachedToId
	stateAttachedToId := stateData.AttachedToId

	// Handle volume resize
	if !plan.VolumeGb.Equal(stateData.VolumeGb) {
		if plan.VolumeGb.ValueInt64() < stateData.VolumeGb.ValueInt64() {
			resp.Diagnostics.AddError("Validation Error",
				fmt.Sprintf("Volume size can only be increased. Current: %d GB, requested: %d GB",
					stateData.VolumeGb.ValueInt64(), plan.VolumeGb.ValueInt64()))
			return
		}

		if err := r.waitForStableAndRetry(ctx, auth,
			r.volumeStateConfig(auth, volumeId, volumeIdStr),
			func() (*http.Response, error) {
				volumeEdit := convertResourceToVolumeEditRequest(&plan)
				_, response, err := r.apiClient.VolumesAPI.VolumeActions(auth, volumeId).VolumeEdit(*volumeEdit).Execute()
				return response, err
			}, "resize", volumeIdStr); err != nil {
			resp.Diagnostics.AddError("Client Error", err.Error())
			return
		}
	}

	// Handle attachment changes
	if !planAttachedToId.Equal(stateAttachedToId) {
		// Detach from old instance
		if !stateAttachedToId.IsNull() && !stateAttachedToId.IsUnknown() {
			oldVmId, err := convert.Int64ToInt32(stateAttachedToId)
			if err != nil {
				resp.Diagnostics.AddError("Validation Error",
					fmt.Sprintf("Invalid VM ID for detachment: %v", err))
				return
			}

			if err := state.NewStateTransitionManager(r.vmStateConfig(auth, oldVmId)).WaitForStableState(auth); err != nil {
				resp.Diagnostics.AddError("State Transition Error", err.Error())
				return
			}
			if err := r.waitForStableAndRetry(ctx, auth,
				r.volumeStateConfig(auth, volumeId, volumeIdStr),
				func() (*http.Response, error) {
					volumeDetach := emmaSdk.NewVolumeDetach("detach", volumeId)
					vmActionsReq := emmaSdk.VolumeDetachAsVmActionsRequest(volumeDetach)
					_, response, err := r.apiClient.VirtualMachinesAPI.VmActions(auth, oldVmId).VmActionsRequest(vmActionsReq).Execute()
					return response, err
				}, "detach", volumeIdStr); err != nil {
				resp.Diagnostics.AddError("Client Error", err.Error())
				return
			}
		}

		// Attach to new instance
		if !planAttachedToId.IsNull() && !planAttachedToId.IsUnknown() {
			newVmId, err := convert.Int64ToInt32(planAttachedToId)
			if err != nil {
				resp.Diagnostics.AddError("Validation Error",
					fmt.Sprintf("Invalid VM ID for attachment: %v", err))
				return
			}

			if err := state.NewStateTransitionManager(r.vmStateConfig(auth, newVmId)).WaitForStableState(auth); err != nil {
				resp.Diagnostics.AddError("State Transition Error", err.Error())
				return
			}
			if err := r.waitForStableAndRetry(ctx, auth,
				r.volumeStateConfig(auth, volumeId, volumeIdStr),
				func() (*http.Response, error) {
					volumeAttach := emmaSdk.NewVolumeAttach("attach", volumeId)
					vmActionsReq := emmaSdk.VolumeAttachAsVmActionsRequest(volumeAttach)
					_, response, err := r.apiClient.VirtualMachinesAPI.VmActions(auth, newVmId).VmActionsRequest(vmActionsReq).Execute()
					return response, err
				}, "attach", volumeIdStr); err != nil {
				resp.Diagnostics.AddError("Client Error", err.Error())
				return
			}
		}
	}

	// Refresh volume state after updates
	volume, response, err := r.apiClient.VolumesAPI.GetVolume(auth, volumeId).Execute()
	if err != nil {
		statusCode := 0
		apiError := ""
		if response != nil {
			statusCode = response.StatusCode
			apiError = tools.ExtractErrorMessage(response)
		}
		resp.Diagnostics.AddError("Client Error",
			errors.NewError("emma_volume", "Update").WithID(volumeIdStr).
				WithStatusCode(statusCode).WithAPIError(apiError).
				WithMessage(errors.MapHTTPError(statusCode, apiError)).Build().Error())
		return
	}

	convertVolumeResponseToResource(ctx, &plan, volume, resp.Diagnostics)

	if !planAttachedToId.Equal(stateAttachedToId) {
		plan.AttachedToId = planAttachedToId
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *volumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data volumeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Extract volume ID from state
	volumeId, err := convert.StringToInt32(data.Id)
	if err != nil {
		resourceErr := errors.NewError("emma_volume", "Delete").
			WithID(data.Id.ValueString()).
			WithMessage(fmt.Sprintf("Invalid volume ID: %v", err)).
			Build()
		
		resp.Diagnostics.AddError("Validation Error", resourceErr.Error())
		return
	}
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)

	// Check if volume is a system volume
	if !data.IsSystem.IsNull() && !data.IsSystem.IsUnknown() && data.IsSystem.ValueBool() {
		resourceErr := errors.NewError("emma_volume", "Delete").
			WithID(data.Id.ValueString()).
			WithMessage("Cannot delete system volume. System volumes contain the operating system and cannot be deleted.").
			Build()
		
		resp.Diagnostics.AddError("Validation Error", resourceErr.Error())
		return
	}

	if !data.AttachedToId.IsNull() && !data.AttachedToId.IsUnknown() {
		vmId, err := convert.Int64ToInt32(data.AttachedToId)
		if err != nil {
			resourceErr := errors.NewError("emma_volume", "Delete").
				WithID(data.Id.ValueString()).
				WithMessage(fmt.Sprintf("Invalid VM ID for detachment: %v", err)).
				Build()

			resp.Diagnostics.AddError("Validation Error", resourceErr.Error())
			return
		}

		_, vmResponse, vmErr := r.apiClient.VirtualMachinesAPI.GetVm(auth, vmId).Execute()
		if vmErr != nil && vmResponse != nil && vmResponse.StatusCode == 404 {
			tflog.Info(ctx, "VM no longer exists, proceeding to delete volume", map[string]interface{}{
				"vm_id":     vmId,
				"volume_id": data.Id.ValueString(),
			})
		} else {
			tflog.Info(ctx, "Skipping explicit detach during destroy — API handles it", map[string]interface{}{
				"vm_id":     vmId,
				"volume_id": data.Id.ValueString(),
			})
		}
	}

	// Call Emma API to delete volume
	_, response, err := r.apiClient.VolumesAPI.VolumeDelete(auth, volumeId).Execute()

	if err != nil {
		// Handle 404 errors as successful deletion (idempotent)
		if response != nil && response.StatusCode == 404 {
			// Volume already deleted, treat as success
			return
		}

		statusCode := 0
		apiError := ""
		if response != nil {
			statusCode = response.StatusCode
			apiError = tools.ExtractErrorMessage(response)
		}
		
		resourceErr := errors.NewError("emma_volume", "Delete").
			WithID(data.Id.ValueString()).
			WithStatusCode(statusCode).
			WithAPIError(apiError).
			WithMessage(errors.MapHTTPError(statusCode, apiError)).
			Build()
		
		resp.Diagnostics.AddError("Client Error", resourceErr.Error())
		return
	}

	// Resource is automatically removed from state after successful Delete
}

func (r *volumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Use ImportStatePassthroughID to set the ID from the import string
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Call Read operation to populate the rest of the state
	r.Read(ctx, resource.ReadRequest{State: resp.State, Private: resp.Private},
		&resource.ReadResponse{State: resp.State, Private: resp.Private, Diagnostics: resp.Diagnostics})
}

// Helper function to get attribute types for provider nested object
func (o volumeResourceProviderModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":   types.Int64Type,
		"name": types.StringType,
	}
}

// Helper function to get attribute types for location nested object
func (o volumeResourceLocationModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":        types.Int64Type,
		"name":      types.StringType,
		"continent": types.StringType,
		"region":    types.StringType,
		"latitude":  types.Float64Type,
		"longitude": types.Float64Type,
	}
}

// Helper function to get attribute types for data center nested object
func (o volumeResourceDataCenterModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":            types.StringType,
		"name":          types.StringType,
		"provider_id":   types.Int64Type,
		"provider_name": types.StringType,
		"location_id":   types.Int64Type,
		"location_name": types.StringType,
	}
}

// Helper function to get attribute types for cost nested object
func (o volumeResourceCostModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"unit":     types.StringType,
		"currency": types.StringType,
		"price":    types.Float64Type,
	}
}

// convertVolumeResponseToResource converts Emma API Volume response to Terraform resource model
func convertVolumeResponseToResource(ctx context.Context, data *volumeResourceModel, volume *emmaSdk.Volume, diags diag.Diagnostics) {
	// Set ID using shared utility
	data.Id = convert.Int32ToString(volume.Id)

	// Set basic attributes using shared utilities
	data.Name = convert.StringPointerToString(volume.Name)
	data.VolumeGb = convert.Int32ToInt64(volume.SizeGb)
	data.VolumeType = convert.StringPointerToString(volume.Type)
	data.IsSystem = convert.BoolPointerToBool(volume.IsSystem)
	data.Status = convert.StringPointerToString(volume.Status)
	data.ProjectId = convert.Int32ToInt64(volume.ProjectId)
	data.AttachedToId = convert.Int32ToInt64(volume.AttachedToId)
	data.CreatedAt = convert.StringPointerToString(volume.CreatedAt)

	// Convert provider nested object
	if volume.Provider != nil {
		providerModel := volumeResourceProviderModel{
			Id:   convert.Int32ToInt64(volume.Provider.Id),
			Name: convert.StringPointerToString(volume.Provider.Name),
		}
		providerObj, providerDiag := types.ObjectValueFrom(ctx, providerModel.attrTypes(), providerModel)
		data.CloudProvider = providerObj
		diags.Append(providerDiag...)
	} else {
		data.CloudProvider = types.ObjectNull(volumeResourceProviderModel{}.attrTypes())
	}

	// Convert location nested object
	if volume.Location != nil {
		locationModel := volumeResourceLocationModel{
			Id:        convert.Int32ToInt64(volume.Location.Id),
			Name:      convert.StringPointerToString(volume.Location.Name),
			Continent: convert.StringPointerToString(volume.Location.Continent),
			Region:    convert.StringPointerToString(volume.Location.Region),
		}
		if volume.Location.Latitude != nil {
			locationModel.Latitude = types.Float64Value(*volume.Location.Latitude)
		} else {
			locationModel.Latitude = types.Float64Null()
		}
		if volume.Location.Longitude != nil {
			locationModel.Longitude = types.Float64Value(*volume.Location.Longitude)
		} else {
			locationModel.Longitude = types.Float64Null()
		}
		locationObj, locationDiag := types.ObjectValueFrom(ctx, locationModel.attrTypes(), locationModel)
		data.Location = locationObj
		diags.Append(locationDiag...)
	} else {
		data.Location = types.ObjectNull(volumeResourceLocationModel{}.attrTypes())
	}

	// Convert data center nested object
	if volume.DataCenter != nil {
		dataCenterModel := volumeResourceDataCenterModel{
			Id:   convert.StringPointerToString(volume.DataCenter.Id),
			Name: convert.StringPointerToString(volume.DataCenter.Name),
		}
		if volume.DataCenter.ProviderId != nil {
			dataCenterModel.ProviderId = types.Int64Value(int64(*volume.DataCenter.ProviderId))
		} else {
			dataCenterModel.ProviderId = types.Int64Null()
		}
		dataCenterModel.ProviderName = convert.StringPointerToString(volume.DataCenter.ProviderName)
		if volume.DataCenter.LocationId != nil {
			dataCenterModel.LocationId = types.Int64Value(int64(*volume.DataCenter.LocationId))
		} else {
			dataCenterModel.LocationId = types.Int64Null()
		}
		dataCenterModel.LocationName = convert.StringPointerToString(volume.DataCenter.LocationName)
		dataCenterObj, dataCenterDiag := types.ObjectValueFrom(ctx, dataCenterModel.attrTypes(), dataCenterModel)
		data.DataCenter = dataCenterObj
		diags.Append(dataCenterDiag...)
	} else {
		data.DataCenter = types.ObjectNull(volumeResourceDataCenterModel{}.attrTypes())
	}

	// Set data_center_id from the data center object (for consistency with API response)
	if volume.DataCenter != nil && volume.DataCenter.Id != nil {
		data.DataCenterId = types.StringValue(*volume.DataCenter.Id)
	}

	// Set created_by / modified_by fields
	data.CreatedByName = convert.StringPointerToString(volume.CreatedByName)
	if volume.CreatedById != nil {
		data.CreatedById = types.Int64Value(int64(*volume.CreatedById))
	} else {
		data.CreatedById = types.Int64Null()
	}
	data.ModifiedAt = convert.StringPointerToString(volume.ModifiedAt)
	data.ModifiedByName = convert.StringPointerToString(volume.ModifiedByName)
	if volume.ModifiedById != nil {
		data.ModifiedById = types.Int64Value(int64(*volume.ModifiedById))
	} else {
		data.ModifiedById = types.Int64Null()
	}

	// Convert cost nested object
	if volume.Cost != nil {
		costModel := volumeResourceCostModel{
			Unit:     convert.StringPointerToString(volume.Cost.Unit),
			Currency: convert.StringPointerToString(volume.Cost.Currency),
		}
		if volume.Cost.Price != nil {
			costModel.Price = types.Float64Value(float64(*volume.Cost.Price))
		} else {
			costModel.Price = types.Float64Null()
		}
		costObj, costDiag := types.ObjectValueFrom(ctx, costModel.attrTypes(), costModel)
		data.Cost = costObj
		diags.Append(costDiag...)
	} else {
		data.Cost = types.ObjectNull(volumeResourceCostModel{}.attrTypes())
	}
}

// convertResourceToVolumeCreateRequest converts Terraform resource model to SDK VolumeCreate request
func convertResourceToVolumeCreateRequest(data *volumeResourceModel) *emmaSdk.VolumeCreate {
	// Create VolumeCreate request with required fields
	volumeCreateRequest := emmaSdk.NewVolumeCreate(
		data.DataCenterId.ValueString(),
		int32(data.VolumeGb.ValueInt64()),
		data.VolumeType.ValueString(),
	)

	// Add optional attached_to_id if provided
	if !data.AttachedToId.IsNull() && !data.AttachedToId.IsUnknown() {
		attachedToId, err := convert.Int64ToInt32(data.AttachedToId)
		if err == nil {
			volumeCreateRequest.SetAttachedToId(attachedToId)
		}
	}

	return volumeCreateRequest
}

// convertResourceToVolumeEditRequest converts Terraform resource model to SDK VolumeEdit request
func convertResourceToVolumeEditRequest(data *volumeResourceModel) *emmaSdk.VolumeEdit {
	// Create VolumeEdit request with action and new size
	volumeEdit := emmaSdk.NewVolumeEdit(
		"edit",
		int32(data.VolumeGb.ValueInt64()),
	)

	return volumeEdit
}
