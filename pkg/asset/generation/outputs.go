package generation

import (
	"log"

	"github.com/softlayer/softlayer-go/datatypes"

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
	Subnets        datatypes.Network_Subnet

	// vcenter server
	Server string
}

type VSphereEnvironmentsConfig struct {
	configv1.VSpherePlatformSpec
	PortGroupSubnets []PortGroupSubnet
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
			log.Print(fd.Region)
			log.Print(fd.Zone)
			log.Print(fd.Topology)
		}

		/*

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

			portGroups, err := vmeta.GetDistributedPortGroups(k, "ci-")
			if err != nil {
				return nil, err
			}

			portGroupSubnetsMap := make(map[int32]PortGroupSubnet)

			for _, pg := range portGroups {
				portSetting := pg.Config.DefaultPortConfig.(*types.VMwareDVSPortSetting)
				vlanId := portSetting.Vlan.(*types.VmwareDistributedVirtualSwitchVlanIdSpec).VlanId

				portGroupSubnetsMap[vlanId] = PortGroupSubnet{
					Name:   pg.Name,
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

			log.Printf("vCenter: %s, IP: %s, DC: %s, Pod: %s, Router Hostname: %s",
				k,
				location.IPAddress.String(),
				*location.DatacenterName,
				*location.PodName,
				*location.PrimaryRouterHostname)

			networkVlans, err := imeta.GetVlanSubnets(account, *location.DatacenterName, *location.PodName)
			if err != nil {
				return nil, err
			}

			for _, nv := range *networkVlans {
				vlanNumber := int32(*nv.VlanNumber)
				if _, ok := portGroupSubnetsMap[vlanNumber]; ok {
					envs.PortGroupSubnets = append(envs.PortGroupSubnets, portGroupSubnetsMap[vlanNumber])
				}
			}

			// todo: define failure domains...

			// region-zone pair -> datacenter, cluster, all multi host accessible datastores
			// resource pool?

		*/
	}

	/*
		envs.VCenters[0].Datacenters
		envs.VCenters[0].Server

		envs.FailureDomains[0].Server
		envs.FailureDomains[0].Name
		envs.FailureDomains[0].Region
		envs.FailureDomains[0].Zone
		envs.FailureDomains[0].Topology.Datacenter
		envs.FailureDomains[0].Topology.Datastore
		envs.FailureDomains[0].Topology.Folder
		envs.FailureDomains[0].Topology.ComputeCluster
		envs.FailureDomains[0].Topology.Networks
		envs.FailureDomains[0].Topology.ResourcePool
		envs.FailureDomains[0].Topology.Template

	*/

	return &envs, nil
}
