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

// Data contains variables used by Packet.net API.
type Data struct {
	APIKey    string
	HostName  string
	ProjectID string
	Plan      string
	OS        string
	Facility  string
	Billing   string
}

//--------------------------------------------------------------------------
// func: Deploy
//--------------------------------------------------------------------------

// Deploy Kato's infrastructure on Packet.net
func (d *Data) Deploy() error {
	return nil
}

//--------------------------------------------------------------------------
// func: Setup
//--------------------------------------------------------------------------

// Setup a Packet.net project to be used by katoctl.
func (d *Data) Setup() error {
	return nil
}

//--------------------------------------------------------------------------
// func: Run
//--------------------------------------------------------------------------

// Run uses Packet.net API to launch a new server.
func (d *Data) Run(udata []byte) error {

	// Connect and authenticate to the API endpoint:
	client := packngo.NewClient("", d.APIKey, nil)

	// Forge the request:
	createRequest := &packngo.DeviceCreateRequest{
		HostName:     d.HostName,
		Plan:         d.Plan,
		Facility:     d.Facility,
		OS:           d.OS,
		BillingCycle: d.Billing,
		ProjectID:    d.ProjectID,
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
