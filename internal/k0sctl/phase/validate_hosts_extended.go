package phase

import (
	"fmt"
	"strings"

	k0sctl_phase "github.com/k0sproject/k0sctl/phase"

	k0sctl_cluster "github.com/k0sproject/k0sctl/pkg/apis/k0sctl.k0sproject.io/v1beta1/cluster"
	"github.com/k0sproject/rig/exec"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

type ValidateHostsExtended struct {
	k0sctl_phase.GenericPhase
}

// Title for the phase.
func (p *ValidateHostsExtended) Title() string {
	return "Validate hosts"
}

// Run the phase.
func (p *ValidateHostsExtended) Run() error {
	return p.validateWorkerCount()
}

func (p *ValidateHostsExtended) validateWorkerCount() error {

	configWorkerMachineIDs, err := p.getWorkerMachineIDs()
	if err != nil {
		return err
	}

	logrus.Debugf("Machine IDs of the hosts in the configuration: %s", configWorkerMachineIDs)

	nodeNames, leader, err := p.getNodeNamesAndLeader()
	if err != nil {
		return err
	}

	logrus.Debugf("Node names: %s", nodeNames)

	for _, node := range nodeNames {
		err := p.validateAndDeleteNode(node, configWorkerMachineIDs, leader)
		if err != nil {
			logrus.Errorf("Error occurred while validating and deleting node: %s", err)
		}
	}

	logrus.Debug("ValidateHostsExtended phase ran successfully")
	return nil
}

func (p *ValidateHostsExtended) getWorkerMachineIDs() ([]string, error) {
	var configWorkerMachineIDs []string

	for _, h := range p.Config.Spec.Hosts.Workers() {
		id, err := h.Configurer.MachineID(h)
		if err != nil {
			return nil, err
		}
		configWorkerMachineIDs = append(configWorkerMachineIDs, id)
	}

	return configWorkerMachineIDs, nil
}

func (p *ValidateHostsExtended) getNodeNamesAndLeader() ([]string, *k0sctl_cluster.Host, error) {
	var nodeNames []string
	var leader *k0sctl_cluster.Host

	for i, h := range p.Config.Spec.Hosts.Controllers() {
		output, err := h.ExecOutput(h.Configurer.KubectlCmdf(h, h.K0sDataDir(), "get nodes -o custom-columns=NAME:.metadata.name --no-headers"), exec.Sudo(h))
		if err != nil {
			if i < len(p.Config.Spec.Hosts.Controllers()) {
				continue
			} else {
				logrus.Errorf("Could not retrieve the list of worker nodes. Error: %s", err)
				return nil, nil, err
			}
		}
		nodeNames = strings.Fields(output)
		leader = h
		break
	}

	return nodeNames, leader, nil
}

func (p *ValidateHostsExtended) validateAndDeleteNode(node string, configWorkerMachineIDs []string, leader *k0sctl_cluster.Host) error {
	machineID, err := leader.ExecOutput(leader.Configurer.KubectlCmdf(leader, leader.K0sDataDir(), fmt.Sprintf("describe node %s | grep -i 'Machine ID:' | awk '{print $3}'", node)), exec.Sudo(leader))
	if err != nil {
		return err
	}

	if !slices.Contains(configWorkerMachineIDs, machineID) {
		output, err := leader.ExecOutput(leader.Configurer.KubectlCmdf(leader, leader.K0sDataDir(), fmt.Sprintf("delete node %s", node)), exec.Sudo(leader))
		if err != nil {
			logrus.Errorf("Error occurred while deleting node: %s. Kubectl output: %s. Error: %s", node, output, err)
			return err
		}
		logrus.Debugf("Node %s successfully deleted", node)
	}

	return nil
}
