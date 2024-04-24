package ibmcloud

import "github.com/softlayer/softlayer-go/datatypes"

const (
	vlanSubnetMask = `mask[id,name,vlanNumber,fullyQualifiedName,subnets[id,ipAddressCount,gateway,cidr,netmask,networkIdentifier,subnetType,ipAddresses[ipAddress,isNetwork,isBroadcast,isGateway]],primaryRouter[hostname]]`
)

func GetVlanSubnets(session *SoftlayerSession) ([]datatypes.Network_Vlan, error) {
	vlans, err := session.AccountSession.Mask(vlanSubnetMask).GetNetworkVlans()
	if err != nil {
		return nil, err
	}
	return vlans, nil
}
