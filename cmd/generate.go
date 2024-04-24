package main

import (
	"context"
	"log"

	api "github.com/openshift-splat-team/vsphere-capacity-manager/pkg/apis/vspherecapacitymanager.splat.io/v1"

	"github.com/openshift-splat-team/vsphere-capacity-manager-data/pkg/ibmcloud"
)

func main() {
	/*var vcenterCreds map[string]vsphere.VCenterCredential
	data, err := os.ReadFile(os.Args[1])

	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(data, &vcenterCreds)

	if err != nil {
		log.Fatal(err)
	}
	metadata := vsphere.NewMetadata()

	for k, v := range vcenterCreds {
		metadata.AddCredentials(k, v.Username, v.Password)
	}
	//portSetting := pgMo.Config.DefaultPortConfig.(*types.VMwareDVSPortSetting)

	*/
	ibmTesting()
	network := api.Network{}
	log.Print(network)
}

func ibmTesting() {
	m := ibmcloud.NewMetadata()

	err := m.AddCredentials("ours", "", "")
	if err != nil {
		log.Fatal(err)
	}

	session, err := m.Session(context.TODO(), "ours")

	if err != nil {
		log.Fatal(err)
	}

	vlans, err := ibmcloud.GetVlanSubnets(session)

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
}
