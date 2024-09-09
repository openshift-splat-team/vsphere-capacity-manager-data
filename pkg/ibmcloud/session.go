package ibmcloud

import (
	"context"
	"fmt"

	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
)

type SoftlayerCredentials struct {
	Username string
	ApiToken string
}

type SoftlayerSession struct {
	Session        *session.Session
	AccountSession services.Account

	NetworkVlansCache *[]datatypes.Network_Vlan
	SubnetsCache      *[]datatypes.Network_Subnet
}

type Metadata struct {
	sessions    map[string]*SoftlayerSession
	credentials map[string]*SoftlayerCredentials
}

func NewMetadata() *Metadata {
	return &Metadata{
		sessions:    make(map[string]*SoftlayerSession),
		credentials: make(map[string]*SoftlayerCredentials),
	}
}

func (m *Metadata) AddCredentials(account, username, apiToken string) error {
	if _, ok := m.credentials[account]; !ok {
		m.credentials[account] = &SoftlayerCredentials{
			Username: username,
			ApiToken: apiToken,
		}
	}
	return nil
}
func (m *Metadata) Session(ctx context.Context, account string) (*SoftlayerSession, error) {
	// m.sessions is not stored in the json state file - there is no real reason to do this
	// but upon returning to Session (create manifest, create cluster) the sessions map is
	// nil, re-make it.
	if m.sessions == nil {
		m.sessions = make(map[string]*SoftlayerSession)
	}

	return m.unlockedSession(ctx, account)
}

func (m *Metadata) unlockedSession(ctx context.Context, account string) (*SoftlayerSession, error) {
	var err error

	// if nil we haven't created a session
	if _, ok := m.sessions[account]; ok {
		// is the session still valid? if not re-run GetOrCreate.

		if m.sessions[account].Session != nil {
			if _, err := m.sessions[account].AccountSession.GetCurrentUser(); err != nil {
				m.sessions[account].Session = session.New(m.credentials[account].Username, m.credentials[account].ApiToken)
				m.sessions[account].AccountSession = services.GetAccountService(m.sessions[account].Session)
				if m.sessions[account].Session == nil {
					return nil, fmt.Errorf("error getting session for account %s", account)
				}
			}
			return m.sessions[account], nil
		}
	}

	// If we have gotten here there is no session for the server name, create.
	tempSession := session.New(m.credentials[account].Username, m.credentials[account].ApiToken)
	tempAccountSession := services.GetAccountService(tempSession)
	if tempSession == nil {
		return nil, fmt.Errorf("error getting session for account %s", account)
	}
	m.sessions[account] = &SoftlayerSession{
		Session:        tempSession,
		AccountSession: tempAccountSession,
	}

	return m.sessions[account], err
}
