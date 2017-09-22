package pkt

//---------------------------------------------------------------------------
// Package factored import statement:
//---------------------------------------------------------------------------

import (

	// Stdlib:
	"io/ioutil"
	"os"

	// Community:
	log "github.com/Sirupsen/logrus"
	"github.com/packethost/packngo"
)

//----------------------------------------------------------------------------
// Typedefs:
//----------------------------------------------------------------------------

// Data contains variables used by Packet.net API.
type Data struct {
	command   string
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
func (d *Data) Deploy() {
}

//--------------------------------------------------------------------------
// func: Setup
//--------------------------------------------------------------------------

// Setup a Packet.net project to be used by katoctl.
func (d *Data) Setup() {
}

//--------------------------------------------------------------------------
// func: Run
//--------------------------------------------------------------------------

// Run uses Packet.net API to launch a new server.
func (d *Data) Run() {

	// Set current command:
	d.command = "run"

	// Read udata from stdin:
	udata, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.WithField("cmd", "pkt:"+d.command).Fatal(err)
	}

	// Connect and authenticate to the API endpoint:
	client := packngo.NewClient("", d.APIKey, nil)

	// Forge the request:
	createRequest := &packngo.DeviceCreateRequest{
		Hostname:     d.HostName,
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
		log.WithField("cmd", "pkt:"+d.command).Fatal(err)
	}

	// Pretty-print the response data:
	log.WithField("cmd", "pkt:"+d.command).Info(newDevice)
}
