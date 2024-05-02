package main

import (
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/openshift-splat-team/vsphere-capacity-manager-data/cmd"
)

func main() {
	ctrl.SetLogger(klog.Background())
	cmd.Execute()
}
