package ibmcloud

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/softlayer/softlayer-go/datatypes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	vcmv1 "github.com/openshift-splat-team/vsphere-capacity-manager/pkg/apis/vspherecapacitymanager.splat.io/v1"
)

const (
	vlanSubnetMask = `mask[id,name,vlanNumber,podName,fullyQualifiedName,datacenter[name],subnets[ipAddressCount,gateway,cidr,netmask,networkIdentifier,subnetType,ipAddresses[ipAddress]],primaryRouter[hostname]]`

	//backup copy before removal of parameters that maybe we don't need to make the config more readable
	//vlanSubnetMask = `mask[id,name,vlanNumber,podName,fullyQualifiedName,datacenter[name],subnets[id,ipAddressCount,gateway,cidr,netmask,networkIdentifier,subnetType,ipAddresses[ipAddress,isNetwork,isBroadcast,isGateway]],primaryRouter[hostname]]`
)

type VCenterLocation struct {
	DatacenterName        *string
	PodName               *string
	PrimaryRouterHostname *string
	VlanNumber            *int

	IPAddress net.IP
}

func (m *Metadata) GetVlanSubnets(account, datacenterName, podName string) (*[]datatypes.Network_Vlan, error) {
	_, err := m.Session(context.TODO(), account)
	if err != nil {
		return nil, err
	}

	if m.sessions[account].NetworkVlansCache == nil || len(*m.sessions[account].NetworkVlansCache) == 0 {
		vlans, err := m.sessions[account].AccountSession.Mask(vlanSubnetMask).GetNetworkVlans()
		if err != nil {
			return nil, err
		}

		m.sessions[account].NetworkVlansCache = &vlans
	}

	subsetNetworkVlans := make([]datatypes.Network_Vlan, 0, len(*m.sessions[account].NetworkVlansCache))

	if datacenterName != "" && podName != "" {
		for _, v := range *m.sessions[account].NetworkVlansCache {
			if *v.Datacenter.Name == datacenterName && *v.PodName == podName {
				// ** NOTE ** removing all but the first 20
				maxIpAddresses := uint(20)
				for i, subnet := range v.Subnets {
					if *subnet.IpAddressCount < maxIpAddresses {
						maxIpAddresses = *subnet.IpAddressCount
					}

					v.Subnets[i].IpAddresses = v.Subnets[i].IpAddresses[:maxIpAddresses]
				}

				subsetNetworkVlans = append(subsetNetworkVlans, v)
			}
		}
		return &subsetNetworkVlans, nil
	}

	return m.sessions[account].NetworkVlansCache, nil
}

func (m *Metadata) FindVCenterPhyDC(account string, vCenterIPAddresses []net.IP) (*VCenterLocation, error) {
	var vcloc VCenterLocation

	_, err := m.Session(context.TODO(), account)
	if err != nil {
		return nil, err
	}

	vlans, err := m.GetVlanSubnets(account, "", "")
	if err != nil {
		return nil, err
	}

vlanloop:
	for _, v := range *vlans {
		for _, s := range v.Subnets {

			subnet := vcmv1.Subnet{
				ObjectMeta: metav1.ObjectMeta{
					Name: string(*v.Id),
				},
			}
			log.Print(subnet)

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
