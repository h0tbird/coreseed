package r53

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"strings"
	"time"

	// AWS SDK:
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"

	// Community:
	log "github.com/Sirupsen/logrus"
)

//-----------------------------------------------------------------------------
// Typedefs:
//-----------------------------------------------------------------------------

// Data struct for Route53.
type Data struct {
	r53     *route53.Route53
	command string
	APIKey  string
	Zone    string
	Zones   []string
	Records []string
}

//-----------------------------------------------------------------------------
// func: getZoneID
//-----------------------------------------------------------------------------

func (d *Data) getZoneID(zone string) (string, error) {

	// Forge the list request:
	params := &route53.ListHostedZonesByNameInput{
		DNSName:  aws.String(zone),
		MaxItems: aws.String("1"),
	}

	// Send the list request:
	resp, err := d.r53.ListHostedZonesByName(params)
	if err != nil {
		return "", err
	}

	// Zone does not exist:
	if len(resp.HostedZones) < 1 || *resp.HostedZones[0].Name != zone+"." {
		return "", nil
	}

	// Return the zone ID:
	return *resp.HostedZones[0].Id, nil
}

//-----------------------------------------------------------------------------
// func: AddZones
//-----------------------------------------------------------------------------

// AddZones adds one or more zones to Route 53.
func (d *Data) AddZones() {

	// Set the current command:
	d.command = "zone:add"

	// Create the service handler:
	d.r53 = route53.New(session.Must(session.NewSession()))

	// For each requested zone:
	for _, zone := range d.Zones {

		// Get the zone ID:
		zoneID, err := d.getZoneID(zone)
		if err != nil {
			log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
				Fatal(err)
		}

		// If the zone is missing:
		if zoneID == "" {

			// Forge the new zone request:
			params := &route53.CreateHostedZoneInput{
				CallerReference: aws.String(time.Now().Format(time.RFC3339Nano)),
				Name:            aws.String(zone),
			}

			// Send the new zone request:
			if _, err := d.r53.CreateHostedZone(params); err != nil {
				log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
					Fatal(err)
			}

			// Log the new zone creation:
			log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
				Info("New DNS zone created")

			return
		}

		// Log zone already exists:
		log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
			Info("DNS zone already exists")
	}
}

//-----------------------------------------------------------------------------
// func: DelZones
//-----------------------------------------------------------------------------

// DelZones deletes one or more zones from Route 53.
func (d *Data) DelZones() {

	// Set the current command:
	d.command = "zone:del"

	// Create the service handler:
	d.r53 = route53.New(session.Must(session.NewSession()))

	// For each requested zone:
	for _, zone := range d.Zones {

		// Get the zone ID:
		zoneID, err := d.getZoneID(zone)
		if err != nil {
			log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
				Fatal(err)
		}

		// If the zone exists:
		if zoneID != "" {

			// Forge the delete zone request:
			params := &route53.DeleteHostedZoneInput{
				Id: aws.String(zoneID),
			}

			// Send the delete zone request:
			if _, err := d.r53.DeleteHostedZone(params); err != nil {
				log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
					Fatal(err)
			}

			// Log zone deletion:
			log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
				Info("DNS zone deleted")

			return
		}

		// Log zone already gone:
		log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
			Info("Ops! this zone does not exist")
	}
}

//-----------------------------------------------------------------------------
// func: AddRecords
//-----------------------------------------------------------------------------

// AddRecords adds one or more records to a Route 53 zone.
func (d *Data) AddRecords() {

	// Set the current command:
	d.command = "record:add"

	// Create the service handler:
	d.r53 = route53.New(session.Must(session.NewSession()))

	// Get the zone ID:
	zoneID, err := d.getZoneID(d.Zone)
	if err != nil {
		log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": d.Zone}).
			Fatal(err)
	}

	// Return if zone is missing:
	if zoneID == "" {
		log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": d.Zone}).
			Fatal("Ops! This zone does not exist")
	}

	// For each requested record:
	for _, record := range d.Records {

		// New record handler:
		s := strings.Split(record, ":")

		// Forge the change request:
		params := &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(zoneID),
			ChangeBatch: &route53.ChangeBatch{
				Changes: []*route53.Change{
					{
						Action: aws.String("UPSERT"),
						ResourceRecordSet: &route53.ResourceRecordSet{
							Name: aws.String(s[2] + "." + d.Zone),
							Type: aws.String(s[1]),
							TTL:  aws.Int64(300),
							ResourceRecords: []*route53.ResourceRecord{
								{
									Value: aws.String(s[0]),
								},
							},
						},
					},
				},
			},
		}

		// Send the change request:
		if _, err := d.r53.ChangeResourceRecordSets(params); err != nil {
			log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": record}).
				Fatal(err)
		}

		// Log record creation:
		log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": record}).
			Info("DNS record created/updated")
	}
}
