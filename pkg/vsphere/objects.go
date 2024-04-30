package vsphere

import (
	"context"
	"fmt"
	"log"
	"path"
	"strings"
	"time"

	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"sigs.k8s.io/cluster-api-provider-vsphere/pkg/session"

	v1 "github.com/openshift/api/config/v1"
)

const (
	timeout                   = time.Second * 60
	openshiftZoneTagCatName   = "openshift-zone"
	openshiftRegionTagCatName = "openshift-region"
)

func (m *Metadata) FindVCenterVirtualMachine(server string) (*mo.VirtualMachine, error) {
	sess, err := m.Session(context.TODO(), server)
	if err != nil {
		return nil, err
	}

	mgr := view.NewManager(sess.Client.Client)
	kind := []string{"VirtualMachine"}

	v, err := mgr.CreateContainerView(context.TODO(), sess.ServiceContent.RootFolder, kind, true)
	if err != nil {
		return nil, err
	}

	var virtualMachines []mo.VirtualMachine
	err = v.Retrieve(context.TODO(), kind, []string{"config", "guest", "network"}, &virtualMachines)
	if err != nil {
		return nil, err
	}

	for _, vm := range virtualMachines {
		if vm.Guest.HostName == server {
			return &vm, nil
		}
	}

	return nil, nil
}

func (m *Metadata) GetPortGroupVlanFromMoRef(networks []types.ManagedObjectReference, server string) ([]int32, error) {
	sess, err := m.Session(context.TODO(), server)

	if err != nil {
		return nil, err
	}

	var dvpgs []mo.DistributedVirtualPortgroup
	var vlanIds []int32

	err = sess.Retrieve(context.TODO(), networks, []string{"config"}, &dvpgs)
	if err != nil {
		return nil, err
	}
	for _, pg := range dvpgs {
		portSetting := pg.Config.DefaultPortConfig.(*types.VMwareDVSPortSetting)
		vlanIds = append(vlanIds, portSetting.Vlan.(*types.VmwareDistributedVirtualSwitchVlanIdSpec).VlanId)
	}
	return vlanIds, nil
}

func (m *Metadata) GetTagCategories(server string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	sess, err := m.Session(ctx, server)
	if err != nil {
		return err
	}

	// We only want to retrieve the categories once
	if len(m.VCenterContexts[server].TagCategories) > 0 {
		return nil
	}

	vctrCtx := m.VCenterContexts[server]

	vctrCtx.TagCategories, err = sess.TagManager.GetCategories(ctx)
	if err != nil {
		return err
	}

	m.VCenterContexts[server] = vctrCtx
	return nil
}

/*
   "FailureDomains": [
       {
           "name": "us-east-4",
           "vcenter": "vcs8e-vc.ocp2.dev.cluster.com",
           "zone": "us-east-4a",
           "region": "us-east"
       },
       {
           "name": "us-south-1",
           "vcenter": "v8c-2-vcenter.ocp2.dev.cluster.com",
           "zone": "us-south-1a",
           "region": "us-south"
       }
   ],
*/

