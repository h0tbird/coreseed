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

type zoneData struct {
	route53.HostedZone
	route53.ResourceRecordSet
}

// Data struct for Route53.
type Data struct {
	r53     *route53.Route53
	command string
	APIKey  string
	Record  string
	Zone    zoneData
	Records []string
	Zones   []string
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

	// Get the zone data:
	zone := *d.Zone.HostedZone.Name
	if err := d.getZone(); err != nil {
		log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
			Fatal(err)
	}

	// Return if zone is missing:
	if d.Zone.Id == nil || *d.Zone.Id == "" {
		log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
			Fatal("Ops! This zone does not exist")
	}

	// For each requested record:
	for _, d.Record = range d.Records {
		if err := d.addRecord(); err != nil {
			log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": d.Record}).
				Fatal(err)
		}
	}
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

		// Add the child zone:
		d.Zone.HostedZone.Name = &zone
		if err := d.addZone(); err != nil {
			log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
				Fatal(err)
		}

		// Get the parent zone ID...
		parentZoneID, err := d.getParentZoneID()
		if err != nil {
			log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
				Fatal(err)
		}

		// ...if any:
		if parentZoneID != "" {
			if err := d.delegateZone(parentZoneID); err != nil {
				log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
					Fatal(err)
			}
		}

		// Clean zone data:
		d.Zone.Id = nil
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
		d.Zone.HostedZone.Name = &zone
		if err := d.delZone(); err != nil {
			log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": d.Zone}).
				Fatal(err)
		}
	}
}

//-----------------------------------------------------------------------------
// func: addRecord
//-----------------------------------------------------------------------------

func (d *Data) addRecord() error {

	// Split into data:type:name
	s := strings.Split(d.Record, ":")
	resourceName := s[0]
	resourceType := s[1]
	resourceData := s[2]

	// Resource records (innermost matryoshka):
	resourceRecords := []*route53.ResourceRecord{}
	for _, resource := range strings.Split(resourceData, ",") {
		resourceRecords = append(resourceRecords, &route53.ResourceRecord{
			Value: aws.String(resource),
		})
	}

	// Changes (middle matryoshka):
	zone := *d.Zone.HostedZone.Name
	changes := []*route53.Change{{
		Action: aws.String("UPSERT"),
		ResourceRecordSet: &route53.ResourceRecordSet{
			Name:            aws.String(resourceName + "." + zone),
			Type:            aws.String(resourceType),
			TTL:             aws.Int64(300),
			ResourceRecords: resourceRecords,
		},
	}}

	// Forge the change request (outermost matryoshka):
	params := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(*d.Zone.Id),
		ChangeBatch: &route53.ChangeBatch{
			Changes: changes,
		},
	}

	// Send the change request:
	if _, err := d.r53.ChangeResourceRecordSets(params); err != nil {
		return err
	}

	// Log record creation:
	log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": d.Record}).
		Info("DNS record created/updated")

	return nil
}

//-----------------------------------------------------------------------------
// func: addZone
//-----------------------------------------------------------------------------

func (d *Data) addZone() error {

	// Get the zone data:
	zone := *d.Zone.HostedZone.Name
	if err := d.getZone(); err != nil {
		return err
	}

	// If the zone is missing:
	if d.Zone.Id == nil || *d.Zone.Id == "" {

		// Forge the new zone request:
		params := &route53.CreateHostedZoneInput{
			CallerReference: aws.String(time.Now().Format(time.RFC3339Nano)),
			Name:            aws.String(zone),
		}

		// Send the new zone request:
		if _, err := d.r53.CreateHostedZone(params); err != nil {
			return err
		}

		// Get the zone data:
		if err := d.getZone(); err != nil {
			return err
		}

		// Log the new zone creation:
		log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
			Info("New DNS zone created")

	} else {

		// Log zone already exists:
		log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
			Info("DNS zone already exists")
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: delZone
//-----------------------------------------------------------------------------

func (d *Data) delZone() error {

	// Get the zone data:
	zone := *d.Zone.HostedZone.Name
	if err := d.getZone(); err != nil {
		return err
	}

	// If the zone exists:
	if d.Zone.Id != nil && *d.Zone.Id != "" {

		// Forge the delete zone request:
		params := &route53.DeleteHostedZoneInput{
			Id: aws.String(*d.Zone.Id),
		}

		// Send the delete zone request:
		if _, err := d.r53.DeleteHostedZone(params); err != nil {
			return err
		}

		// Log zone deletion:
		log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
			Info("DNS zone deleted")

		return nil
	}

	// Log zone already gone:
	log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
		Info("Ops! this zone does not exist")

	return nil
}

//-----------------------------------------------------------------------------
// func: getParentZoneID
//-----------------------------------------------------------------------------

func (d *Data) getParentZoneID() (string, error) {

	var err error
	var zoneID, parent string

	// Split the given zone:
	zone := *d.Zone.HostedZone.Name
	split := strings.Split(zone, ".")

	// Iterate over parent zones:
	for i := len(split) - 1; i > 1; i-- {

		// Get the parent zone ID (if any):
		parent = strings.Join(split[len(split)-i:], ".")
		zoneID, err = d.getZoneID(parent)
		if err != nil {
			return "", err
		}

		// Break if found:
		if zoneID != "" {
			break
		}
	}

	// Not found:
	if zoneID == "" {
		return "", nil
	}

	return zoneID, nil
}

//-----------------------------------------------------------------------------
// func: getZone
//-----------------------------------------------------------------------------

func (d *Data) getZone() error {

	// Forge the list request:
	zone := *d.Zone.HostedZone.Name
	params := &route53.ListHostedZonesByNameInput{
		DNSName:  aws.String(zone),
		MaxItems: aws.String("1"),
	}

	// Send the list request:
	resp, err := d.r53.ListHostedZonesByName(params)
	if err != nil {
		return err
	}

	// Zone does not exist:
	if len(resp.HostedZones) < 1 || *resp.HostedZones[0].Name != zone+"." {
		return nil
	}

	// Save the zone data:
	d.Zone.HostedZone = *resp.HostedZones[0]
	return nil
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
// func: delegateZone
//-----------------------------------------------------------------------------

func (d *Data) delegateZone(zoneID string) error {
	return nil
}
