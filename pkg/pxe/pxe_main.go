package pxe

//-----------------------------------------------------------------------------
// Typedefs:
//-----------------------------------------------------------------------------

// State data.
type State struct {
	Quadruplets []string `json:"-"`
	QuorumCount int      `json:"QuorumCount"`
	MasterCount int      `json:"MasterCount"`
	DNSProvider string   `json:"DNSProvider"`
	DNSApiKey   string   `json:"DNSApiKey"`
	Domain      string   `json:"Domain"`
	EtcdToken   string   `json:"EtcdToken"`
}

// Data struct for PXE instance and state data.
type Data struct {
	command string
	State
}
