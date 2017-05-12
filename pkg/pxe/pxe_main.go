package pxe

//-----------------------------------------------------------------------------
// Typedefs:
//-----------------------------------------------------------------------------

// State data.
type State struct {
	Quadruplets []string `json:"-"`
	QuorumCount int      `json:"QuorumCount"`
	MasterCount int      `json:"MasterCount"`
}

// Data struct for PXE instance and state data.
type Data struct {
	command string
	State
}
