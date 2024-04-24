package ibmcloud

import (
	"context"
	"fmt"
	"net"

	"github.com/softlayer/softlayer-go/datatypes"
)

const (
	vlanSubnetMask = `mask[id,name,vlanNumber,podName,fullyQualifiedName,datacenter[name],subnets[id,ipAddressCount,gateway,cidr,netmask,networkIdentifier,subnetType,ipAddresses[ipAddress,isNetwork,isBroadcast,isGateway]],primaryRouter[hostname]]`
)

type VCenterLocation struct {
	DatacenterName        *string
	PodName               *string
	PrimaryRouterHostname *string
	VlanNumber            *int

	IPAddress net.IP
}

func (m *Metadata) GetVlanSubnets(account string) ([]datatypes.Network_Vlan, error) {
	_, err := m.Session(context.TODO(), account)
	if err != nil {
		return nil, err
	}
	vlans, err := m.sessions[account].AccountSession.Mask(vlanSubnetMask).GetNetworkVlans()
	if err != nil {
		return nil, err
	}
	return vlans, nil
}

func (m *Metadata) FindVCenterPhyDC(account string, vCenterIPAddresses []net.IP) (*VCenterLocation, error) {
	var vcloc VCenterLocation

	_, err := m.Session(context.TODO(), account)
	if err != nil {
		return nil, err
	}

	vlans, err := m.GetVlanSubnets(account)
	if err != nil {
		return nil, err
	}

vlanloop:
	for _, v := range vlans {
		for _, s := range v.Subnets {
			ip := fmt.Sprintf("%s/%d", *s.NetworkIdentifier, *s.Cidr)

			_, ipNet, err := net.ParseCIDR(ip)
			if err != nil {
				return nil, err
			}

			for _, vcIP := range vCenterIPAddresses {
				if ipNet.Contains(vcIP) {
					vcloc.PrimaryRouterHostname = v.PrimaryRouter.Hostname
					vcloc.PodName = v.PodName
					vcloc.DatacenterName = v.Datacenter.Name
					vcloc.IPAddress = vcIP
					vcloc.VlanNumber = v.VlanNumber
					break vlanloop

				}
			}
		}
	}

	return &vcloc, nil
}
