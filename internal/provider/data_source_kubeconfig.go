package provider

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/talos-systems/talos/cmd/talosctl/pkg/talos/helpers"
	"github.com/talos-systems/talos/pkg/machinery/api/machine"
	tc "github.com/talos-systems/talos/pkg/machinery/client"
	"github.com/talos-systems/talos/pkg/machinery/constants"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"
	api "k8s.io/client-go/tools/clientcmd/api/v1"
)

func dataSourceKubeconfig() *schema.Resource {
	return &schema.Resource{
		Description: "Download the kubeconfig information from a Talos node.",

		ReadContext: dataSourceKubeconfigRead,

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

func dataSourceKubeconfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	return kubeconfigRead(ctx, client, d)
}

func kubeconfigRead(ctx context.Context, client machine.MachineServiceClient, d *schema.ResourceData) diag.Diagnostics {
	stream, err := client.Kubeconfig(ctx, &emptypb.Empty{})
	if err != nil {
		return diag.FromErr(err)
	}

	r, errCh, err := tc.ReadStream(stream)
	if err != nil {
		return diag.FromErr(err)
	}

	defer r.Close()

	kubeconfigRaw, err := helpers.ExtractFileFromTarGz("kubeconfig", r)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := <-errCh; err != nil {
		return diag.FromErr(err)
	}

	d.Set("raw", string(kubeconfigRaw))

	var kubeconfig api.Config
	err = yaml.Unmarshal(kubeconfigRaw, &kubeconfig)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(kubeconfig.Clusters) == 0 || len(kubeconfig.AuthInfos) == 0 {
		return diag.Errorf("Invalid kubeconfig file.")
	}

	cluster := kubeconfig.Clusters[0].Cluster
	user := kubeconfig.AuthInfos[0].AuthInfo

	d.Set("client_certificate", string(user.ClientCertificateData))

	d.Set("client_key", string(user.ClientKeyData))

	d.Set("cluster_ca_certificate", string(cluster.CertificateAuthorityData))

	d.SetId(fmt.Sprintf("%x", sha256.Sum256(kubeconfigRaw)))

	return nil
}
