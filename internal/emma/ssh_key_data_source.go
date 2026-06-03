package emma

import (
	"context"
	"fmt"
	"net/http"

	emmaSdk "github.com/emma-community/emma-go-sdk"
	"github.com/emma-community/terraform-provider-emma/tools"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &sshKeyDataSource{}

func NewSshKeyDataSource() datasource.DataSource {
	return &sshKeyDataSource{}
}

// sshKeyDataSource defines the data source implementation.
type sshKeyDataSource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

// sshKeyDataSourceModel describes the data source data model.
type sshKeyDataSourceModel struct {
	Id            types.Int64  `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Key           types.String `tfsdk:"key"`
	KeyType       types.String `tfsdk:"key_type"`
	Fingerprint   types.String `tfsdk:"fingerprint"`
	CreatedAt     types.String `tfsdk:"created_at"`
	CreatedByName types.String `tfsdk:"created_by_name"`
	CreatedById   types.Int64  `tfsdk:"created_by_id"`
}

func (d *sshKeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (d *sshKeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides information about an SSH key by its ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "SSH key ID",
				Required:    true,
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
	}
}

func (d *sshKeyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sshKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data sshKeyDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int32(data.Id.ValueInt64())

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *d.token.AccessToken)
	sshKey, response, err := d.apiClient.SSHKeysAPI.GetSshKey(auth, id).Execute()

	if err != nil {
		if response != nil && response.StatusCode == http.StatusNotFound {
			resp.Diagnostics.AddError("SSH key not found",
				fmt.Sprintf("SSH key with id %d not found", id))
		} else {
			resp.Diagnostics.AddError("Client Error",
				fmt.Sprintf("Unable to read SSH key, got error: %s",
					tools.ExtractErrorMessage(response)))
		}
		return
	}

	if sshKey.Id != nil {
		data.Id = types.Int64Value(int64(*sshKey.Id))
	} else {
		data.Id = types.Int64Null()
	}

	if sshKey.Name != nil {
		data.Name = types.StringValue(*sshKey.Name)
	} else {
		data.Name = types.StringNull()
	}

	if sshKey.Key != nil {
		data.Key = types.StringValue(*sshKey.Key)
	} else {
		data.Key = types.StringNull()
	}

	if sshKey.KeyType != nil {
		data.KeyType = types.StringValue(*sshKey.KeyType)
	} else {
		data.KeyType = types.StringNull()
	}

	if sshKey.Fingerprint != nil {
		data.Fingerprint = types.StringValue(*sshKey.Fingerprint)
	} else {
		data.Fingerprint = types.StringNull()
	}

	if sshKey.CreatedAt != nil {
		data.CreatedAt = types.StringValue(*sshKey.CreatedAt)
	} else {
		data.CreatedAt = types.StringNull()
	}

	if sshKey.CreatedByName != nil {
		data.CreatedByName = types.StringValue(*sshKey.CreatedByName)
	} else {
		data.CreatedByName = types.StringNull()
	}

	if sshKey.CreatedById != nil {
		data.CreatedById = types.Int64Value(int64(*sshKey.CreatedById))
	} else {
		data.CreatedById = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
