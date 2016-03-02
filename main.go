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

	// coreseed: top level command
	app   = kingpin.New("coreseed", "Coreseed defines and deploys CoreOS clusters.")
	debug = app.Flag("debug", "Enable debug mode.").Bool()

	// udata: nested command
	cmdData     = app.Command("data", "Generate CoreOS cloud-config user-data.")
	flHostName  = cmdData.Flag("hostname", "Short host name as in (hostname -s).").Default("core-1").String()
	flDomain    = cmdData.Flag("domain", "Domain name as in (hostname -d).").Default("demo.lan").String()
	flRole      = cmdData.Flag("role", "Choose one of [ master | slave | edge]").Required().String()
	flNs1Apikey = cmdData.Flag("ns1apikey", "NS1 API key.").Required().String()
	flFleetTags = cmdData.Flag("fleettags", "Comma separated list of tags.").String()

	// run: nested command
	cmdRun      = app.Command("run", "Starts a CoreOS instance.")
	flPktApiKey = cmdRun.Arg("pktApiKey", "Packet API key.").Required().String()
)

//----------------------------------------------------------------------------
//
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
