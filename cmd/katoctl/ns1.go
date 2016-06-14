package main

//-----------------------------------------------------------------------------
// 'katoctl ns1' command flags definitions:
//-----------------------------------------------------------------------------

var (

	//------------------------
	// ns1: top level command
	//------------------------

	cmdNs1 = app.Command("ns1", "Manages NS1 zones and records.")

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
		"List of zones to publish").Strings()
)
