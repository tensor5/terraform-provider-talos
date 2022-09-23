package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/talos-systems/talos/cmd/talosctl/cmd/mgmt"
	"github.com/talos-systems/talos/cmd/talosctl/pkg/mgmt/helpers"
	"github.com/talos-systems/talos/pkg/images"
	"github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
	"github.com/talos-systems/talos/pkg/machinery/constants"
	"github.com/tensor5/terraform-provider-talos/internal/provider/attribute_plan_modifier"
	"gopkg.in/yaml.v3"
)

var _ resource.Resource = &GenConfigResource{}

func NewGenConfigResource() resource.Resource {
	return &GenConfigResource{}
}

type GenConfigResource struct{}

type GenConfigResourceModel struct {
	ClusterName             types.String `tfsdk:"cluster_name"`
	ClusterEndpoint         types.String `tfsdk:"cluster_endpoint"`
	KubernetesVersion       types.String `tfsdk:"kubernetes_version"`
	ConfigPatch             types.List   `tfsdk:"config_patch"`
	ConfigPatchControlPlane types.List   `tfsdk:"config_patch_control_plane"`
	ConfigPatchWorker       types.List   `tfsdk:"config_patch_worker"`
	ControlPlaneConfig      types.String `tfsdk:"control_plane_config"`
	WorkerConfig            types.String `tfsdk:"worker_config"`
	TalosConfig             types.String `tfsdk:"talos_config"`
	InstallDisk             types.String `tfsdk:"install_disk"`
	InstallImage            types.String `tfsdk:"install_image"`
	AdditionalSans          types.List   `tfsdk:"additional_sans"`
	DnsDomain               types.String `tfsdk:"dns_domain"`
	Persist                 types.Bool   `tfsdk:"persist"`
	WithClusterDiscovery    types.Bool   `tfsdk:"with_cluster_discovery"`
	TalosVersion            types.String `tfsdk:"talos_version"`
	RegistryMirrors         types.Map    `tfsdk:"registry_mirrors"`
	WithKubespan            types.Bool   `tfsdk:"with_kubespan"`
}

func (r *GenConfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gen_config"
}

func (r *GenConfigResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Generates a configuration for Talos cluster.",

		Attributes: map[string]tfsdk.Attribute{
			"cluster_name": {
				MarkdownDescription: "Cluster name.",
				Required:            true,
				Type:                types.StringType,
			},
			"cluster_endpoint": {
				MarkdownDescription: "Cluster endpoint.",
				Required:            true,
				Type:                types.StringType,
			},
			"kubernetes_version": {
				Computed:            true,
				MarkdownDescription: fmt.Sprintf("Desired kubernetes version to run (default \"%s\").", constants.DefaultKubernetesVersion),
				Optional:            true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					attribute_plan_modifier.DefaultValue(types.String{Value: constants.DefaultKubernetesVersion}),
				},
				Type: types.StringType,
			},
			"config_patch": {
				MarkdownDescription: "Patch generated machineconfigs (applied to all node types).",
				Optional:            true,
				Type: types.ListType{
					ElemType: types.StringType,
				},
			},
			"config_patch_control_plane": {
				MarkdownDescription: "Patch generated machineconfigs (applied to 'init' and 'controlplane' types).",
				Optional:            true,
				Type: types.ListType{
					ElemType: types.StringType,
				},
			},
			"config_patch_worker": {
				MarkdownDescription: "Patch generated machineconfigs (applied to 'worker' type).",
				Optional:            true,
				Type: types.ListType{
					ElemType: types.StringType,
				},
			},
			"control_plane_config": {
				Computed:  true,
				Sensitive: true,
				Type:      types.StringType,
			},
			"worker_config": {
				Computed:  true,
				Sensitive: true,
				Type:      types.StringType,
			},
			"talos_config": {
				Computed:  true,
				Sensitive: true,
				Type:      types.StringType,
			},
			"install_disk": {
				Computed:            true,
				MarkdownDescription: "The disk to install to.",
				Optional:            true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					attribute_plan_modifier.DefaultValue(types.String{Value: "/dev/sda"}),
				},
				Type: types.StringType,
			},
			"install_image": {
				Computed:            true,
				MarkdownDescription: "The image used to perform an installation.",
				Optional:            true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					attribute_plan_modifier.DefaultValue(types.String{Value: helpers.DefaultImage(images.DefaultInstallerImageRepository)}),
				},
				Type: types.StringType,
			},
			"additional_sans": {
				MarkdownDescription: "Additional Subject-Alt-Names for the APIServer certificate.",
				Optional:            true,
				Type: types.ListType{
					ElemType: types.StringType,
				},
			},
			"dns_domain": {
				Computed:            true,
				MarkdownDescription: "The dns domain to use for cluster.",
				Optional:            true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					attribute_plan_modifier.DefaultValue(types.String{Value: "cluster.local"}),
				},
				Type: types.StringType,
			},
			"persist": {
				Computed:            true,
				MarkdownDescription: "The desired persist value for configs.",
				Optional:            true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					attribute_plan_modifier.DefaultValue(types.Bool{Value: true}),
				},
				Type: types.BoolType,
			},
			"with_cluster_discovery": {
				Computed:            true,
				MarkdownDescription: "Enable cluster discovery feature.",
				Optional:            true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					attribute_plan_modifier.DefaultValue(types.Bool{Value: true}),
				},
				Type: types.BoolType,
			},
			"talos_version": {
				MarkdownDescription: "The desired Talos version to generate config for (backwards compatibility, e.g. v0.8).",
				Optional:            true,
				Type:                types.StringType,
			},
			"registry_mirrors": {
				MarkdownDescription: "List of registry mirrors to use in format: <registry host>=<mirror URL>.",
				Optional:            true,
				Type: types.MapType{
					ElemType: types.StringType,
				},
			},
			"with_kubespan": {
				Computed:            true,
				MarkdownDescription: "Enable KubeSpan feature.",
				Optional:            true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					attribute_plan_modifier.DefaultValue(types.Bool{Value: false}),
				},
				Type: types.BoolType,
			},
		},
	}, nil
}