func (m *Metadata) GetFailureDomainsViaTag(server string) (*[]v1.VSpherePlatformFailureDomainSpec, error) {
	failureDomainMap := make(map[string]v1.VSpherePlatformFailureDomainSpec)

	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	sess, err := m.Session(ctx, server)
	if err != nil {
		return nil, err
	}

	regionCategory, err := sess.TagManager.GetCategory(ctx, openshiftRegionTagCatName)
	if err != nil {
		return nil, err
	}
	zoneCategory, err := sess.TagManager.GetCategory(ctx, openshiftZoneTagCatName)
	if err != nil {
		return nil, err
	}

	// Retrieve all the tags for openshift-region category
	regionTags, err := sess.TagManager.GetTagsForCategory(ctx, regionCategory.ID)
	if err != nil {
		return nil, err
	}
	regionTagIds := make([]string, len(regionTags))
	for i := 0; i < len(regionTags); i++ {
		regionTagIds[i] = regionTags[i].ID
	}

	// Retrieve all the attached objects for openshift-region category
	attachedRegionObjects, err := sess.TagManager.GetAttachedObjectsOnTags(ctx, regionTagIds)
	if err != nil {
		return nil, err
	}

	// Retrieve all the tags for openshift-zone category
	zoneTags, err := sess.TagManager.GetTagsForCategory(ctx, zoneCategory.ID)
	if err != nil {
		return nil, err
	}
	zoneTagIds := make([]string, len(zoneTags))
	for i := 0; i < len(zoneTags); i++ {
		zoneTagIds[i] = zoneTags[i].ID
	}

	// Retrieve all the attached objects for openshift-zone category
	attachedZoneObjects, err := sess.TagManager.GetAttachedObjectsOnTags(ctx, zoneTagIds)
	if err != nil {
		return nil, err
	}

	clusterTagMap := make(map[string]string)

	for _, za := range attachedZoneObjects {
		for _, zaObj := range za.ObjectIDs {
			ref, err := sess.Finder.ObjectReference(ctx, zaObj.Reference())
			if err != nil {
				return nil, err
			}
			// we only care about cluster objects
			if cObj, ok := ref.(object.ClusterComputeResource); ok {
				clusterTagMap[cObj.Reference().Value] = za.Tag.Name
			}
		}
	}

	for _, ra := range attachedRegionObjects {
		for _, raObj := range ra.ObjectIDs {
			ref, err := sess.Finder.ObjectReference(ctx, raObj.Reference())
			if err != nil {
				return nil, err
			}
			// we only care about datacenter objects
			if dcObj, ok := ref.(object.Datacenter); ok {
				clusterObjects, err := sess.Finder.ClusterComputeResourceList(ctx, path.Join(dcObj.InventoryPath, "host", "..."))
				if err != nil {
					return nil, err
				}

				for _, clusterObj := range clusterObjects {

					//var datastore map[string]bool

					var datastore map[types.ManagedObjectReference]bool

					if tagName, ok := clusterTagMap[clusterObj.Reference().Value]; ok {

						var cMo mo.ClusterComputeResource

						if err := clusterObj.Properties(ctx, clusterObj.Reference(), []string{"host,datastore,network"}, &cMo); err != nil {
							return nil, err
						}

						networks := make([]string, len(cMo.Network))

						for _, n := range cMo.Network {
							var objref object.Reference
							if objref, err = sess.Finder.ObjectReference(ctx, n.Reference()); err != nil {
								return nil, err
							}

							networks = append(networks, objref.(object.Network).InventoryPath)
						}

						// datastores available on the cluster
						for _, ds := range cMo.Datastore {
							datastore[ds] = true
						}

						var hMo mo.HostSystem
						// hosts available on the cluster
						for _, h := range cMo.Host {
							if err := sess.PropertyCollector().RetrieveOne(ctx, h.Reference(), []string{"datastore"}, &hMo); err != nil {
								return nil, err
							}

							// datastores available on a host
							for _, ds := range hMo.Datastore {
								// datastore exists on host and cluster
								if _, ok := datastore[ds]; !ok {
									datastore[ds] = false
								} else {
									datastore[ds] = datastore[ds] && true
								}
							}
						}

						datastorePaths := make([]string, len(datastore))
						for k, v := range datastore {
							if v {
								dsref, err := sess.Finder.ObjectReference(ctx, k)
								if err != nil {
									return nil, err
								}

								datastorePaths = append(datastorePaths, dsref.(object.Datastore).InventoryPath)
							}
						}

						joinedDSPaths := strings.Join(datastorePaths, ",")

						clusterName := clusterObj.Name()
						datacenterName := dcObj.Name()

						key := fmt.Sprintf("%s-%s", datacenterName, clusterName)

						failureDomainMap[key] = v1.VSpherePlatformFailureDomainSpec{
							Region: ra.Tag.Name,
							Zone:   tagName,
							Server: server,
							Topology: v1.VSpherePlatformTopology{
								Datacenter:     dcObj.InventoryPath,
								ComputeCluster: clusterObj.InventoryPath,
								Networks:       networks,

								Datastore:    joinedDSPaths,
								Folder:       "",
								Template:     "",
								ResourcePool: "",
							},
						}
					}
				}
			}
		}
	}

	return nil, nil
}

