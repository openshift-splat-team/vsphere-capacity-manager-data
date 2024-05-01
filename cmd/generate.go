package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/openshift-splat-team/vsphere-capacity-manager-data/pkg/asset/generation"
	api "github.com/openshift-splat-team/vsphere-capacity-manager/pkg/apis/vspherecapacitymanager.splat.io/v1"
)

func main() {
	env, err := generation.CreateVSphereEnvironmentsConfig()
	if err != nil {
		log.Fatal(err)
	}

	b, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("output.json", b, 0644)
	if err != nil {
		log.Fatal(err)
	}

	network := api.Network{}
	log.Print(network)
}
