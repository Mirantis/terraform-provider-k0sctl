package phase

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	k0sctl_phase "github.com/k0sproject/k0sctl/phase"
	"github.com/k0sproject/rig/log"
	"github.com/sirupsen/logrus"
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
	return p.validateHostCounts()
}

func (p *ValidateHostsExtended) validateHostCounts() error {

	tflog.Error(context.Background(), "#################### ValidateHostsExtended phase started ###############################")
	log.Errorf("#################### from rig: ValidateHostsExtended phase started ###############################")
	logrus.Error("#################### from logrus: ValidateHostsExtended phase started ###############################")
	if len(p.Config.Metadata.EtcdMembers) == len(p.Config.Spec.Hosts.Controllers()) {
		return nil
		//return errors.New("##################### test 2 ########################")
	}
	return nil

	//return errors.New("##################### test 3 ########################")
}
