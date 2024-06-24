package phase

import (
	k0sctl_phase "github.com/k0sproject/k0sctl/phase"
	"github.com/k0sproject/rig/log"
)

type ValidateHostsExtended struct {
	k0sctl_phase.GenericPhase
}

// Title for the phase
func (p *ValidateHostsExtended) Title() string {
	return "Validate hosts"
}

// Run the phase
func (p *ValidateHostsExtended) Run() error {

	if len(p.Config.Metadata.EtcdMembers) == len(p.Config.Spec.Hosts.Controllers()) {
		log.Debugf("etcd members in the cluster and controllers listed in the k0sctl configuration are equal")
	}

	return nil
}
