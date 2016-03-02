//-----------------------------------------------------------------------------
// Package membership:
//-----------------------------------------------------------------------------

package main

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Standard library:
	"fmt"
	"os"
	"text/template"

	// Community:
	"gopkg.in/alecthomas/kingpin.v2"
)

//----------------------------------------------------------------------------
// Typedefs:
//----------------------------------------------------------------------------

type Udata struct {
	Hostname  string
	Domain    string
	Role      string
	Ns1apikey string
	Fleettags string
}

//-----------------------------------------------------------------------------
// Package variable declarations factored into a block:
//-----------------------------------------------------------------------------

var (

	//-----------------------------
	// coreseed: top level command
	//-----------------------------

	app = kingpin.New("coreseed", "Coreseed defines and deploys CoreOS clusters.")

	flVerbose = app.Flag("verbose", "Enable verbose mode.").
			OverrideDefaultFromEnvar("CS_VERBOSE").
			Short('v').Bool()

	//-----------------------
	// udata: nested command
	//-----------------------

	cmdData = app.Command("data", "Generate CoreOS cloud-config user-data.")

	flHostName = cmdData.Flag("hostname", "Short host name as in (hostname -s).").
			Required().PlaceHolder("$CS_HOSTNAME").
			OverrideDefaultFromEnvar("CS_HOSTNAME").
			Short('h').String()

	flDomain = cmdData.Flag("domain", "Domain name as in (hostname -d).").
			Required().PlaceHolder("$CS_DOMAIN").
			OverrideDefaultFromEnvar("CS_DOMAIN").
			Short('d').String()

	flHostRole = cmdData.Flag("role", "Choose one of [ master | slave | edge].").
			Required().PlaceHolder("$CS_ROLE").
			OverrideDefaultFromEnvar("CS_ROLE").
			Short('r').String()

	flNs1Apikey = cmdData.Flag("ns1apikey", "NS1 private API key.").
			Required().PlaceHolder("$CS_NS1_KEY").
			OverrideDefaultFromEnvar("CS_NS1_KEY").
			Short('k').String()

	flFleetTags = cmdData.Flag("tags", "Comma separated list of fleet tags.").
			PlaceHolder("$CS_TAGS").
			OverrideDefaultFromEnvar("CS_TAGS").
			Short('t').String()

	//---------------------
	// run: nested command
	//---------------------

	cmdRun = app.Command("run", "Starts a CoreOS instance.")

	flPktApiKey = cmdRun.Arg("pktApiKey", "Packet API key.").
			Required().String()
)

//----------------------------------------------------------------------------
// Entry point:
//----------------------------------------------------------------------------

func main() {

	// Sub-command selector:
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case cmdData.FullCommand():
		cmd_data()
	case cmdRun.FullCommand():
		cmd_run()
	}
}

//--------------------------------------------------------------------------
// func: cmd_data
//--------------------------------------------------------------------------

func cmd_data() {

	// Template data structure:
	udata := Udata{
		Hostname:  *flHostName,
		Domain:    *flDomain,
		Role:      *flHostRole,
		Ns1apikey: *flNs1Apikey,
		Fleettags: *flFleetTags,
	}

	if *flVerbose {
		fmt.Printf("hostname: %s\n", udata.Hostname)
		fmt.Printf("domain: %s\n", udata.Domain)
		fmt.Printf("role: %s\n", udata.Role)
		fmt.Printf("ns1apikey: %s\n", udata.Ns1apikey)
		fmt.Printf("fleettags: %s\n", udata.Fleettags)
	}

	// Render the template for the selected role:
	switch *flHostRole {
	case "master":
		t := template.New("master_udata")
		t, err := t.Parse(templ_master)
		err = t.Execute(os.Stdout, udata)
		checkError(err)
	case "slave":
		t := template.New("slave_udata")
		t, err := t.Parse(templ_slave)
		err = t.Execute(os.Stdout, udata)
		checkError(err)
	case "edge":
		t := template.New("edge_udata")
		t, err := t.Parse(templ_edge)
		err = t.Execute(os.Stdout, udata)
		checkError(err)
	}
}

//--------------------------------------------------------------------------
// func: cmd_run
//--------------------------------------------------------------------------

func cmd_run() {
	println("CMD: run")
}

//---------------------------------------------------------------------------
// func: checkError
//---------------------------------------------------------------------------

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
