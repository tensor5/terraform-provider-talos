package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/talos-systems/talos/cmd/talosctl/cmd/mgmt"
	"github.com/talos-systems/talos/cmd/talosctl/pkg/mgmt/helpers"
	"github.com/talos-systems/talos/pkg/images"
	"github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
	"github.com/talos-systems/talos/pkg/machinery/constants"
	"gopkg.in/yaml.v3"
)

func resourseGenConfig() *schema.Resource {
	return &schema.Resource{
		Description: "Generates a configuration for Talos cluster.",

		CreateContext: resourceGenConfigCreate,
		ReadContext:   resourceGenConfigRead,
		UpdateContext: resourceGenConfigUpdate,
		DeleteContext: resourceGenConfigDelete,

		Schema: map[string]*schema.Schema{
			"cluster_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Cluster name.",
			},
			"cluster_endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Cluster endpoint.",
			},
			"kubernetes_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     constants.DefaultKubernetesVersion,
				Description: fmt.Sprintf("Desired kubernetes version to run (default \"%s\").", constants.DefaultKubernetesVersion),
			},
			"config_patch": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Patch generated machineconfigs (applied to all node types).",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"config_patch_control_plane": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Patch generated machineconfigs (applied to 'init' and 'controlplane' types).",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"config_patch_worker": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Patch generated machineconfigs (applied to 'worker' type).",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"control_plane_config": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
			"worker_config": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
			"talos_config": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
			"install_disk": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "/dev/sda",
				Description: "The disk to install to.",
			},
			"install_image": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     helpers.DefaultImage(images.DefaultInstallerImageRepository),
				Description: "The image used to perform an installation.",
			},
			"additional_sans": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Additional Subject-Alt-Names for the APIServer certificate.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"dns_domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "cluster.local",
				Description: "The dns domain to use for cluster.",
			},
			"persist": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "The desired persist value for configs.",
			},
			"with_cluster_discovery": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Enable cluster discovery feature.",
			},
			"talos_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The desired Talos version to generate config for (backwards compatibility, e.g. v0.8).",
			},
			"registry_mirrors": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "List of registry mirrors to use in format: <registry host>=<mirror URL>.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"with_kubespan": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable KubeSpan feature.",
			},
		},
	}
}

func toStringArray(arr []interface{}) []string {
	var ret = make([]string, len(arr))
	for i, elem := range arr {
		ret[i] = elem.(string)
	}
	return ret
}

func resourceGenConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var genOptions []generate.GenOption

	var registryMirrors = d.Get("registry_mirrors").(map[string]interface{})
	for key, value := range registryMirrors {
		genOptions = append(genOptions, generate.WithRegistryMirror(key, value.(string)))
	}

	var talosVersion = d.Get("talos_version").(string)
	if talosVersion != "" {
		versionContract, err := config.ParseContractFromVersion(talosVersion)
		if err != nil {
			return diag.FromErr(err)
		}

		genOptions = append(genOptions, generate.WithVersionContract(versionContract))
	}

	if d.Get("with_kubespan").(bool) {
		genOptions = append(genOptions,
			generate.WithNetworkOptions(
				v1alpha1.WithKubeSpan(),
			),
		)
	}

	genOptions = append(genOptions,
		generate.WithInstallDisk(d.Get("install_disk").(string)),
		generate.WithInstallImage(d.Get("install_image").(string)),
		generate.WithAdditionalSubjectAltNames(toStringArray(d.Get("additional_sans").([]interface{}))),
		generate.WithDNSDomain(d.Get("dns_domain").(string)),
		generate.WithPersist(d.Get("persist").(bool)),
		generate.WithClusterDiscovery(d.Get("with_cluster_discovery").(bool)),
	)

	configBundle, err := mgmt.GenV1Alpha1Config(
		genOptions,
		d.Get("cluster_name").(string),
		d.Get("cluster_endpoint").(string),
		d.Get("kubernetes_version").(string),
		toStringArray(d.Get("config_patch").([]interface{})),
		toStringArray(d.Get("config_patch_control_plane").([]interface{})),
		toStringArray(d.Get("config_patch_worker").([]interface{})),
	)
	if err != nil {
		return diag.FromErr(err)
	}

	controlPlaneConfig, err := yaml.Marshal(configBundle.ControlPlaneCfg)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("control_plane_config", string(controlPlaneConfig))

	workerConfig, err := yaml.Marshal(configBundle.WorkerCfg)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("worker_config", string(workerConfig))

	talosConfig, err := yaml.Marshal(configBundle.TalosCfg)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("talos_config", string(talosConfig))

	d.SetId(uuid.NewString())

	return nil
}

func resourceGenConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceGenConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceGenConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
