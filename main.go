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
	Hostname string
}

//-----------------------------------------------------------------------------
// Package variable declarations factored into a block:
//-----------------------------------------------------------------------------

var (

	// coreseed: top level command
	app = kingpin.New("coreseed", "Coreseed defines and deploys CoreOS clusters.")

	flVerbose = app.Flag("verbose", "Enable verbose mode.").
			OverrideDefaultFromEnvar("CORESEED_VERBOSE").
			Short('v').Bool()

	// udata: nested command
	cmdData = app.Command("data", "Generate CoreOS cloud-config user-data.")

	flHostName = cmdData.Flag("hostname", "Short host name as in (hostname -s).").
			Required().PlaceHolder("CORESEED_HOSTNAME").
			OverrideDefaultFromEnvar("CORESEED_HOSTNAME").
			Short('h').String()

	flDomain = cmdData.Flag("domain", "Domain name as in (hostname -d).").
			Required().PlaceHolder("CORESEED_DOMAIN").
			OverrideDefaultFromEnvar("CORESEED_DOMAINE").
			Short('d').String()

	flHostRole = cmdData.Flag("role", "Choose one of [ master | slave | edge].").
			Required().PlaceHolder("CORESEED_ROLE").
			OverrideDefaultFromEnvar("CORESEED_ROLE").
			Short('r').String()

	flNs1Apikey = cmdData.Flag("ns1apikey", "NS1 API key.").
			Required().PlaceHolder("CORESEED_NS1_KEY").
			OverrideDefaultFromEnvar("CORESEED_NS1_KEY").
			Short('k').String()

	flFleetTags = cmdData.Flag("fleettags", "Comma separated list of tags.").
			PlaceHolder("CORESEED_FLEET_FLAGS").
			OverrideDefaultFromEnvar("CORESEED_FLEET_FLAGS").
			Short('f').String()

	// run: nested command
	cmdRun      = app.Command("run", "Starts a CoreOS instance.")
	flPktApiKey = cmdRun.Arg("pktApiKey", "Packet API key.").Required().String()
)

//----------------------------------------------------------------------------
// Entry point:
//----------------------------------------------------------------------------

func main() {

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case cmdData.FullCommand():
		println("CMD: udata")
	case cmdRun.FullCommand():
		println("CMD: start")
	}

	udata := Udata{
		Hostname: *flHostName,
	}

	switch *flHostRole {
	case "master":
		t := template.New("udata")
		t, err := t.Parse(templ_master)
		err = t.Execute(os.Stdout, udata)
		checkError(err)
	case "slave":
		t := template.New("udata")
		t, err := t.Parse(templ_slave)
		err = t.Execute(os.Stdout, udata)
		checkError(err)
	case "edge":
		t := template.New("udata")
		t, err := t.Parse(templ_edge)
		err = t.Execute(os.Stdout, udata)
		checkError(err)
	}
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
