package emma

import (
	"context"
	"fmt"
	emmaSdk "github.com/emma-community/emma-go-sdk"
	"github.com/emma-community/terraform-provider-emma/tools"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strconv"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &sshKeyResource{}

func NewSshKeyResource() resource.Resource {
	return &sshKeyResource{}
}

// sshKeyResource defines the resource implementation.
type sshKeyResource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

// sshKeyResourceModel describes the resource data model.
type sshKeyResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Key         types.String `tfsdk:"key"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	KeyType     types.String `tfsdk:"key_type"`
	PrivateKey  types.String `tfsdk:"private_key"`
}

func (r *sshKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (r *sshKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SshKey resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "SshKey id configurable attribute",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "SshKey name configurable attribute",
				Computed:            false,
				Required:            true,
				Optional:            false,
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "SshKey key configurable attribute",
				Computed:            true,
				Required:            false,
				Optional:            true,
			},
			"fingerprint": schema.StringAttribute{
				MarkdownDescription: "SshKey fingerprint configurable attribute",
				Computed:            true,
			},
			"key_type": schema.StringAttribute{
				MarkdownDescription: "SshKey key_type configurable attribute",
				Computed:            true,
				Required:            false,
				Optional:            true,
			},
			"private_key": schema.StringAttribute{
				MarkdownDescription: "SshKey private_key configurable attribute",
				Computed:            true,
			},
		},
	}
}

func (r *sshKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData))
		return
	}
	r.apiClient = client.apiClient
	r.token = client.token
}

func (r *sshKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data sshKeyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if !data.Key.IsUnknown() && !data.KeyType.IsUnknown() {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create ssh key: contradicting fields: key_type, key"))
	} else if data.Key.IsUnknown() && data.KeyType.IsUnknown() {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create ssh key: key or key_type is required"))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	var sshKeyCreateImportRequest emmaSdk.SshKeysCreateImportRequest
	ConvertToSshKeyCreateImportRequest(data, &sshKeyCreateImportRequest)
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	sshKey, response, err := r.apiClient.SSHKeysAPI.SshKeysCreateImport(auth).SshKeysCreateImportRequest(sshKeyCreateImportRequest).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create ssh key, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	ConvertSshKey201ResponseToResource(&data, sshKey)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a ssh key resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sshKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data sshKeyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	sshKey, response, err := r.apiClient.SSHKeysAPI.GetSshKey(auth, tools.StringToInt32(data.Id.ValueString())).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read ssh key, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	ConvertSshKeyResponseToResource(&data, sshKey)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sshKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planData sshKeyResourceModel
	var stateData sshKeyResourceModel

	// Read Terraform plan planData into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)

	if !planData.Key.IsUnknown() && !planData.KeyType.IsUnknown() {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to update ssh key: contradicting fields: key_type, key"))
	} else if planData.Key.IsUnknown() && planData.KeyType.IsUnknown() {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to update ssh key: key or key_type is required"))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// auth context for all api calls
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)

	if !planData.Key.Equal(stateData.Key) || !planData.KeyType.Equal(stateData.KeyType) {
		var sshKeyCreateImportRequest emmaSdk.SshKeysCreateImportRequest
		ConvertToSshKeyCreateImportRequest(planData, &sshKeyCreateImportRequest)
		auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
		sshKey, response, err := r.apiClient.SSHKeysAPI.SshKeysCreateImport(auth).SshKeysCreateImportRequest(sshKeyCreateImportRequest).Execute()

		if err != nil {
			resp.Diagnostics.AddError("Client Error",
				fmt.Sprintf("Unable to create ssh key, got error: %s",
					tools.ExtractErrorMessage(response)))
			return
		}

		Delete(auth, r, stateData, resp.Diagnostics)

		ConvertSshKey201ResponseToResource(&stateData, sshKey)

	} else {
		var sshKeyUpdateRequest emmaSdk.SshKeyUpdate
		ConvertToSshKeyUpdateRequest(planData, &sshKeyUpdateRequest)
		sshKey, response, err := r.apiClient.SSHKeysAPI.SshKeyUpdate(auth, tools.StringToInt32(stateData.Id.ValueString())).SshKeyUpdate(sshKeyUpdateRequest).Execute()

		if err != nil {
			resp.Diagnostics.AddError("Client Error",
				fmt.Sprintf("Unable to update ssh key, got error: %s",
					tools.ExtractErrorMessage(response)))
			return
		}

		ConvertSshKeyResponseToResource(&stateData, sshKey)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "updated a ssh key resource")

	// Save planData into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &stateData)...)
}

func (r *sshKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data sshKeyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	Delete(auth, r, data, resp.Diagnostics)
}

func Delete(ctx context.Context, r *sshKeyResource, stateData sshKeyResourceModel, diag diag.Diagnostics) {
	response, err := r.apiClient.SSHKeysAPI.SshKeyDelete(ctx, tools.StringToInt32(stateData.Id.ValueString())).Execute()

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	if err != nil {
		diag.AddError("Client Error",
			fmt.Sprintf("Unable to delete ssh key, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}
}

func (r *sshKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	r.Read(ctx, resource.ReadRequest{State: resp.State, Private: resp.Private},
		&resource.ReadResponse{State: resp.State, Private: resp.Private, Diagnostics: resp.Diagnostics})
}

func ConvertToSshKeyCreateImportRequest(data sshKeyResourceModel, sshKeyCreate *emmaSdk.SshKeysCreateImportRequest) {
	if !data.KeyType.IsUnknown() {
		sshKeyCreateRequest := emmaSdk.SshKeyCreate{}
		sshKeyCreateRequest.Name = data.Name.ValueString()
		sshKeyCreateRequest.KeyType = data.KeyType.ValueString()
		sshKeyCreate.SshKeyCreate = &sshKeyCreateRequest
	} else if !data.Key.IsUnknown() {
		sshKeyImportRequest := emmaSdk.SshKeyImport{}
		sshKeyImportRequest.Name = data.Name.ValueString()
		sshKeyImportRequest.Key = data.Key.ValueString()
		sshKeyCreate.SshKeyImport = &sshKeyImportRequest
	}
}

func ConvertToSshKeyUpdateRequest(data sshKeyResourceModel, sshKeyUpdate *emmaSdk.SshKeyUpdate) {
	sshKeyUpdate.Name = data.Name.ValueString()
}

func ConvertSshKey201ResponseToResource(data *sshKeyResourceModel, sshKeyResponse *emmaSdk.SshKeysCreateImport201Response) {
	if sshKeyResponse.SshKey != nil {
		ConvertSshKeyResponseToResource(data, sshKeyResponse.SshKey)
	} else if sshKeyResponse.SshKeyGenerated != nil {
		data.Id = types.StringValue(strconv.Itoa(int(*sshKeyResponse.SshKeyGenerated.Id)))
		data.Name = types.StringValue(*sshKeyResponse.SshKeyGenerated.Name)
		data.Key = types.StringValue(*sshKeyResponse.SshKeyGenerated.Key)
		data.Fingerprint = types.StringValue(*sshKeyResponse.SshKeyGenerated.Fingerprint)
		data.KeyType = types.StringValue(*sshKeyResponse.SshKeyGenerated.KeyType)
		if sshKeyResponse.SshKeyGenerated.PrivateKey != nil {
			data.PrivateKey = types.StringValue(*sshKeyResponse.SshKeyGenerated.PrivateKey)
		} else if !data.PrivateKey.IsUnknown() && !data.PrivateKey.IsNull() {
			//ignore
		} else {
			data.PrivateKey = types.StringValue("")
		}
	}
}

func ConvertSshKeyResponseToResource(data *sshKeyResourceModel, sshKeyResponse *emmaSdk.SshKey) {
	data.Id = types.StringValue(strconv.Itoa(int(*sshKeyResponse.Id)))
	data.Name = types.StringValue(*sshKeyResponse.Name)
	data.Key = types.StringValue(*sshKeyResponse.Key)
	data.Fingerprint = types.StringValue(*sshKeyResponse.Fingerprint)
	data.KeyType = types.StringValue(*sshKeyResponse.KeyType)
	if data.PrivateKey.IsNull() || data.PrivateKey.IsUnknown() {
		data.PrivateKey = types.StringValue("")
	}
}