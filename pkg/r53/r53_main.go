package r53

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"os"
	"os/exec"
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
	zone := normalizeZoneName(*d.Zone.HostedZone.Name)
	*d.Zone.HostedZone.Name = zone
	if _, err := d.getZone(zone); err != nil {
		log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
			Fatal(err)
	}

	// Return if zone is missing:
	if d.Zone.Id == nil || *d.Zone.Id == "" {
		log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
			Fatal("Ops! This zone does not exist")
	}

	// For each requested record:
	for _, record := range d.Records {
		if err := d.addRecord(record); err != nil {
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

		// Normalize the zone name:
		zone = normalizeZoneName(zone)
		d.Zone.HostedZone.Name = &zone

		// Add the child zone:
		if err := d.addZone(); err != nil {
			log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
				Fatal(err)
		}

		// Get the parent zone:
		pZone, err := d.getParentZone()
		if err != nil {
			log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": zone}).
				Fatal(err)
		}

		// If any:
		if pZone != "" {
			if err := d.delegateZone(pZone); err != nil {
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

		// Normalize the zone name:
		zone = normalizeZoneName(zone)
		d.Zone.HostedZone.Name = &zone

		// Delete the child zone:
		if err := d.delZone(); err != nil {
			log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": d.Zone}).
				Fatal(err)
		}
	}
}

//-----------------------------------------------------------------------------
// func: addRecord
//-----------------------------------------------------------------------------

func (d *Data) addRecord(record string) error {

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
	log.WithFields(log.Fields{"cmd": "r53:" + d.command, "id": resourceName}).
		Info("DNS record created/updated")

	return nil
}

//-----------------------------------------------------------------------------
// func: addZone
//-----------------------------------------------------------------------------

func (d *Data) addZone() error {

	// Get the zone data:
	zone := *d.Zone.HostedZone.Name
	if _, err := d.getZone(zone); err != nil {
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
		if _, err := d.getZone(zone); err != nil {
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
	if _, err := d.getZone(zone); err != nil {
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
// func: getParentZone
//-----------------------------------------------------------------------------

func (d *Data) getParentZone() (string, error) {

	var err error
	var id, parent string

	// Split the given zone:
	zone := *d.Zone.HostedZone.Name
	split := strings.Split(zone, ".")

	// Iterate over parent zones:
	for i := len(split) - 1; i > 1; i-- {

		// Get the parent zone ID (if any):
		parent = strings.Join(split[len(split)-i:], ".")
		id, err = d.getZone(parent)
		if err != nil {
			return "", err
		}

		// Break if found:
		if id != "" {
			break
		}
	}

	// Not found:
	if id == "" {
		return "", nil
	}

	return parent, nil
}

//-----------------------------------------------------------------------------
// func: getZone
//-----------------------------------------------------------------------------

func (d *Data) getZone(zone string) (string, error) {

	// Forge the zone list request:
	pZone := &route53.ListHostedZonesByNameInput{
		DNSName:  aws.String(zone),
		MaxItems: aws.String("1"),
	}

	// Send the zone list request:
	rZone, err := d.r53.ListHostedZonesByName(pZone)
	if err != nil {
		return "", err
	}

	// Zone does not exist:
	if len(rZone.HostedZones) < 1 || *rZone.HostedZones[0].Name != zone {
		return "", nil
	}

	// If adding a zone:
	if zone == *d.Zone.HostedZone.Name {

		// Forge the NS record list request:
		pRsrc := &route53.ListResourceRecordSetsInput{
			HostedZoneId:    aws.String(*rZone.HostedZones[0].Id),
			MaxItems:        aws.String("1"),
			StartRecordName: aws.String(zone),
			StartRecordType: aws.String("NS"),
		}

		// Send the NS record list request:
		rRsrc, err := d.r53.ListResourceRecordSets(pRsrc)
		if err != nil {
			return "", err
		}

		// Save the data:
		d.Zone.HostedZone = *rZone.HostedZones[0]
		d.Zone.ResourceRecordSet = *rRsrc.ResourceRecordSets[0]
	}

	return *rZone.HostedZones[0].Id, nil
}

//-----------------------------------------------------------------------------
// func: delegateZone
//-----------------------------------------------------------------------------

func (d *Data) delegateZone(pZone string) error {

	// Extract name servers:
	var ns []string
	for _, record := range d.Zone.ResourceRecords {
		ns = append(ns, *record.Value)
	}

	// Forge the 'record add' command:
	zone := strings.Replace(*d.Zone.HostedZone.Name, "."+pZone, "", 1)
	cmd := exec.Command("katoctl", "r53", "record", "add",
		"--zone", pZone, zone+":NS:"+strings.Join(ns, ","))

	// Execute the 'record add' command:
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: normalizeZoneName
//-----------------------------------------------------------------------------

func normalizeZoneName(zone string) string {
	if zone[len(zone)-1] != []byte(".")[0] {
		zone = zone + "."
	}
	return zone
}
