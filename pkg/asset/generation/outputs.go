package generation

import (
	"log"
	"net"

	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/openshift-splat-team/vsphere-capacity-manager-data/pkg/ibmcloud"
	"github.com/openshift-splat-team/vsphere-capacity-manager-data/pkg/vsphere"
	configv1 "github.com/openshift/api/config/v1"
)

type PortGroupSubnet struct {
	// IBM and Port Group vlan id
	VlanId int32
	// Port Group name
	Name string

	// IBM Variables
	PodName        *string
	DatacenterName *string
	Subnets        []datatypes.Network_Subnet

	// vcenter server
	Server string
}

type FailureDomainResourceCapacity struct {
	NumCpuCores int16
	TotalMemory int64
	Name        string
}

type VSphereEnvironmentsConfig struct {
	configv1.VSpherePlatformSpec
	PortGroupSubnets               []PortGroupSubnet
	FailureDomainsResourceCapacity []FailureDomainResourceCapacity
}

func CreateVSphereEnvironmentsConfig() (*VSphereEnvironmentsConfig, error) {
	var envs VSphereEnvironmentsConfig

	server := "vcs8e-vc.ocp2.dev.cluster.com"
	vmeta := vsphere.NewMetadata()

	vcenterCreds := make(map[string]vsphere.VCenterCredential)
	vcenterCreds[server] = vsphere.VCenterCredential{
		Username: "",
		Password: "",
	}

	account := "foo"

	imeta := ibmcloud.NewMetadata()
	err := imeta.AddCredentials(account, "", "")
	if err != nil {
		return nil, err
	}

	for k, v := range vcenterCreds {
		_, err := vmeta.AddCredentials(k, v.Username, v.Password)
		if err != nil {
			log.Fatal(err)
		}

		failureDomains, err := vmeta.GetFailureDomainsViaTag(k)
		if err != nil {
			return nil, err
		}

		for _, fd := range *failureDomains {
			cObj, err := vmeta.GetClusterByPath(fd.Server, fd.Topology.ComputeCluster)
			if err != nil {
				return nil, err
			}

			cpu, memory, err := vmeta.GetClusterCapacity(fd.Server, cObj)
			if err != nil {
				return nil, err
			}

			envs.FailureDomainsResourceCapacity = append(envs.FailureDomainsResourceCapacity, FailureDomainResourceCapacity{
				Name:        fd.Name,
				NumCpuCores: cpu,
				TotalMemory: memory,
			})
		}

		envs.FailureDomains = append(envs.FailureDomains, *failureDomains...)

		var dcPaths []string
		datacenters, err := vmeta.GetDatacenters(k)
		if err != nil {
			return nil, err
		}

		for _, dc := range datacenters {
			dcPaths = append(dcPaths, dc.InventoryPath)
		}

		envs.VCenters = append(envs.VCenters, configv1.VSpherePlatformVCenterSpec{
			Server:      k,
			Datacenters: dcPaths,
		})

		portGroups, err := vmeta.GetDistributedPortGroups(k, "ci-vlan")
		if err != nil {
			return nil, err
		}

		portGroupSubnetsMap := make(map[int32]PortGroupSubnet)

		for _, pg := range portGroups {
			portSetting := pg.Config.DefaultPortConfig.(*types.VMwareDVSPortSetting)
			vlanId := portSetting.Vlan.(*types.VmwareDistributedVirtualSwitchVlanIdSpec).VlanId

			portGroupSubnetsMap[vlanId] = PortGroupSubnet{
				Name:   pg.Config.Name,
				VlanId: vlanId,
				Server: k,
			}
		}

		url, err := vmeta.GetHostnameUrlVpxd(k)
		if err != nil {
			return nil, err
		}

		if k != *url {
			log.Printf("WARN: vCenter URL does not match %s != %s", k, *url)
		}

		vcIP, err := net.LookupIP(k)
		if err != nil {
			log.Fatal(err)
		}

		location, err := imeta.FindVCenterPhyDC(account, vcIP)
		if err != nil {
			return nil, err
		}

		networkVlans, err := imeta.GetVlanSubnets(account, *location.DatacenterName, *location.PodName)
		if err != nil {
			return nil, err
		}

		for _, nv := range *networkVlans {
			vlanNumber := int32(*nv.VlanNumber)
			if _, ok := portGroupSubnetsMap[vlanNumber]; ok {

				pg := PortGroupSubnet{
					VlanId:         vlanNumber,
					Name:           portGroupSubnetsMap[vlanNumber].Name,
					Subnets:        nv.Subnets,
					Server:         k,
					PodName:        location.PodName,
					DatacenterName: location.DatacenterName,
				}

				envs.PortGroupSubnets = append(envs.PortGroupSubnets, pg)
			}
		}
	}

	return &envs, nil
}
