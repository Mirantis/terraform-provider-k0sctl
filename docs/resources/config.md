---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "k0sctl_config Resource - terraform-provider-k0sctl"
subcategory: ""
description: |-
  Mirantis installation using launchpad, parametrized
---

# k0sctl_config (Resource)

Mirantis installation using launchpad, parametrized



<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `disable_downgrade_check` (Boolean) Skip downgrade check
- `force` (Boolean) Attempt a forced installation in case of certain failures
- `kube_skiptlsverify` (Boolean) K8 Kubernetes endpoint TLS should not be verified
- `metadata` (Block, Optional) Metadata for the launchpad cluster (see [below for nested schema](#nestedblock--metadata))
- `no_drain` (Boolean) Do not drain worker nodes when upgrading
- `no_wait` (Boolean) Do not wait for worker nodes to join
- `restore_from` (String) Path to cluster backup archive to restore the state from
- `skip_create` (Boolean) Skip apply on create
- `skip_destroy` (Boolean) Skip reset on destroy
- `spec` (Block, Optional) Launchpad install specifications (see [below for nested schema](#nestedblock--spec))

### Read-Only

- `ca_cert` (String) K8 Server CA certificate
- `client_cert` (String) K8 Client certificate for the user
- `id` (String) Example identifier
- `k0s_yaml` (String) K0S yaml for debugging
- `kube_host` (String) K8 Kubernetes API host endpoint
- `kube_yaml` (String, Sensitive) K8 Kubernetes API client configuration yaml file
- `private_key` (String) K8 Private key for the user

<a id="nestedblock--metadata"></a>
### Nested Schema for `metadata`

Required:

- `name` (String) Cluster name


<a id="nestedblock--spec"></a>
### Nested Schema for `spec`

Optional:

- `host` (Block List) Individual host configuration, for each machine in the cluster (see [below for nested schema](#nestedblock--spec--host))
- `k0s` (Block, Optional) K0S installation configuration (see [below for nested schema](#nestedblock--spec--k0s))

<a id="nestedblock--spec--host"></a>
### Nested Schema for `spec.host`

Required:

- `role` (String) Host machine role in the cluster

Optional:

- `hooks` (Block List) Hook configuration for the host (see [below for nested schema](#nestedblock--spec--host--hooks))
- `hostname` (String) Hostname override for the host
- `install_flags` (List of String) String install flags passed to k0s (e.g. '--taints=mytaint')
- `no_taints` (Boolean) Do not apply taints to the host, used in conjunction with the controller+worker role
- `private_address` (String) Private address override for the host
- `ssh` (Block List) SSH configuration for the host (see [below for nested schema](#nestedblock--spec--host--ssh))
- `winrm` (Block List) WinRM configuration for the host (see [below for nested schema](#nestedblock--spec--host--winrm))

<a id="nestedblock--spec--host--hooks"></a>
### Nested Schema for `spec.host.hooks`

Optional:

- `apply` (Block List) Launchpad.Apply string hooks for the host (see [below for nested schema](#nestedblock--spec--host--hooks--apply))

<a id="nestedblock--spec--host--hooks--apply"></a>
### Nested Schema for `spec.host.hooks.apply`

Optional:

- `after` (List of String) String hooks to run on hosts after the Apply operation is run.
- `before` (List of String) String hooks to run on hosts before the Apply operation is run.



<a id="nestedblock--spec--host--ssh"></a>
### Nested Schema for `spec.host.ssh`

Required:

- `address` (String) SSH endpoint
- `user` (String) SSH endpoint

Optional:

- `bastion` (Block List) SSH bastion configuration for the host (see [below for nested schema](#nestedblock--spec--host--ssh--bastion))
- `key_content` (String) Content of the ssh key
- `key_path` (String) SSH endpoint
- `port` (Number) SSH Port

<a id="nestedblock--spec--host--ssh--bastion"></a>
### Nested Schema for `spec.host.ssh.bastion`

Required:

- `address` (String) bastion endpoint
- `user` (String) bastion endpoint

Optional:

- `key_content` (String) Content of the ssh key for the bastion host
- `key_path` (String) bastion endpoint
- `port` (Number) bastion Port



<a id="nestedblock--spec--host--winrm"></a>
### Nested Schema for `spec.host.winrm`

Required:

- `address` (String) WinRM endpoint
- `password` (String) WinRM password
- `user` (String) WinRM user

Optional:

- `insecure` (Boolean) If false, then no SSL certificate validation is used
- `port` (Number) WinRM Port
- `use_https` (Boolean) If false, then no HTTP is used for winrm transport



<a id="nestedblock--spec--k0s"></a>
### Nested Schema for `spec.k0s`

Required:

- `version` (String) K0s version to install

Optional:

- `config` (String) K0s config yaml as a string
