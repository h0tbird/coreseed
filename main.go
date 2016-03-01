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

type Udata struct {
	Hostname string
}

//-----------------------------------------------------------------------------
// Package variable declarations factored into a block:
//-----------------------------------------------------------------------------

var (

	// Top level:
	app   = kingpin.New("coreseed", "CoreSeed is used to deploy CoreOS clusters.")
	debug = app.Flag("debug", "Enable debug mode.").Bool()

	// udata: nested command
	cmdUdata    = app.Command("udata", "Generate CoreOS cloud-config user-data.")
	flHostName  = cmdUdata.Flag("hostname", "Short host name as in (hostname -s).").Required().String()
	flDomain    = cmdUdata.Flag("domain", "Domain name as in (hostname -d).").Required().String()
	flRole      = cmdUdata.Flag("role", "Choose one of [ master | slave | edge]").Required().String()
	flNs1Apikey = cmdUdata.Flag("ns1apikey", "NS1 API key.").Required().String()
	flFleetTags = cmdUdata.Flag("fleettags", "Comma separated list of tags.").String()

	// run: nested command
	cmdRun      = app.Command("run", "Starts a CoreOS instance.")
	flPktApiKey = cmdRun.Arg("pktApiKey", "Packet API key.").Required().String()
)

//----------------------------------------------------------------------------
//
//----------------------------------------------------------------------------

func main() {

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case cmdUdata.FullCommand():
		println("CMD: udata")
	case cmdRun.FullCommand():
		println("CMD: start")
	}

	udata := Udata{
		Hostname: *flHostName,
	}

	switch *flRole {
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
//
//---------------------------------------------------------------------------

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