func (r *GenConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *GenConfigResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var genOptions []generate.GenOption

	var registryMirrors map[string]string
	resp.Diagnostics.Append(data.RegistryMirrors.ElementsAs(ctx, &registryMirrors, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	for key, value := range registryMirrors {
		genOptions = append(genOptions, generate.WithRegistryMirror(key, value))
	}

	if !data.TalosVersion.Null {
		versionContract, err := config.ParseContractFromVersion(data.TalosVersion.Value)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error parsing Talos version",
				err.Error(),
			)
			return
		}

		genOptions = append(genOptions, generate.WithVersionContract(versionContract))
	}

	if !data.WithKubespan.Null && data.WithKubespan.Value {
		genOptions = append(genOptions,
			generate.WithNetworkOptions(
				v1alpha1.WithKubeSpan(),
			),
		)
	}

	var additionalSans []string
	resp.Diagnostics.Append(data.AdditionalSans.ElementsAs(ctx, &additionalSans, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	genOptions = append(genOptions,
		generate.WithInstallDisk(data.InstallDisk.Value),
		generate.WithInstallImage(data.InstallImage.Value),
		generate.WithAdditionalSubjectAltNames(additionalSans),
		generate.WithDNSDomain(data.DnsDomain.Value),
		generate.WithPersist(data.Persist.Value),
		generate.WithClusterDiscovery(data.WithClusterDiscovery.Value),
	)

	var configPatch []string
	resp.Diagnostics.Append(data.ConfigPatch.ElementsAs(ctx, &configPatch, false)...)
	var configPatchControlPlane []string
	resp.Diagnostics.Append(data.ConfigPatchControlPlane.ElementsAs(ctx, &configPatchControlPlane, false)...)
	var configPatchWorker []string
	resp.Diagnostics.Append(data.ConfigPatchWorker.ElementsAs(ctx, &configPatchWorker, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	configBundle, err := mgmt.GenV1Alpha1Config(
		genOptions,
		data.ClusterName.Value,
		data.ClusterEndpoint.Value,
		data.KubernetesVersion.Value,
		configPatch,
		configPatchControlPlane,
		configPatchWorker,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error generating config",
			err.Error(),
		)
		return
	}

	controlPlaneConfig, err := yaml.Marshal(configBundle.ControlPlaneCfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting control plane configuration to YAML",
			err.Error(),
		)
		return
	}
	data.ControlPlaneConfig = types.String{Value: string(controlPlaneConfig)}

	workerConfig, err := yaml.Marshal(configBundle.WorkerCfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting worker configuration to YAML",
			err.Error(),
		)
		return
	}
	data.WorkerConfig = types.String{Value: string(workerConfig)}

	talosConfig, err := yaml.Marshal(configBundle.TalosCfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting Talos configuration to YAML",
			err.Error(),
		)
		return
	}
	data.TalosConfig = types.String{Value: string(talosConfig)}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a Talos cluster configuration resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GenConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *GenConfigResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GenConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *GenConfigResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GenConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *GenConfigResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}
