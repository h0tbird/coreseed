package r53

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/katosys/kato/pkg/cli"
)

//-----------------------------------------------------------------------------
// 'katoctl r53' command flags definitions:
//-----------------------------------------------------------------------------

var (

	// r53 zone/record:
	cmdR53       = cli.App.Command("r53", "Manages Route 53 zones and records.")
	flR53APIKey  = cmdR53.Flag("api-key", "R53 private API key.").String()
	cmdR53Zone   = cmdR53.Command("zone", "Manage Route 53 zones.")
	cmdR53Record = cmdR53.Command("record", "Manage Route 53 records.")

	// r53 zone add:
	cmdR53ZoneAdd    = cmdR53Zone.Command("add", "Adds Route 53 zones.")
	arR53ZoneAddName = cmdR53ZoneAdd.Arg("fqdn",
		"List of zones to publish.").Required().Strings()

	// r53 zone del:
	cmdR53ZoneDel    = cmdR53Zone.Command("del", "Deletes Route 53 zones.")
	arR53ZoneDelName = cmdR53ZoneDel.Arg("fqdn",
		"List of zones to delete.").Required().Strings()

	// r53 record add:
	cmdR53RecordAdd    = cmdR53Record.Command("add", "Adds records to Route 53 zones.")
	flR53RecordAddZone = cmdR53RecordAdd.Flag("zone",
		"DNS zone where records are added.").Required().String()
	arR53RecordAddName = cmdR53RecordAdd.Arg("record",
		"List of name:type:data records.").Required().Strings()
)

//-----------------------------------------------------------------------------
// RunCmd:
//-----------------------------------------------------------------------------

// RunCmd runs the cmd if owned by this package.
func RunCmd(cmd string) bool {

	switch cmd {

	// katoctl r53 zone add:
	case cmdR53ZoneAdd.FullCommand():
		d := Data{
			APIKey: *flR53APIKey,
			Zones:  *arR53ZoneAddName,
		}
		d.AddZones()

	// katoctl r53 zone del:
	case cmdR53ZoneDel.FullCommand():
		d := Data{
			APIKey: *flR53APIKey,
			Zones:  *arR53ZoneDelName,
		}
		d.DelZones()

	// katoctl r53 record add:
	case cmdR53RecordAdd.FullCommand():
		d := Data{
			APIKey: *flR53APIKey,
			Zone: zoneData{
				HostedZone: route53.HostedZone{
					Name: flR53RecordAddZone,
				},
			},
			Records: *arR53RecordAddName,
		}
		d.AddRecords()

	// Nothing to do:
	default:
		return false
	}

	return true
}
