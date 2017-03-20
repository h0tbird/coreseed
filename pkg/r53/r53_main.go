package r53

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
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

// Data struct for Route 53 information.
type Data struct {
	command string
	Zones   []string // zone:add |
}

//-----------------------------------------------------------------------------
// func: AddZones
//-----------------------------------------------------------------------------

// AddZones adds one or more zones to Route 53.
func (d *Data) AddZones() {

	// Set the current command:
	d.command = "zone:add"

	// Create the service handler:
	r53 := route53.New(session.Must(session.NewSession()))

	// For each requested zone:
	for _, zone := range d.Zones {

		// Forge the list request:
		pList := &route53.ListHostedZonesByNameInput{
			DNSName:  aws.String(zone),
			MaxItems: aws.String("1"),
		}

		// Send the list request:
		resp, err := r53.ListHostedZonesByName(pList)
		if err != nil {
			log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
				Fatal(err)
		}

		// If zone doesn't exist:
		if len(resp.HostedZones) < 1 || *resp.HostedZones[0].Name != zone+"." {

			// Forge the new zone request:
			pZone := &route53.CreateHostedZoneInput{
				CallerReference: aws.String(time.Now().Format(time.RFC3339Nano)),
				Name:            aws.String(zone),
			}

			// Send the new zone request:
			if _, err := r53.CreateHostedZone(pZone); err != nil {
				log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
					Fatal(err)
			}
		}

		// Log zone creation:
		log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
			Info("New DNS zone created")
	}
}

//-----------------------------------------------------------------------------
// func: AddRecords
//-----------------------------------------------------------------------------

// AddRecords adds one or more records to a Route 53 zone.
func (d *Data) AddRecords() {}
