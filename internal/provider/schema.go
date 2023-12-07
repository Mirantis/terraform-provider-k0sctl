package provider

import (
	"context"
	"io"

	"gopkg.in/yaml.v2"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	k0s_dig "github.com/k0sproject/dig"
	k0sctl_v1beta1 "github.com/k0sproject/k0sctl/pkg/apis/k0sctl.k0sproject.io/v1beta1"
	k0sctl_v1beta1_cluster "github.com/k0sproject/k0sctl/pkg/apis/k0sctl.k0sproject.io/v1beta1/cluster"
	k0s_rig "github.com/k0sproject/rig"
	k0sversion "github.com/k0sproject/version"
)

const (
	k0sctl_schema_kind = "cluster"
)

func k0sctl_v1beta1_schema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Mirantis installation using launchpad, parametrized",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Example identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"skip_destroy": schema.BoolAttribute{
				MarkdownDescription: "Skip reset on destroy",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"skip_create": schema.BoolAttribute{
				MarkdownDescription: "Skip apply on create",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},

			"force": schema.BoolAttribute{
				MarkdownDescription: "Attempt a forced installation in case of certain failures",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},

			"no_wait": schema.BoolAttribute{
				MarkdownDescription: "Do not wait for worker nodes to join",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"no_drain": schema.BoolAttribute{
				MarkdownDescription: "Do not drain worker nodes when upgrading",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"disable_downgrade_check": schema.BoolAttribute{
				MarkdownDescription: "Skip downgrade check",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"restore_from": schema.StringAttribute{
				MarkdownDescription: "Path to cluster backup archive to restore the state from",
				Optional:            true,
			},

			"kube_skiptlsverify": schema.BoolAttribute{
				MarkdownDescription: "K8 Kubernetes endpoint TLS should not be verified",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},

			"k0s_yaml": schema.StringAttribute{
				MarkdownDescription: "K0S yaml for debugging ",
				Computed:            true,
			},

			"kube_yaml": schema.StringAttribute{
				MarkdownDescription: "K8 Kubernetes API client configuration yaml file",
				Computed:            true,
				Sensitive:           true,
			},

			"kube_host": schema.StringAttribute{
				MarkdownDescription: "K8 Kubernetes API host endpoint",
				Computed:            true,
			},

			"private_key": schema.StringAttribute{
				MarkdownDescription: "K8 Private key for the user",
				Computed:            true,
			},
			"client_cert": schema.StringAttribute{
				MarkdownDescription: "K8 Client certificate for the user",
				Computed:            true,
			},
			"ca_cert": schema.StringAttribute{
				MarkdownDescription: "K8 Server CA certificate",
				Computed:            true,
			},
		},

		Blocks: map[string]schema.Block{

			"metadata": schema.SingleNestedBlock{
				MarkdownDescription: "Metadata for the launchpad cluster",

				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Cluster name",
						Required:            true,
					},
				},
			},

			"spec": schema.SingleNestedBlock{
				MarkdownDescription: "Launchpad install specifications",

				Blocks: map[string]schema.Block{

					"k0s": schema.SingleNestedBlock{
						MarkdownDescription: "K0S installation configuration",

						Attributes: map[string]schema.Attribute{
							"version": schema.StringAttribute{
								MarkdownDescription: "K0s version to install",
								Required:            true,
							},

							"config": schema.StringAttribute{
								MarkdownDescription: "K0s config yaml as a string",
								Optional:            true,
							},
						},
					},

					"host": schema.ListNestedBlock{
						MarkdownDescription: "Individual host configuration, for each machine in the cluster",

						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},

						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"role": schema.StringAttribute{
									MarkdownDescription: "Host machine role in the cluster",
									Required:            true,
								},

								"install_flags": schema.ListAttribute{
									MarkdownDescription: "String install flags passed to k0s (e.g. '--taints=mytaint')",
									Optional:            true,
									ElementType:         types.StringType,
								},
							},

							Blocks: map[string]schema.Block{

								"hooks": schema.ListNestedBlock{
									MarkdownDescription: "Hook configuration for the host",

									Validators: []validator.List{
										listvalidator.SizeAtMost(1),
									},

									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{},
										Blocks: map[string]schema.Block{

											"apply": schema.ListNestedBlock{
												MarkdownDescription: "Launchpad.Apply string hooks for the host",

												Validators: []validator.List{
													listvalidator.SizeAtMost(1),
												},

												NestedObject: schema.NestedBlockObject{
													Attributes: map[string]schema.Attribute{
														"before": schema.ListAttribute{
															MarkdownDescription: "String hooks to run on hosts before the Apply operation is run.",
															ElementType:         types.StringType,
															Optional:            true,
															Computed:            true,
															Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
														},
														"after": schema.ListAttribute{
															MarkdownDescription: "String hooks to run on hosts after the Apply operation is run.",
															ElementType:         types.StringType,
															Optional:            true,
															Computed:            true,
															Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
														},
													},
												},
											},
										},
									},
								},

								"ssh": schema.ListNestedBlock{
									MarkdownDescription: "SSH configuration for the host",

									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"address": schema.StringAttribute{
												MarkdownDescription: "SSH endpoint",
												Required:            true,
											},
											"key_path": schema.StringAttribute{
												MarkdownDescription: "SSH endpoint",
												Required:            true,
											},
											"user": schema.StringAttribute{
												MarkdownDescription: "SSH endpoint",
												Required:            true,
											},
											"port": schema.Int64Attribute{
												MarkdownDescription: "SSH Port",
												Optional:            true,
												Computed:            true,
												Default:             int64default.StaticInt64(22),
											},
										},
									},
								},
								"winrm": schema.ListNestedBlock{
									MarkdownDescription: "WinRM configuration for the host",

									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"address": schema.StringAttribute{
												MarkdownDescription: "WinRM endpoint",
												Required:            true,
											},
											"user": schema.StringAttribute{
												MarkdownDescription: "WinRM user",
												Required:            true,
											},
											"password": schema.StringAttribute{
												MarkdownDescription: "WinRM password",
												Required:            true,
											},
											"port": schema.Int64Attribute{
												MarkdownDescription: "WinRM Port",
												Optional:            true,
												Computed:            true,
												Default:             int64default.StaticInt64(5985),
											},
											"use_https": schema.BoolAttribute{
												MarkdownDescription: "If false, then no HTTP is used for winrm transport",
												Optional:            true,
												Computed:            true,
												Default:             booldefault.StaticBool(true),
											},
											"insecure": schema.BoolAttribute{
												MarkdownDescription: "If false, then no SSL certificate validation is used",
												Optional:            true,
												Computed:            true,
												Default:             booldefault.StaticBool(true),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

type k0sctlSchemaModel struct {
	Id          types.String `tfsdk:"id"`
	SkipCreate  types.Bool   `tfsdk:"skip_create"`
	SkipDestroy types.Bool   `tfsdk:"skip_destroy"`

	Force                 types.Bool `tfsdk:"force"`
	NoWait                types.Bool `tfsdk:"no_wait"`
	NoDrain               types.Bool `tfsdk:"no_drain"`
	DisableDowngradeCheck types.Bool `tfsdk:"disable_downgrade_check"`

	RestoreFrom types.String `tfsdk:"restore_from"`

	K0sYaml types.String `tfsdk:"k0s_yaml"`

	PrivateKey types.String `tfsdk:"private_key"`
	ClientCert types.String `tfsdk:"client_cert"`
	CaCert     types.String `tfsdk:"ca_cert"`

	KubeYaml          types.String `tfsdk:"kube_yaml"`
	KubeHost          types.String `tfsdk:"kube_host"`
	KubeSkipTLSVerify types.Bool   `tfsdk:"kube_skiptlsverify"`

	Metadata k0sctlSchemaClusterMetadata `tfsdk:"metadata"`
	Spec     k0sctlSchemaModelSpec       `tfsdk:"spec"`
}

// Cluster build a k0sctl cluster configuration struct from the model data.
func (ksm *k0sctlSchemaModel) Cluster(ctx context.Context) (k0sctl_v1beta1.Cluster, diag.Diagnostics) {
	tflog.Info(ctx, "Creating k0sctl Cluster from schema", map[string]interface{}{})

	var c k0sctl_v1beta1.Cluster
	var d diag.Diagnostics

	var v *k0sversion.Version

	if vv, err := k0sversion.NewVersion(ksm.Spec.K0s.Version.ValueString()); err != nil {
		d.AddError("Could not interpret version", "Passed K0s version could not be parsed")
	} else {
		v = vv
	}

	c = k0sctl_v1beta1.Cluster{
		APIVersion: k0sctl_v1beta1.APIVersion,
		Kind:       k0sctl_schema_kind,

		Metadata: &k0sctl_v1beta1.ClusterMetadata{
			Name: ksm.Metadata.Name.ValueString(),
		},

		Spec: &k0sctl_v1beta1_cluster.Spec{
			Hosts: k0sctl_v1beta1_cluster.Hosts{},
			K0s: &k0sctl_v1beta1_cluster.K0s{
				Version: v,
			},
		},
	}

	if kc := ksm.Spec.K0s.Config.ValueString(); kc != "" {
		// if k0s.Config is not empty, then it needs to be
		// converted to a dig.Mapping
		var dm k0s_dig.Mapping

		if err := yaml.Unmarshal([]byte(kc), &dm); err != nil {
			d.AddWarning("K0s config unmarshal failed", err.Error())
		} else {
			c.Spec.K0s.Config = dm
		}
	}

	for _, sh := range ksm.Spec.Hosts {
		h := k0sctl_v1beta1_cluster.Host{
			Role:  sh.Role.ValueString(),
			Hooks: k0sctl_v1beta1_cluster.Hooks{},
		}

		if len(sh.InstallFlags) > 0 {
			var shifs = make([]string, len(sh.InstallFlags))
			for i, shif := range sh.InstallFlags {
				shifs[i] = shif.ValueString()
			}
			h.InstallFlags = k0sctl_v1beta1_cluster.Flags(shifs)
		}
		if len(sh.SSH) > 0 {
			shssh := sh.SSH[0]

			h.Connection = k0s_rig.Connection{
				SSH: &k0s_rig.SSH{
					Address: shssh.Address.ValueString(),
					KeyPath: shssh.KeyPath.ValueStringPointer(),
					User:    shssh.User.ValueString(),
					Port:    int(shssh.Port.ValueInt64()),
				},
			}
		} else if len(sh.WinRM) > 0 {
			shwinrm := sh.WinRM[0]

			h.Connection = k0s_rig.Connection{
				WinRM: &k0s_rig.WinRM{
					Address:  shwinrm.Address.ValueString(),
					Password: shwinrm.Password.ValueString(),
					User:     shwinrm.User.ValueString(),
					Port:     int(shwinrm.Port.ValueInt64()),
					UseHTTPS: shwinrm.UseHTTPS.ValueBool(),
					Insecure: shwinrm.Insecure.ValueBool(),
				},
			}
		}

		if len(sh.Hooks) > 0 {
			shh := sh.Hooks[0]

			if len(shh.Apply) > 0 {
				ha := shh.Apply[0]

				hha := map[string][]string{
					"before": {},
					"after":  {},
				}
				var shab []string
				if diag := ha.Before.ElementsAs(context.Background(), &shab, true); diag == nil {
					hha["before"] = shab
				}
				var shaa []string
				if diag := ha.After.ElementsAs(context.Background(), &shaa, true); diag == nil {
					hha["after"] = shab
				}

				h.Hooks["apply"] = hha
			}
		}

		c.Spec.Hosts = append(c.Spec.Hosts, &h)
	}

	// add the cluster yaml to the model
	if kyb, err := yaml.Marshal(c); err != nil {
		d.AddWarning("failed to marshall the k0sctl config to yaml", err.Error())
	} else {
		ksm.K0sYaml = types.StringValue(string(kyb))
	}

	return c, d
}

// AddKubeconfig read bytes for a kube config file, and interpret it into parametrized config values.
func (ksm *k0sctlSchemaModel) AddKubeconfig(r io.Reader) diag.Diagnostics {
	d := diag.Diagnostics{}
	k8bytes, _ := io.ReadAll(r)

	// Struct representation of a kube config file.
	// see https://zhwt.github.io/yaml-to-go/
	var cbkHolder struct {
		APIVersion  string            `yaml:"apiVersion"`
		Kind        string            `yaml:"kind"`
		Preferences map[string]string `yaml:"preferences"`
		Clusters    []struct {
			Name    string `yaml:"name"`
			Cluster struct {
				CertificateAuthorityData string `yaml:"certificate-authority-data"`
				Server                   string `yaml:"server"`
			} `yaml:"cluster"`
		} `yaml:"clusters"`
		Contexts []struct {
			Name    string `yaml:"name"`
			Context struct {
				Cluster string `yaml:"cluster"`
				User    string `yaml:"user"`
			} `yaml:"context"`
		} `yaml:"contexts"`
		CurrentContext string `yaml:"current-context"`
		Users          []struct {
			Name string `yaml:"name"`
			User struct {
				ClientCertificateData string `yaml:"client-certificate-data"`
				ClientKeyData         string `yaml:"client-key-data"`
			} `yaml:"user"`
		} `yaml:"users"`
	}

	ksm.KubeYaml = types.StringValue(string(k8bytes))

	if err := yaml.UnmarshalStrict(k8bytes, &cbkHolder); err != nil {
		d.AddError("Error interpreting k8s context from k0sctl response", err.Error())
		return d
	}

	var contextName, clusterName, userName string

	contextName = cbkHolder.CurrentContext

	for _, context := range cbkHolder.Contexts {
		if context.Name == contextName {
			clusterName = context.Context.Cluster
			userName = context.Context.User
			break
		}
	}

	for _, cluster := range cbkHolder.Clusters {
		if cluster.Name == clusterName {
			ksm.KubeHost = types.StringValue(cluster.Cluster.Server)
			ksm.CaCert = types.StringValue(helperStringBase64Decode(cluster.Cluster.CertificateAuthorityData))
			break
		}
	}

	for _, user := range cbkHolder.Users {
		if user.Name == userName {
			ksm.PrivateKey = types.StringValue(helperStringBase64Decode(user.User.ClientKeyData))
			ksm.ClientCert = types.StringValue(helperStringBase64Decode(user.User.ClientCertificateData))
			break
		}
	}

	return d
}

type k0sctlSchemaClusterMetadata struct {
	Name types.String `tfsdk:"name"`
}

type k0sctlSchemaModelSpec struct {
	Hosts []k0sctlSchemaModelSpecHost `tfsdk:"host"`
	K0s   k0sctlSchemaModelSpecK0s    `tfsdk:"k0s"`
}

type k0sctlSchemaModelSpecK0s struct {
	Version types.String `tfsdk:"version"`
	Config  types.String `tfsdk:"config"`
}

type k0sctlSchemaModelSpecHost struct {
	Role         types.String                     `tfsdk:"role"`
	InstallFlags []types.String                   `tfsdk:"install_flags"`
	Hooks        []k0sctlSchemaModelSpecHostHooks `tfsdk:"hooks"`
	SSH          []k0sctlSchemaModelSpecHostSSH   `tfsdk:"ssh"`
	WinRM        []k0sctlSchemaModelSpecHostWinrm `tfsdk:"winrm"`
}
type k0sctlSchemaModelSpecHostHooks struct {
	Apply []k0sctlSchemaModelSpecHostHookAction `tfsdk:"apply"`
}
type k0sctlSchemaModelSpecHostHookAction struct {
	Before types.List `tfsdk:"before"`
	After  types.List `tfsdk:"after"`
}
type k0sctlSchemaModelSpecHostSSH struct {
	Address types.String `tfsdk:"address"`
	KeyPath types.String `tfsdk:"key_path"`
	User    types.String `tfsdk:"user"`
	Port    types.Int64  `tfsdk:"port"`
}
type k0sctlSchemaModelSpecHostWinrm struct {
	Address  types.String `tfsdk:"address"`
	User     types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
	Port     types.Int64  `tfsdk:"port"`
	UseHTTPS types.Bool   `tfsdk:"use_https"`
	Insecure types.Bool   `tfsdk:"insecure"`
}
