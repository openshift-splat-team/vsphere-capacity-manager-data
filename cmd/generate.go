package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/openshift-splat-team/vsphere-capacity-manager-data/pkg/vsphere"
)

func main() {
	var vcenterCreds map[string]vsphere.VCenterCredential
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
}
