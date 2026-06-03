package emma

import (
	"context"
	"fmt"
	emmaSdk "github.com/emma-community/emma-go-sdk"
	"github.com/emma-community/terraform-provider-emma/tools"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &sshKeysDataSource{}

func NewSshKeysDataSource() datasource.DataSource {
	return &sshKeysDataSource{}
}

// sshKeysDataSource defines the data source implementation.
type sshKeysDataSource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

// sshKeysDataSourceModel describes the data source data model.
type sshKeysDataSourceModel struct {
	SshKeys types.List `tfsdk:"ssh_keys"`
}

// sshKeyItemModel describes individual SSH key items in the list.
type sshKeyItemModel struct {
	Id            types.Int64  `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Key           types.String `tfsdk:"key"`
	KeyType       types.String `tfsdk:"key_type"`
	Fingerprint   types.String `tfsdk:"fingerprint"`
	CreatedAt     types.String `tfsdk:"created_at"`
	CreatedByName types.String `tfsdk:"created_by_name"`
	CreatedById   types.Int64  `tfsdk:"created_by_id"`
}

func (d *sshKeysDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_keys"
}

func (d *sshKeysDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a list of all SSH keys in the project.",
		Attributes: map[string]schema.Attribute{
			"ssh_keys": schema.ListNestedAttribute{
				Description: "List of SSH keys",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "SSH key ID",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "SSH key name",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "SSH public key content",
							Computed:    true,
						},
						"key_type": schema.StringAttribute{
							Description: "SSH key type (RSA or ED25519)",
							Computed:    true,
						},
						"fingerprint": schema.StringAttribute{
							Description: "SSH key fingerprint",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Date and time when the SSH key was created",
							Computed:    true,
						},
						"created_by_name": schema.StringAttribute{
							Description: "Name of the user who created the SSH key",
							Computed:    true,
						},
						"created_by_id": schema.Int64Attribute{
							Description: "ID of the user who created the SSH key",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *sshKeysDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.apiClient = client.apiClient
	d.token = client.token
}

func (d *sshKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data sshKeysDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *d.token.AccessToken)
	sshKeys, response, err := d.apiClient.SSHKeysAPI.SshKeys(auth).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read SSH keys, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	keyList, diags := convertSshKeysToList(ctx, sshKeys)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.SshKeys = keyList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Helper function to get attribute types for SSH key nested object
func (o sshKeyItemModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":              types.Int64Type,
		"name":            types.StringType,
		"key":             types.StringType,
		"key_type":        types.StringType,
		"fingerprint":     types.StringType,
		"created_at":      types.StringType,
		"created_by_name": types.StringType,
		"created_by_id":   types.Int64Type,
	}
}

// convertSshKeysToList converts Emma API SSH keys response to Terraform list
func convertSshKeysToList(ctx context.Context, sshKeys []emmaSdk.SshKey) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	var itemModels []sshKeyItemModel

	for _, k := range sshKeys {
		item := sshKeyItemModel{}

		if k.Id != nil {
			item.Id = types.Int64Value(int64(*k.Id))
		} else {
			item.Id = types.Int64Null()
		}

		if k.Name != nil {
			item.Name = types.StringValue(*k.Name)
		} else {
			item.Name = types.StringNull()
		}

		if k.Key != nil {
			item.Key = types.StringValue(*k.Key)
		} else {
			item.Key = types.StringNull()
		}

		if k.KeyType != nil {
			item.KeyType = types.StringValue(*k.KeyType)
		} else {
			item.KeyType = types.StringNull()
		}

		if k.Fingerprint != nil {
			item.Fingerprint = types.StringValue(*k.Fingerprint)
		} else {
			item.Fingerprint = types.StringNull()
		}

		if k.CreatedAt != nil {
			item.CreatedAt = types.StringValue(*k.CreatedAt)
		} else {
			item.CreatedAt = types.StringNull()
		}

		if k.CreatedByName != nil {
			item.CreatedByName = types.StringValue(*k.CreatedByName)
		} else {
			item.CreatedByName = types.StringNull()
		}

		if k.CreatedById != nil {
			item.CreatedById = types.Int64Value(int64(*k.CreatedById))
		} else {
			item.CreatedById = types.Int64Null()
		}

		itemModels = append(itemModels, item)
	}

	keyList, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: sshKeyItemModel{}.attrTypes()}, itemModels)
	diags.Append(listDiags...)

	return keyList, diags
}