func (m *Metadata) GetTopologyByTags(server string, objectID []mo.Reference) error {
	var openshiftZoneTagCatId string
	var openshiftRegionTagCatId string

	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	sess, err := m.Session(ctx, server)
	if err != nil {
		return err
	}

	err = m.GetTagCategories(server)
	if err != nil {
		return err
	}

	attachedTags, err := sess.TagManager.GetAttachedTagsOnObjects(ctx, objectID)
	if err != nil {
		return err
	}

	for _, tc := range m.VCenterContexts[server].TagCategories {
		if tc.Name == openshiftZoneTagCatName {
			openshiftZoneTagCatId = tc.ID
		}
		if tc.Name == openshiftRegionTagCatName {
			openshiftRegionTagCatId = tc.ID
		}
	}

	for _, atag := range attachedTags {
		for _, tag := range atag.Tags {
			if tag.CategoryID == openshiftZoneTagCatId {

				log.Print(tag.Name)
				log.Print(atag.ObjectID)

			}
			if tag.CategoryID == openshiftRegionTagCatId {

			}
		}
	}

	return nil
}

func (m *Metadata) GetHostnameUrlVpxd(server string) (*string, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()
	sess, err := m.Session(ctx, server)
	if err != nil {
		return nil, err
	}

	optmgr := object.NewOptionManager(sess.Client.Client, *sess.ServiceContent.Setting)

	baseOptionValue, err := optmgr.Query(context.TODO(), "config.vpxd.hostnameUrl")
	if err != nil {
		return nil, err
	}

	url := baseOptionValue[0].GetOptionValue().Value.(string)

	return &url, nil
}

func (m *Metadata) GetDistributedPortGroups(server, portGroupSubString string) ([]mo.DistributedVirtualPortgroup, error) {
	var portGroupManagedObjects []mo.DistributedVirtualPortgroup
	sess, err := m.Session(context.TODO(), server)
	if err != nil {
		return nil, err
	}

	mgr := view.NewManager(sess.Client.Client)
	kind := []string{"DistributedVirtualPortgroup"}

	v, err := mgr.CreateContainerView(context.TODO(), sess.ServiceContent.RootFolder, kind, true)
	if err != nil {
		return nil, err
	}

	err = v.Retrieve(context.TODO(), kind, []string{"config"}, portGroupManagedObjects)
	if err != nil {
		return nil, err
	}

	var portGroups []mo.DistributedVirtualPortgroup
	for _, pg := range portGroupManagedObjects {
		if strings.Contains(pg.Name, portGroupSubString) {
			portGroups = append(portGroups, pg)
		}
	}

	return portGroups, nil
}
func (m *Metadata) GetPortGroups(server string, datacenter *object.Datacenter) ([]*mo.DistributedVirtualPortgroup, error) {
	var err error
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	sess, err := m.Session(ctx, server)
	if err != nil {
		return nil, err
	}

	var networks []object.NetworkReference
	var portGroupManagedObjects []*mo.DistributedVirtualPortgroup

	networks, err = sess.Finder.NetworkList(ctx, path.Join(datacenter.InventoryPath, "networks", "..."))
	if err != nil {
		return nil, err
	}

	for _, n := range networks {

		switch n.Reference().Type {
		// We only care about dvPGs
		case "DistributedVirtualPortgroup":
			var pgMo mo.DistributedVirtualPortgroup
			err = n.(object.DistributedVirtualPortgroup).Properties(ctx, n.Reference(), []string{"config"}, &pgMo)
			if err != nil {
				return nil, err
			}
			portGroupManagedObjects = append(portGroupManagedObjects, &pgMo)
		}
	}

	return portGroupManagedObjects, nil
}
func (m *Metadata) GetDatacenters(server string) ([]*object.Datacenter, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	sess, err := m.Session(ctx, server)
	if err != nil {
		return nil, err
	}

	return sess.Finder.DatacenterList(ctx, "/...")
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
