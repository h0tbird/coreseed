package ns1

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (
	"github.com/katosys/kato/pkg/cli"
)

//-----------------------------------------------------------------------------
// 'katoctl ns1' command flags definitions:
//-----------------------------------------------------------------------------

var (

	// ns1:
	cmdNs1      = cli.App.Command("ns1", "Manages NS1 zones and records.")
	flNs1APIKey = cmdNs1.Flag("api-key",
		"NS1 private API key").Required().String()

	// ns1 zone add:
	cmdNs1Zone       = cmdNs1.Command("zone", "Manage NS1 zones.")
	cmdNs1ZoneAdd    = cmdNs1Zone.Command("add", "Adds NS1 zones.")
	flNs1ZoneAddLink = cmdNs1ZoneAdd.Flag("link",
		"Links the added zone to the link zone.").String()
	arNs1ZoneAddName = cmdNs1ZoneAdd.Arg("fqdn",
		"List of zones to publish").Required().Strings()

	// ns1 record add:
	cmdNs1Record       = cmdNs1.Command("record", "Manage NS1 records.")
	cmdNs1RecordAdd    = cmdNs1Record.Command("add", "Adds records to NS1 zones.")
	flNs1RecordAddZone = cmdNs1RecordAdd.Flag("zone",
		"DNS zone where records are added.").Required().String()
	arNs1RecordAddName = cmdNs1RecordAdd.Arg("record",
		"List of ip:type:dns records.").Required().Strings()
)

//-----------------------------------------------------------------------------
// RunCmd:
//-----------------------------------------------------------------------------

// RunCmd runs the cmd if owned by this package.
func RunCmd(cmd string) bool {

	switch cmd {

	// katoctl ns1 zone add:
	case cmdNs1ZoneAdd.FullCommand():
		d := Data{
			APIKey: *flNs1APIKey,
			Link:   *flNs1ZoneAddLink,
			Zones:  *arNs1ZoneAddName,
		}
		d.AddZones()

	// katoctl ns1 record add:
	case cmdNs1RecordAdd.FullCommand():
		d := Data{
			APIKey:  *flNs1APIKey,
			Zone:    *flNs1RecordAddZone,
			Records: *arNs1RecordAddName,
		}
		d.AddRecords()

	// Nothing to do:
	default:
		return false
	}

	return true
}
