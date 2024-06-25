package provider

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	k0sctl_phase "github.com/k0sproject/k0sctl/phase"

	phase "github.com/mirantis/terraform-provider-k0sctl/internal/k0sctl/phase"

	k0sctl_action "github.com/k0sproject/k0sctl/action"

	k0sctl_v1beta1 "github.com/k0sproject/k0sctl/pkg/apis/k0sctl.k0sproject.io/v1beta1"
)

var _ resource.Resource = &K0sctlConfigResource{}

type K0sctlConfigResource struct {
	testingMode bool
}

func NewK0sctlConfigResource() resource.Resource {
	return &K0sctlConfigResource{}
}

func (r *K0sctlConfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	tflog.Info(ctx, "k0sctl metadata run", map[string]interface{}{})
	resp.TypeName = req.ProviderTypeName + "_config"
}

func (r *K0sctlConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = k0sctl_v1beta1_schema()
}

func (r *K0sctlConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	kpm, ok := req.ProviderData.(*K0sctlProviderModel)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *LaunchpadProviderModel, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.testingMode = kpm.testingMode
}

func (r *K0sctlConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var kcsm k0sctlSchemaModel
	var kcc k0sctl_v1beta1.Cluster

	resp.Diagnostics.Append(req.Plan.Get(ctx, &kcsm)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if tkcc, ds := kcsm.Cluster(ctx); ds.HasError() {
		resp.Diagnostics.Append(ds...)
	} else if err := tkcc.Validate(); err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("k0sctl cluster validation failed", err.Error()))
	} else {
		kcc = tkcc
	}

	var pm *k0sctl_phase.Manager
	var kc io.ReadWriter // will be used to contain kubeconfig, written in the phasemanager, and passed back to the model

	if tpm, err := k0sctl_phase.NewManager(&kcc); err != nil {
		d := diag.NewErrorDiagnostic("k0sctl phase manager creation failed", err.Error())
		resp.Diagnostics.Append(d)
	} else {
		pm = tpm
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Error(ctx, "#################### Adding a new phase within create ###############################", map[string]interface{}{})

	//pm.AddPhase(&phase.ValidateHostsExtended{})

	kc = bytes.NewBuffer([]byte{})

	aa := k0sctl_action.Apply{
		Force:         kcsm.Force.ValueBool(),
		Manager:       pm,
		KubeconfigOut: kc,
		//KubeconfigAPIAddress:  kcsm.??
		NoWait:                kcsm.NoWait.ValueBool(),
		NoDrain:               kcsm.NoDrain.ValueBool(),
		DisableDowngradeCheck: kcsm.DisableDowngradeCheck.ValueBool(),
		RestoreFrom:           kcsm.RestoreFrom.ValueString(),
	}

	kcsm.KubeYaml = types.StringNull()
	kcsm.KubeHost = types.StringNull()
	kcsm.CaCert = types.StringNull()
	kcsm.PrivateKey = types.StringNull()
	kcsm.ClientCert = types.StringNull()
	kcsm.Id = kcsm.Metadata.Name

	if kcsm.SkipCreate.ValueBool() {
		resp.Diagnostics.AddWarning("skipping create", "Skipping the k0sctl create because of configuration flag.")
		resp.Diagnostics.Append(resp.State.Set(ctx, kcsm)...)
	} else if r.testingMode {
		resp.Diagnostics.AddWarning("testing mode warning", "k0sctl config resource handler is in testing mode, no installation will be run.")
		resp.Diagnostics.Append(resp.State.Set(ctx, kcsm)...)
	} else if err := aa.Run(); err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("error running k0sctl apply", err.Error()))
	} else {
		// populate the model kubernetes conf from the action
		resp.Diagnostics.Append(kcsm.AddKubeconfig(kc)...)
	}

	tflog.Error(ctx, "#################### After running apply ###############################", map[string]interface{}{})

	if resp.Diagnostics.HasError() {
		return
	}

	kcsm.Id = kcsm.Metadata.Name

	if diags := resp.State.Set(ctx, kcsm); diags != nil {
		resp.Diagnostics.Append(diags...)
	}
}

func (r *K0sctlConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// k0sctl has no good way to discover existing installation, so we don't do anything
	tflog.Error(ctx, "#################### Start calling read ###############################", map[string]interface{}{})
}

