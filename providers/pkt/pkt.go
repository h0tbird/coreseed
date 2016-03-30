package pkt

//---------------------------------------------------------------------------
// Package factored import statement:
//---------------------------------------------------------------------------

import (

	// Stdlib:
	"fmt"

	// Community:
	"github.com/packethost/packngo"
)

//----------------------------------------------------------------------------
// Typedefs:
//----------------------------------------------------------------------------

// PacketData contains variables used by Packet.net API.
type PacketData struct {
	APIKey   string
	HostName string
	Plan     string
	Facility string
	Osys     string
	Billing  string
	ProjID   string
}

//--------------------------------------------------------------------------
// func: Run
//--------------------------------------------------------------------------

// Run uses Packet.net API to launch a new server.
func (d *PacketData) Run(udata []byte) error {

	// Connect and authenticate to the API endpoint:
	client := packngo.NewClient("", d.APIKey, nil)

	// Forge the request:
	createRequest := &packngo.DeviceCreateRequest{
		HostName:     d.HostName,
		Plan:         d.Plan,
		Facility:     d.Facility,
		OS:           d.Osys,
		BillingCycle: d.Billing,
		ProjectID:    d.ProjID,
		UserData:     string(udata),
	}

	// Send the request:
	newDevice, _, err := client.Devices.Create(createRequest)
	if err != nil {
		return err
	}

	// Pretty-print the response data:
	fmt.Println(newDevice)
	return nil
}
