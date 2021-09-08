package provider

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/talos-systems/talos/pkg/machinery/api/machine"
	"github.com/talos-systems/talos/pkg/machinery/constants"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func resourseBootstrap() *schema.Resource {
	return &schema.Resource{
		Description: "Bootstrap a Talos cluster and download kubeconfig.",

		CreateContext: resourceBootstrapCreate,
		ReadContext:   dataSourceKubeconfigRead,
		UpdateContext: dataSourceKubeconfigRead,
		DeleteContext: resourceBootstrapDelete,

		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Address of Talos node handling the request.",
			},
			"machine_ca": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "PEM-encoded root certificates bundle for TLS authentication.",
			},
			"machine_crt": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "PEM-encoded client certificate for TLS authentication.",
			},
			"machine_key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "PEM-encoded client certificate key for TLS authentication.",
			},
			"client_certificate": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "PEM-encoded client certificate for TLS authentication.",
			},
			"client_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "PEM-encoded client certificate key for TLS authentication.",
			},
			"cluster_ca_certificate": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "PEM-encoded root certificates bundle for TLS authentication.",
			},
			"raw": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Content of kubeconfig file.",
			},
		},
	}
}

func resourceBootstrapCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clientCert, err := tls.X509KeyPair(
		[]byte(d.Get("machine_crt").(string)),
		[]byte(d.Get("machine_key").(string)),
	)
	if err != nil {
		return diag.FromErr(err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM([]byte(d.Get("machine_ca").(string))) {
		return diag.Errorf("failed to add server CA's certificate")
	}

	tlsCredentials := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	})

	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", d.Get("endpoint"), constants.ApidPort),
		grpc.WithTransportCredentials(tlsCredentials),
		grpc.WithBlock(),
	)
	if err != nil {
		return diag.FromErr(err)
	}
	defer conn.Close()
	client := machine.NewMachineServiceClient(conn)

	if _, err := client.Bootstrap(ctx, &machine.BootstrapRequest{}); err != nil {
		return diag.FromErr(err)
	}

	return kubeconfigRead(ctx, client, d)
}

func resourceBootstrapDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
