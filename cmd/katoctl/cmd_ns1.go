package katoctl

//-----------------------------------------------------------------------------
// Import:
//-----------------------------------------------------------------------------

import "github.com/katosys/kato/pkg/cli"

//-----------------------------------------------------------------------------
// 'katoctl ns1' command flags definitions:
//-----------------------------------------------------------------------------

var (

	//------------------------
	// ns1: top level command
	//------------------------

	cmdNs1 = cli.App.Command("ns1", "Manages NS1 zones and records.")

	flNs1APIKey = cmdNs1.Flag("api-key",
		"NS1 private API key").Required().String()

	//--------------------------
	// ns1 zone: nested command
	//--------------------------

	cmdNs1Zone = cmdNs1.Command("zone", "Manage NS1 zones.")

	//------------------------------
	// ns1 zone add: nested command
	//------------------------------

	cmdNs1ZoneAdd = cmdNs1Zone.Command("add", "Adds NS1 zones.")

	flNs1ZoneAddLink = cmdNs1ZoneAdd.Flag("link",
		"Links the added zone to the link zone.").String()

	arNs1ZoneAddName = cmdNs1ZoneAdd.Arg("fqdn",
		"List of zones to publish").Required().Strings()

	//----------------------------
	// ns1 record: nested command
	//----------------------------

	cmdNs1Record = cmdNs1.Command("record", "Manage NS1 records.")

	//--------------------------------
	// ns1 record add: nested command
	//--------------------------------

	cmdNs1RecordAdd = cmdNs1Record.Command("add", "Adds records to NS1 zones.")

	flNs1RecordAddZone = cmdNs1RecordAdd.Flag("zone",
		"DNS zone where records are added.").Required().String()

	arNs1RecordAddName = cmdNs1RecordAdd.Arg("record",
		"List of ip:type:dns records.").
		Required().Strings()
)
