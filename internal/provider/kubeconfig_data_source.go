package provider

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/talos-systems/talos/cmd/talosctl/pkg/talos/helpers"
	"github.com/talos-systems/talos/pkg/machinery/api/machine"
	tc "github.com/talos-systems/talos/pkg/machinery/client"
	"github.com/talos-systems/talos/pkg/machinery/constants"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"
	api "k8s.io/client-go/tools/clientcmd/api/v1"
)

var _ datasource.DataSource = &KubeconfigDataSource{}

func NewKubeconfigDataSource() datasource.DataSource {
	return &KubeconfigDataSource{}
}

type KubeconfigDataSource struct{}

type KubeconfigDataSourceModel struct {
	Endpoint             types.String `tfsdk:"endpoint"`
	MachineCa            types.String `tfsdk:"machine_ca"`
	MachineCrt           types.String `tfsdk:"machine_crt"`
	MachineKey           types.String `tfsdk:"machine_key"`
	ClientCertificate    types.String `tfsdk:"client_certificate"`
	ClientKey            types.String `tfsdk:"client_key"`
	ClusterCaCertificate types.String `tfsdk:"cluster_ca_certificate"`
	Raw                  types.String `tfsdk:"raw"`
}

func (d *KubeconfigDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubeconfig"
}

var attributes = map[string]tfsdk.Attribute{
	"endpoint": {
		MarkdownDescription: "Address of Talos node handling the request.",
		Required:            true,
		Type:                types.StringType,
	},
	"machine_ca": {
		MarkdownDescription: "PEM-encoded root certificates bundle for TLS authentication.",
		Required:            true,
		Type:                types.StringType,
	},
	"machine_crt": {
		MarkdownDescription: "PEM-encoded client certificate for TLS authentication.",
		Required:            true,
		Type:                types.StringType,
	},
	"machine_key": {
		MarkdownDescription: "PEM-encoded client certificate key for TLS authentication.",
		Required:            true,
		Type:                types.StringType,
	},
	"client_certificate": {
		Computed:            true,
		MarkdownDescription: "PEM-encoded client certificate for TLS authentication.",
		Type:                types.StringType,
	},
	"client_key": {
		Computed:            true,
		MarkdownDescription: "PEM-encoded client certificate key for TLS authentication.",
		Type:                types.StringType,
	},
	"cluster_ca_certificate": {
		Computed:            true,
		MarkdownDescription: "PEM-encoded root certificates bundle for TLS authentication.",
		Type:                types.StringType,
	},
	"raw": {
		Computed:            true,
		MarkdownDescription: "Content of kubeconfig file.",
		Type:                types.StringType,
	},
}

func (d *KubeconfigDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Download the kubeconfig information from a Talos node.",

		Attributes: attributes,
	}, nil
}

func (d *KubeconfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *KubeconfigDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	clientCert, err := tls.X509KeyPair(
		[]byte(data.MachineCrt.Value),
		[]byte(data.MachineKey.Value),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing key pair",
			err.Error(),
		)
		return
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM([]byte(data.MachineCa.Value)) {
		resp.Diagnostics.AddError(
			"failed to add server CA's certificate",
			"",
		)
		return
	}

	tlsCredentials := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	})

	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", data.Endpoint.Value, constants.ApidPort),
		grpc.WithTransportCredentials(tlsCredentials),
		grpc.WithBlock(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating gRPC connection",
			err.Error(),
		)
		return
	}
	defer conn.Close()
	client := machine.NewMachineServiceClient(conn)

	if err := kubeconfigRead(ctx, client, data); err != nil {
		resp.Diagnostics.AddError(
			"Error reading kubeconfig",
			err.Error(),
		)
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a Talos kubeconfig data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func kubeconfigRead(ctx context.Context, client machine.MachineServiceClient, d *KubeconfigDataSourceModel) error {
	stream, err := client.Kubeconfig(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}

	r, errCh, err := tc.ReadStream(stream)
	if err != nil {
		return err
	}

	defer r.Close()

	kubeconfigRaw, err := helpers.ExtractFileFromTarGz("kubeconfig", r)
	if err != nil {
		return err
	}

	if err := <-errCh; err != nil {
		return err
	}

	d.Raw = types.String{Value: string(kubeconfigRaw)}

	var kubeconfig api.Config
	err = yaml.Unmarshal(kubeconfigRaw, &kubeconfig)
	if err != nil {
		return err
	}

	if len(kubeconfig.Clusters) == 0 || len(kubeconfig.AuthInfos) == 0 {
		return errors.New("invalid kubeconfig file")
	}

	cluster := kubeconfig.Clusters[0].Cluster
	user := kubeconfig.AuthInfos[0].AuthInfo

	d.ClientCertificate = types.String{Value: string(user.ClientCertificateData)}

	d.ClientKey = types.String{Value: string(user.ClientKeyData)}

	d.ClusterCaCertificate = types.String{Value: string(cluster.CertificateAuthorityData)}

	return nil
}
