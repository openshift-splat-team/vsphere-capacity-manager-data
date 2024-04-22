package vsphere

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"sigs.k8s.io/cluster-api-provider-vsphere/pkg/session"
)

const timeout = time.Second * 60

func GetPortGroups(sess *session.Session, datacenter *object.Datacenter) ([]object.NetworkReference, error) {
	var networks []object.NetworkReference
	var err error
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	networks, err = sess.Finder.NetworkList(ctx, path.Join(datacenter.InventoryPath, "networks", "..."))
	if err != nil {
		return nil, err
	}

	for _, n := range networks {
		var pgMo mo.DistributedVirtualPortgroup
		err = datacenter.Properties(ctx, n.Reference(), []string{"*"}, &pgMo)
		if err != nil {
			return nil, err
		}

	}

	return networks, nil
}
func GetDatacenters(sess *session.Session) ([]*object.Datacenter, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()
	datacenters, err := sess.Finder.DatacenterList(ctx, "/...")

	if err != nil {
		return nil, err
	}

	return datacenters, nil
}

func GetClusters(sess *session.Session, datacenter *object.Datacenter) ([]*object.ClusterComputeResource, error) {
	var clusters []*object.ClusterComputeResource
	var err error
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	clusters, err = sess.Finder.ClusterComputeResourceList(ctx, path.Join(datacenter.InventoryPath, "..."))

	if err != nil {
		return nil, err
	}
	return clusters, nil
}

func GetClusterCapacity(sess *session.Session, cluster *object.ClusterComputeResource) (int16, int64, error) {
	var computeResource mo.ClusterComputeResource

	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	if err := cluster.Properties(ctx, cluster.Reference(), []string{"summary"}, &computeResource); err != nil {
		return 0, 0, err
	}

	if summary := computeResource.Summary.GetComputeResourceSummary(); summary != nil {
		return summary.NumCpuCores, summary.TotalMemory, nil
	}

	return 0, 0, fmt.Errorf("unable to get cluster summary")
}

func GetDatastores(sess *session.Session, datacenter *object.Datacenter) ([]*object.Datastore, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	datastores, err := sess.Finder.DatastoreList(ctx, path.Join(datacenter.InventoryPath, "datastores", "..."))
	if err != nil {
		return nil, err
	}
	return datastores, nil

}