func (r *K0sctlConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var kcsm k0sctlSchemaModel
	var kcc k0sctl_v1beta1.Cluster

	tflog.Error(ctx, "#################### Start calling Updated ###############################", map[string]interface{}{})

	resp.Diagnostics.Append(req.Plan.Get(ctx, &kcsm)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if tkcc, ds := kcsm.Cluster(ctx); ds.HasError() {
		resp.Diagnostics.Append(ds...)
	} else if err := tkcc.Validate(); err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("k0sctl cluster validation failed", err.Error()))
	} else {
		kcc = tkcc
	}

	var pm *k0sctl_phase.Manager
	var kc io.ReadWriter // will be used to contain kubeconfig, written in the phasemanager, and passed back to the model

	if tpm, err := k0sctl_phase.NewManager(&kcc); err != nil {
		d := diag.NewErrorDiagnostic("k0sctl phase manager creation failed", err.Error())
		resp.Diagnostics.Append(d)
	} else {
		pm = tpm
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Error(ctx, "#################### Adding a new phase within update method ###############################", map[string]interface{}{})
	pm.AddPhase(&phase.ValidateHostsExtended{})

	kc = bytes.NewBuffer([]byte{})

	aa := k0sctl_action.Apply{
		Force:         kcsm.Force.ValueBool(),
		Manager:       pm,
		KubeconfigOut: kc,
		//KubeconfigAPIAddress:  kcsm.??
		NoWait:                kcsm.NoWait.ValueBool(),
		NoDrain:               kcsm.NoDrain.ValueBool(),
		DisableDowngradeCheck: kcsm.DisableDowngradeCheck.ValueBool(),
		RestoreFrom:           kcsm.RestoreFrom.ValueString(),
	}

	if kcsm.SkipCreate.ValueBool() {
		resp.Diagnostics.AddWarning("skipping update", "Skipping the k0sctl create because of configuration flag.")
	} else if r.testingMode {
		resp.Diagnostics.AddWarning("testing mode warning", "k0sctl config resource handler is in testing mode, no installation will be run.")

		kcsm.KubeYaml = types.StringNull()
		kcsm.KubeHost = types.StringNull()
		kcsm.CaCert = types.StringNull()
		kcsm.PrivateKey = types.StringNull()
		kcsm.ClientCert = types.StringNull()

		kcsm.Id = kcsm.Metadata.Name

		if diags := resp.State.Set(ctx, kcsm); diags != nil {
			resp.Diagnostics.Append(diags...)
		}
	} else if err := aa.Run(); err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("error running k0sctl apply", err.Error()))
	} else {
		// populate the model kubernetes conf from the action
		resp.Diagnostics.Append(kcsm.AddKubeconfig(kc)...)
	}

	tflog.Error(ctx, "#################### After running apply Update ###############################", map[string]interface{}{})

	if resp.Diagnostics.HasError() {
		return
	}

	if diags := resp.State.Set(ctx, kcsm); diags != nil {
		resp.Diagnostics.Append(diags...)
	}
}

func (r *K0sctlConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var kcsm k0sctlSchemaModel
	var kcc k0sctl_v1beta1.Cluster

	resp.Diagnostics.Append(req.State.Get(ctx, &kcsm)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if tkcc, ds := kcsm.Cluster(ctx); ds.HasError() {
		resp.Diagnostics.Append(ds...)
	} else if err := tkcc.Validate(); err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("k0sctl cluster validation failed", err.Error()))
	} else {
		kcc = tkcc
	}

	var pm *k0sctl_phase.Manager

	if tpm, err := k0sctl_phase.NewManager(&kcc); err != nil {
		d := diag.NewErrorDiagnostic("k0sctl phase manager creation failed", err.Error())
		resp.Diagnostics.Append(d)
	} else {
		pm = tpm
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ra := k0sctl_action.Reset{
		Manager: pm,
		Force:   true, // k0sctl asks for confirmation if this is not set to true
		Stdout:  nil,  // TODO: turn this into a tflog outputter?
	}

	if kcsm.SkipDestroy.ValueBool() {
		resp.Diagnostics.AddWarning("skipping create", "Skipping the k0sctl destroy because of configuration flag.")
	} else if r.testingMode {
		resp.Diagnostics.AddWarning("testing mode warning", "k0sctl config resource handler is in testing mode, no reset will be run.")

		kcsm.Id = kcsm.Metadata.Name

		if diags := resp.State.Set(ctx, kcsm); diags != nil {
			resp.Diagnostics.Append(diags...)
		}
	} else if err := ra.Run(); err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("error running k0sctl reset", err.Error()))
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r *K0sctlConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// an import is an invalid operation for k0sctl, as it will want to run anyway. Just add the resource and apply it.
	resp.Diagnostics.AddError("K0sctl imports are invalid", "The k0sctl resource does not support imports, as launchpad itself doesn't maintain state. Just add the resource and hit apply.")
}
