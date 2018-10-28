package nsmvpp

import (
	"fmt"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/networkservicemesh/pkg/nsm/apis/common"
)

type UnimplementedMechanism struct {
	Type common.LocalMechanismType
}

// CreateLocalConnect return error for unimplemented mechanism
func (m UnimplementedMechanism) CreateLocalConnect(apiCh govppapi.Channel, dst Mechanism) (string, error) {
	return "", fmt.Errorf("%s mechanism not implemented", common.LocalMechanismType_name[int32(m.Type)])
}

func (m UnimplementedMechanism) DeleteLocalConnect(apiCh govppapi.Channel, dst Mechanism) error {
	return fmt.Errorf("%s mechanism not implemented", common.LocalMechanismType_name[int32(m.Type)])
}

func (m UnimplementedMechanism) Validate() error {
	return nil
}

func (m UnimplementedMechanism) CreateVppInterface(apiCh govppapi.Channel) (uint32, error) {
	return 0, nil
}

func (m UnimplementedMechanism) DeleteVppInterface(apiCh govppapi.Channel) error {
	return nil
}

func (m UnimplementedMechanism) GetSwIfIndex() uint32 {
	return 0
}

func (m UnimplementedMechanism) GetParameters() map[string]string {
	return nil
}

func NewUnimplementedMechanism(t common.LocalMechanismType) Mechanism {
	return UnimplementedMechanism{Type: t}
}
