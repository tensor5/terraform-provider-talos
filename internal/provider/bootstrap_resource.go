package provider

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/talos-systems/talos/pkg/machinery/api/machine"
	"github.com/talos-systems/talos/pkg/machinery/constants"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var _ resource.Resource = &BootstrapResource{}

func NewBootstrapResource() resource.Resource {
	return &BootstrapResource{}
}

type BootstrapResource struct{}

type BootstrapResourceModel = KubeconfigDataSourceModel

func (r *BootstrapResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bootstrap"
}

func (r *BootstrapResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Bootstrap a Talos cluster and download kubeconfig.",

		Attributes: attributes,
	}, nil
}

func (r *BootstrapResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *BootstrapResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

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

	if _, err := client.Bootstrap(ctx, &machine.BootstrapRequest{}); err != nil {
		resp.Diagnostics.AddError(
			"Error in bootstrap request",
			err.Error(),
		)
		return
	}

	if err := kubeconfigRead(ctx, client, data); err != nil {
		resp.Diagnostics.AddError(
			"Error reading kubeconfig",
			err.Error(),
		)
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a Talos bootstrap resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BootstrapResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *BootstrapResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BootstrapResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *BootstrapResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BootstrapResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *BootstrapResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}
