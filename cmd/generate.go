package main

import (
	"context"
	"github.com/openshift-splat-team/vsphere-capacity-manager-data/pkg/vsphere"
	api "github.com/openshift-splat-team/vsphere-capacity-manager/pkg/apis/vspherecapacitymanager.splat.io/v1"
	"log"
	"net"

	"github.com/openshift-splat-team/vsphere-capacity-manager-data/pkg/ibmcloud"
)

func main() {

	vcenterCreds := make(map[string]vsphere.VCenterCredential)
	/*
		data, err := os.ReadFile(os.Args[1])

		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(data, &vcenterCreds)

		if err != nil {
			log.Fatal(err)
		}
	*/
	metadata := vsphere.NewMetadata()

	vcenterCreds["vcs8e-vc.ocp2.dev.cluster.com"] = vsphere.VCenterCredential{
		Username: "",
		Password: "",
	}

	for k, v := range vcenterCreds {
		_, err := metadata.AddCredentials(k, v.Username, v.Password)
		if err != nil {
			log.Fatal(err)
		}
	}
	server := "vcs8e-vc.ocp2.dev.cluster.com"

	session, err := metadata.Session(context.TODO(), server)

	if err != nil {
		log.Fatal(err)
	}

	dcs, err := vsphere.GetDatacenters(session)
	if err != nil {
		log.Fatal(err)
	}

	for _, dc := range dcs {
		log.Print(dc.Name())
	}

	url, err := metadata.GetHostnameUrlVpxd(server)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(*url)

	if server != *url {
		log.Fatal("these should not be different")

	}

	vm, err := metadata.FindVCenterVirtualMachine(server)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(vm.Guest.HostName)
	log.Print(vm.Network)

	// notes: if vlan ids are 0 assume using native vlan
	// and ignore
	vlans, err := metadata.GetPortGroupVlanFromMoRef(vm.Network, server)
	log.Print(vlans)

	network := api.Network{}
	log.Print(network)
}

func ibmTesting() {
	m := ibmcloud.NewMetadata()

	err := m.AddCredentials("ours", "", "")
	if err != nil {
		log.Fatal(err)
	}

	vcIP, err := net.LookupIP("vcenter.ibmc.devcluster.openshift.com")

	if err != nil {
		log.Fatal(err)
	}

	location, err := m.FindVCenterPhyDC("ours", vcIP)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(location.IPAddress.String())
	log.Print(*location.PrimaryRouterHostname)
	log.Print(*location.PodName)
	log.Print(*location.DatacenterName)
	log.Print(*location.VlanNumber)
	/*
		vlans, err := m.GetVlanSubnets("ours")

		if err != nil {
			log.Fatal(err)
		}

		for _, v := range vlans {
			log.Print(v.Id)

			for _, s := range v.Subnets {
				log.Print(s.Cidr)
				for _, ip := range s.IpAddresses {
					log.Print(ip.IpAddress)
				}
			}
		}

	*/
}
