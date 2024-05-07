package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/openshift-splat-team/vsphere-capacity-manager-data/pkg/asset/generation"
)

var rootCmd = &cobra.Command{
	Use:   "vcmd",
	Short: "vcmd is a CLI tool for managing data integration between vSphere and IBM Cloud",
	Args:  cobra.MinimumNArgs(1),
	Run:   func(cmd *cobra.Command, args []string) {},
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Failure Domains, Capacity data and IBM Cloud subnets",
	Run: func(cmd *cobra.Command, args []string) {
		env, err := generation.CreateVSphereEnvironmentsConfig(VCenterAuthFileName, IBMCloudAuthFileName, IPv6Subnet, PortGroupNameSubstring)
		if err != nil {
			log.Fatal(err)
		}

		b, err := json.MarshalIndent(env, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		err = os.WriteFile(OutputFileName, b, 0644)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var VCenterAuthFileName string
var IBMCloudAuthFileName string
var OutputFileName string
var IPv6Subnet string
var PortGroupNameSubstring string

func init() {
	generateCmd.Flags().StringVarP(&VCenterAuthFileName, "vcenter", "v", "vcenter.json", "vCenter JSON Auth File")
	generateCmd.Flags().StringVarP(&IBMCloudAuthFileName, "ibmcloud", "i", "ibmcloud.json", "vCenter JSON Auth File")
	generateCmd.Flags().StringVarP(&OutputFileName, "output", "o", "output.json", "Output file")
	generateCmd.Flags().StringVarP(&IPv6Subnet, "subnet6", "6", "fd65:a1a8:60ad", "IPv6 Subnet defaults to fd65:a1a8:60ad")
	generateCmd.Flags().StringVarP(&PortGroupNameSubstring, "pg", "p", "ci-vlan-", "Port Group substring defaults to ci-vlan-")

	rootCmd.AddCommand(generateCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
