package vsphere

import (
	"context"
	"fmt"
	"net"

	"github.com/vmware/govmomi/vapi/tags"
	"sigs.k8s.io/cluster-api-provider-vsphere/pkg/session"
)

type ClusterContext struct {
	NumCpuCores int16
	TotalMemory int64
}

type PortGroupContext struct {
	VlanId int16
}

type DatacenterContext struct {
	ClusterContexts   map[string]ClusterContext
	PortGroupContexts map[string]PortGroupContext
}

// VCenterContext maintains context of known vCenters to be used in CAPI manifest reconciliation.
type VCenterContext struct {
	VCenter            string
	DatacenterContexts map[string]DatacenterContext
	IPAddresses        []net.IP
	TagCategories      []tags.Category
}

// VCenterCredential contains the vCenter username and password.
type VCenterCredential struct {
	Username string
	Password string
}

// Metadata holds vcenter stuff.
type Metadata struct {
	sessions    map[string]*session.Session
	credentials map[string]*session.Params

	VCenterContexts map[string]VCenterContext

	VCenterCredentials map[string]VCenterCredential
}

// NewMetadata initializes a new Metadata object.
func NewMetadata() *Metadata {
	return &Metadata{
		sessions:           make(map[string]*session.Session),
		credentials:        make(map[string]*session.Params),
		VCenterContexts:    make(map[string]VCenterContext),
		VCenterCredentials: make(map[string]VCenterCredential),
	}
}

// AddCredentials creates a session param from the vCenter server, username and password
// to the Credentials Map.
func (m *Metadata) AddCredentials(server, username, password string) (*session.Params, error) {
	if _, ok := m.VCenterCredentials[server]; !ok {
		m.VCenterCredentials[server] = VCenterCredential{
			Username: username,
			Password: password,
		}
	}

	// m.credentials is not stored in the json state file - there is no real reason to do this
	// but upon returning to AddCredentials (create manifest, create cluster) the credentials map is
	// nil, re-make it.
	if m.credentials == nil {
		m.credentials = make(map[string]*session.Params)
	}

	if _, ok := m.credentials[server]; !ok {
		m.credentials[server] = session.NewParams().WithServer(server).WithUserInfo(username, password)
	}

	// We need the ip address of vCenter to determine which IBM subnet it is in, then we can determine
	// the physical datacenter and pod vCenter is in.
	// TODO: replace with call to vCenter APIs?
	ipAddrs, err := net.LookupIP(server)
	if err != nil {
		return nil, err
	}

	m.VCenterContexts[server] = VCenterContext{
		VCenter:     server,
		IPAddresses: ipAddrs,
	}

	return m.credentials[server], nil
}

// Session returns a session from unlockedSession based on the server (vCenter URL).
func (m *Metadata) Session(ctx context.Context, server string) (*session.Session, error) {
	// m.sessions is not stored in the json state file - there is no real reason to do this
	// but upon returning to Session (create manifest, create cluster) the sessions map is
	// nil, re-make it.
	if m.sessions == nil {
		m.sessions = make(map[string]*session.Session)
	}

	return m.unlockedSession(ctx, server)
}

func (m *Metadata) unlockedSession(ctx context.Context, server string) (*session.Session, error) {
	var err error
	var ok bool
	var params *session.Params

	if params, ok = m.credentials[server]; !ok {
		if creds, ok := m.VCenterCredentials[server]; ok {
			params, err = m.AddCredentials(server, creds.Username, creds.Password)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("credentials for %s not found", server)
		}
	}

	// if nil we haven't created a session
	if _, ok := m.sessions[server]; ok {
		// is the session still valid? if not re-run GetOrCreate.
		if !m.sessions[server].Valid() {
			m.sessions[server], err = session.GetOrCreate(ctx, params)
			if err != nil {
				return nil, err
			}
		}
		return m.sessions[server], nil
	}

	// If we have gotten here there is no session for the server name, create.
	m.sessions[server], err = session.GetOrCreate(ctx, params)
	return m.sessions[server], err
}
