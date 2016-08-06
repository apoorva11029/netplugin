package mastercfg

import (
	"encoding/json"
	"fmt"
	"github.com/contiv/netplugin/core"
)

// CfgTenantState comment
type CfgTenantState struct {
	core.CommonState
	TenantName string `json:"tenantName"`
	NetCount   int    `json:"numNet"`
}

// Write the state.
func (s *CfgTenantState) Write() error {
	key := fmt.Sprintf(tenantConfigPath, s.ID)
	return s.StateDriver.WriteState(key, s, json.Marshal)
}

// Read the state for a given identifier.
func (s *CfgTenantState) Read(id string) error {
	key := fmt.Sprintf(tenantConfigPath, id)
	return s.StateDriver.ReadState(key, s, json.Unmarshal)
}

// ReadAll state and return the collection.
func (s *CfgTenantState) ReadAll() ([]core.State, error) {
	return s.StateDriver.ReadAllState(tenantConfigPathPrefix, s, json.Unmarshal)
}

// WatchAll state transitions and send them through the channel.
func (s *CfgTenantState) WatchAll(rsps chan core.WatchState) error {
	return s.StateDriver.WatchAllState(tenantConfigPathPrefix, s, json.Unmarshal,
		rsps)
}

// Clear removes the state.
func (s *CfgTenantState) Clear() error {
	key := fmt.Sprintf(tenantConfigPath, s.ID)
	return s.StateDriver.ClearState(key)
}
