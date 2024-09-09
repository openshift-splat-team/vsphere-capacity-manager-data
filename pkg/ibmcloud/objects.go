package ibmcloud

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/softlayer/softlayer-go/datatypes"
)

/*
{
    "broadcastAddress": "",
    "cidr": 64,
    "datacenter": {
        "id": 1854795,
        "locationStatus": {
            "id": 2,
            "status": "Active"
        },
        "longName": "Dallas 12",
        "name": "dal12",
        "statusId": 2
    },
    "gateway": "2607:f0d0:1f01:0022:0000:0000:0000:0001",
    "id": 1388317,
    "ipAddresses": [
        {
            "id": 203902761,
            "ipAddress": "2607:f0d0:1f01:0022:0000:0000:0000:0000"
        },
        {
            "id": 203902763,
            "ipAddress": "2607:f0d0:1f01:0022:0000:0000:0000:0001"
        },
        {
            "id": 203902765,
            "ipAddress": "2607:f0d0:1f01:0022:0000:0000:0000:0002"
        },
        {
            "id": 203902767,
            "ipAddress": "2607:f0d0:1f01:0022:0000:0000:0000:0003"
        },
        {
            "id": 203902769,
            "ipAddress": "2607:f0d0:1f01:0022:0000:0000:0000:0004"
        },
        {
            "id": 203902771,
            "ipAddress": "2607:f0d0:1f01:0022:0000:0000:0000:0005"
        },
        {
            "id": 203902773,
            "ipAddress": "2607:f0d0:1f01:0022:0000:0000:0000:0006"
        },
        {
            "id": 203902545,
            "ipAddress": "2607:f0d0:1f01:0022:ffff:ffff:ffff:ff7e",
            "note": "Reserved for HSRP."
        },
        {
            "id": 203902543,
            "ipAddress": "2607:f0d0:1f01:0022:ffff:ffff:ffff:ff7f",
            "note": "Reserved for HSRP."
        }
    ],
    "isCustomerOwned": false,
    "isCustomerRoutable": true,
    "modifyDate": "2023-08-23T09:43:16-06:00",
    "netmask": "ffff:ffff:ffff:ffff:0000:0000:0000:0000",
    "networkIdentifier": "2607:f0d0:1f01:0022:0000:0000:0000:0000",
    "networkVlan": {
        "accountId": 2626524,
        "fullyQualifiedName": "dal12.fcr01.853",
        "id": 3355087,
        "modifyDate": "2024-07-12T09:14:25-06:00",
        "name": "ipv6-devqe-segment",
        "networkSpace": "PUBLIC",
        "primarySubnetId": 1417493,
        "vlanNumber": 853
    },
    "networkVlanId": 3355087,
    "note": "ipv6 vms",
    "sortOrder": "5",
    "subnetType": "SUBNET_ON_VLAN",
    "tagReferences": [
        {
            "id": 1793811832,
            "resourceTableId": 1388317,
            "tagId": 5277216,
            "tagTypeId": 5,
            "usrRecordId": 10752008
        }
    ],
    "totalIpAddresses": 18446744073709552000,
    "usableIpAddressCount": 18446744073709552000,
    "version": 6
}
*/

const (
	vlanSubnetMask    = `mask[id,name,vlanNumber,podName,fullyQualifiedName,datacenter[name],subnets[ipAddressCount,gateway,cidr,netmask,networkIdentifier,subnetType,ipAddresses[ipAddress]],primaryRouter[hostname],tagReferences[tag[name]]]`
	networkSubnetMask = `mask[id,cidr,gateway,tagReferences[tag[name]],version]`

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

func (m *Metadata) GetSubnetsByTag(account, datacenterName, podName, tag string) (*[]datatypes.Network_Subnet, error) {
	_, err := m.Session(context.TODO(), account)
	if err != nil {
		return nil, err
	}

	if m.sessions[account].SubnetsCache == nil || len(*m.sessions[account].SubnetsCache) == 0 {
		subnets, err := m.sessions[account].AccountSession.Mask(networkSubnetMask).GetSubnets()
		if err != nil {
			return nil, err
		}

		m.sessions[account].SubnetsCache = &subnets
	}

	taggedSubnets := make([]datatypes.Network_Subnet, 0, len(*m.sessions[account].SubnetsCache))

	// I am pretty sure there is a way to do this by the api but /shrug this seems easier at the moment

	for _, s := range *m.sessions[account].SubnetsCache {
		for _, tagRef := range s.TagReferences {
			if strings.Contains(*tagRef.Tag.Name, tag) {
				taggedSubnets = append(taggedSubnets, s)
			}
		}
	}

	return &taggedSubnets, nil
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
