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
	ns1     *api.Client
	command string
	Zones   []string
	APIKey  string
	Zone    string
	Records []string
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
	d.ns1 = api.NewClient(httpClient, api.SetAPIKey(d.APIKey))

	// For each requested record:
	for _, record := range d.Records {
		if err := d.addRecord(record); err != nil {
			log.WithFields(log.Fields{"cmd": "ns1:" + d.command, "id": record}).
				Fatal(err)
		}
	}
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
	d.ns1 = api.NewClient(httpClient, api.SetAPIKey(d.APIKey))

	// For each requested zone:
	for _, zone := range d.Zones {
		if err := d.addZone(zone); err != nil {
			log.WithFields(log.Fields{"cmd": "ns1:" + d.command, "id": zone}).
				Fatal(err)
		}
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
	d.ns1 = api.NewClient(httpClient, api.SetAPIKey(d.APIKey))

	// For each requested zone:
	for _, zone := range d.Zones {
		if err := d.delZone(zone); err != nil {
			log.WithFields(log.Fields{"cmd": "ns1:" + d.command, "id": zone}).
				Fatal(err)
		}
	}
}

//-----------------------------------------------------------------------------
// func: addRecord
//-----------------------------------------------------------------------------

func (d *Data) addRecord(record string) error {

	// Split into name:type:data
	s := strings.Split(record, ":")
	resourceName := s[0]
	resourceType := s[1]
	resourceData := s[2]

	// Forge the record request:
	rec := dns.NewRecord(d.Zone, resourceName, resourceType)
	for _, data := range strings.Split(resourceData, ",") {
		rec.AddAnswer(dns.NewAnswer([]string{data}))
	}

	// Send the record request:
	if _, err := d.ns1.Records.Create(rec); err != nil {
		if err != api.ErrRecordExists {
			return err
		}
	}

	// Log record creation:
	log.WithFields(log.Fields{"cmd": "ns1:" + d.command,
		"id": resourceName + "." + d.Zone}).Info("New DNS record created/updated")

	return nil
}

//-----------------------------------------------------------------------------
// func: addZone
//-----------------------------------------------------------------------------

func (d *Data) addZone(zone string) error {

	// Forge the zone request:
	z := dns.NewZone(zone)

	// Send the zone request:
	if _, err := d.ns1.Zones.Create(z); err != nil {
		if err != api.ErrZoneExists {
			return err
		}
	}

	// Log zone creation:
	log.WithFields(log.Fields{"cmd": "ns1:" + d.command, "id": zone}).
		Info("New DNS zone created")

	return nil
}

//-----------------------------------------------------------------------------
// func: delZone
//-----------------------------------------------------------------------------

func (d *Data) delZone(zone string) error {

	// Send the delete zone request:
	if _, err := d.ns1.Zones.Delete(zone); err != nil {
		return err
	}

	// Log zone deletion:
	log.WithFields(log.Fields{"cmd": "ns1:" + d.command, "id": zone}).
		Info("DNS zone deleted")

	return nil
}
