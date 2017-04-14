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

		// Create the record:
		if err := d.addRecord(record, zoneID); err != nil {
			log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": record}).
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

		// Add the zone:
		if err := d.addZone(zone); err != nil {
			log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
				Fatal(err)
		}

		// Setup delegation:
		if _, err := d.delegateZone(zone); err != nil {
			log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
				Fatal(err)
		}
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

		// Delete the zone:
		if err := d.delZone(zone); err != nil {
			log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
				Fatal(err)
		}
	}
}

//-----------------------------------------------------------------------------
// func: addRecord
//-----------------------------------------------------------------------------

func (d *Data) addRecord(record, zoneID string) error {

	// Split into data:type:name
	s := strings.Split(record, ":")
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
	changes := []*route53.Change{{
		Action: aws.String("UPSERT"),
		ResourceRecordSet: &route53.ResourceRecordSet{
			Name:            aws.String(resourceName + "." + d.Zone),
			Type:            aws.String(resourceType),
			TTL:             aws.Int64(300),
			ResourceRecords: resourceRecords,
		},
	}}

	// Forge the change request (outermost matryoshka):
	params := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
		ChangeBatch: &route53.ChangeBatch{
			Changes: changes,
		},
	}

	// Send the change request:
	if _, err := d.r53.ChangeResourceRecordSets(params); err != nil {
		return err
	}

	// Log record creation:
	log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": record}).
		Info("DNS record created/updated")

	return nil
}

//-----------------------------------------------------------------------------
// func: addZone
//-----------------------------------------------------------------------------

func (d *Data) addZone(zone string) error {

	// Get the zone ID:
	id, err := d.getZoneID(zone)
	if err != nil {
		return err
	}

	// If the zone is missing:
	if id == "" {

		// Forge the new zone request:
		params := &route53.CreateHostedZoneInput{
			CallerReference: aws.String(time.Now().Format(time.RFC3339Nano)),
			Name:            aws.String(zone),
		}

		// Send the new zone request:
		if _, err := d.r53.CreateHostedZone(params); err != nil {
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

func (d *Data) delZone(zone string) error {

	// Get the zone ID:
	id, err := d.getZoneID(zone)
	if err != nil {
		return err
	}

	// If the zone exists:
	if id != "" {

		// Forge the delete zone request:
		params := &route53.DeleteHostedZoneInput{
			Id: aws.String(id),
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
// func: delegateZone
//-----------------------------------------------------------------------------

func (d *Data) delegateZone(zone string) (bool, error) {

	var err error
	var zoneID, parent string

	// Split the given zone:
	split := strings.Split(zone, ".")

	// Iterate over parent zones:
	for i := len(split) - 1; i > 1; i-- {

		// Get the parent zone ID (if any):
		parent = strings.Join(split[len(split)-i:], ".")
		zoneID, err = d.getZoneID(parent)
		if err != nil {
			return false, err
		}

		// Break if found:
		if zoneID != "" {
			break
		}
	}

	// Not found:
	if zoneID == "" {
		return false, nil
	}

	// Parent found:
	log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zoneID}).
		Info("Parent zone is " + parent)

	// Add/update the NS records:
	// TODO

	return true, nil
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
