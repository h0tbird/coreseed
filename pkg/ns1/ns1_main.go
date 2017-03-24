package ns1

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"net/http"
	"strings"
	"time"

	// Community:
	log "github.com/Sirupsen/logrus"
	api "gopkg.in/ns1/ns1-go.v2/rest"
	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
)

//-----------------------------------------------------------------------------
// Typedefs:
//-----------------------------------------------------------------------------

// Data struct for NS1.
type Data struct {
	command string
	Zones   []string
	APIKey  string
	Zone    string
	Records []string
}

//-----------------------------------------------------------------------------
// func: AddZones
//-----------------------------------------------------------------------------

// AddZones adds one or more zones to NS1.
func (d *Data) AddZones() {

	// Set the current command:
	d.command = "zone:add"

	// Create an NS1 API client:
	httpClient := &http.Client{Timeout: time.Second * 10}
	client := api.NewClient(httpClient, api.SetAPIKey(d.APIKey))

	// For each requested zone:
	for _, zone := range d.Zones {

		// New zone handler:
		z := dns.NewZone(zone)

		// Send the new zone request:
		if _, err := client.Zones.Create(z); err != nil {
			if err != api.ErrZoneExists {
				log.WithFields(log.Fields{"cmd": "ns1:" + d.command, "id": zone}).
					Fatal(err)
			}
		}

		// Log zone creation:
		log.WithFields(log.Fields{"cmd": "ns1:" + d.command, "id": zone}).
			Info("New DNS zone created")
	}
}

//-----------------------------------------------------------------------------
// func: DelZones
//-----------------------------------------------------------------------------

// DelZones deletes one or more zones from NS1.
func (d *Data) DelZones() {

	// Set the current command:
	d.command = "zone:del"

	// Create an NS1 API client:
	httpClient := &http.Client{Timeout: time.Second * 10}
	client := api.NewClient(httpClient, api.SetAPIKey(d.APIKey))

	// For each requested zone:
	for _, zone := range d.Zones {

		// Send the delete zone request:
		if _, err := client.Zones.Delete(zone); err != nil {
			log.WithFields(log.Fields{"cmd": "ns1:" + d.command, "id": zone}).
				Fatal(err)
		}

		// Log zone deletion:
		log.WithFields(log.Fields{"cmd": "ns1:" + d.command, "id": zone}).
			Info("DNS zone deleted")
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
	httpClient := &http.Client{Timeout: time.Second * 10}
	client := api.NewClient(httpClient, api.SetAPIKey(d.APIKey))

	// For each requested record:
	for _, record := range d.Records {

		// New record handler:
		s := strings.Split(record, ":")
		r := dns.NewRecord(d.Zone, s[2], s[1])
		a := dns.NewAv4Answer(s[0])
		r.AddAnswer(a)

		// Send the new record request:
		if _, err := client.Records.Create(r); err != nil {
			if err != api.ErrRecordExists {
				log.WithFields(log.Fields{"cmd": "ns1:" + d.command, "id": record}).Fatal(err)
			}
		}

		// Log record creation:
		log.WithFields(log.Fields{"cmd": "ns1:" + d.command, "id": record}).
			Info("New DNS record created")
	}
}
