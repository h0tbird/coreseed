package ns1

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"strings"

	// Community:
	log "github.com/Sirupsen/logrus"
	"github.com/bobtfish/go-nsone-api"
)

//-----------------------------------------------------------------------------
// Typedefs:
//-----------------------------------------------------------------------------

// Data struct for NS1 information.
type Data struct {
	command string
	Link    string   // zone:add |
	Zones   []string // zone:add |
	APIKey  string   // zone:add | record:add
	Zone    string   //          | record:add
	Records []string //          | record:add
}

//-----------------------------------------------------------------------------
// func: AddZones
//-----------------------------------------------------------------------------

// AddZones adds one or more zones to NS1.
func (d *Data) AddZones() {

	// Set the current command:
	d.command = "zone:add"

	// Create an NS1 API client:
	api := nsone.New(d.APIKey)

	// Retrieve the current zone list:
	zones, err := api.GetZones()
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

Zone:
	// For each requested zone:
	for _, e := range d.Zones {

		// New zone handler:
		z := nsone.NewZone(e)

		// Setup link if defined:
		if d.Link != "" {
			z.LinkTo(d.Link)
		}

		// Continue if already exists:
		for _, v := range zones {
			if v.Zone == z.Zone {
				log.WithFields(log.Fields{"cmd": "ns1:" + d.command, "id": e}).
					Info("Using existing DNS zone")
				continue Zone
			}
		}

		// Send the new zone request:
		if err := api.CreateZone(z); err != nil {
			log.WithFields(log.Fields{"cmd": "ns1:" + d.command, "id": e}).Fatal(err)
		}

		// Log zone creation:
		log.WithFields(log.Fields{"cmd": "ns1:" + d.command, "id": e}).
			Info("New DNS zone created")
	}
}

//-----------------------------------------------------------------------------
// func: AddRecords
//-----------------------------------------------------------------------------

// AddRecords adds one or more records to an NS1 zone.
func (d *Data) AddRecords() {

	// Set the current command:
	d.command = "record:add"

	// Create an NS1 API client:
	api := nsone.New(d.APIKey)

Record:
	// For each requested record:
	for _, e := range d.Records {

		// New record handler:
		s := strings.Split(e, ":")
		r1 := nsone.NewRecord(d.Zone, s[2]+"."+d.Zone, s[1])
		r1.Answers = make([]nsone.Answer, 1)
		r1.Answers[0] = nsone.NewAnswer()
		r1.Answers[0].Answer = []string{s[0]}

		// Attempt to retrieve an existing record:
		r2, err := api.GetRecord(d.Zone, s[2]+"."+d.Zone, s[1])
		if err != nil {
			log.WithFields(log.Fields{"cmd": "ns1:" + d.command, "id": e}).Fatal(err)
		}

		// Compare and continue:
		if r1.Domain == r2.Domain {
			log.WithFields(log.Fields{"cmd": "ns1:" + d.command, "id": e}).
				Info("Using existing DNS record")
			continue Record
		}

		// Send the new record request:
		err = api.CreateRecord(r1)
		if err != nil {
			log.WithFields(log.Fields{"cmd": "ns1:" + d.command, "id": e}).Fatal(err)
		}

		// Log record creation:
		log.WithFields(log.Fields{"cmd": "ns1:" + d.command, "id": e}).
			Info("New DNS record created")
	}
}
