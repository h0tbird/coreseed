//-----------------------------------------------------------------------------
// Package membership:
//-----------------------------------------------------------------------------

package main

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Standard library:
	"flag"
	"fmt"
	"os"
	"text/template"
)

type Udata struct {
	Hostname string
}

//-----------------------------------------------------------------------------
// Package variable declarations factored into a block:
//-----------------------------------------------------------------------------

var (

	// udata
	fleettags string
	ns1apikey string
	role      string
	domain    string
	hostname  string

	// start
	pktApiKey       string
	pktProjectName  string
	pktDevicePrefix string
	pktPlan         string
	pktFacility     string
	pktOS           string
)

//-----------------------------------------------------------------------------
// func init() is called after all the variable declarations in the Package
// have evaluated their initializers, and those are evaluated only after all
// the imported packages have been initialized:
//-----------------------------------------------------------------------------

func init() {

	switch os.Args[1] {
	case "udata":
		flag.StringVar(&fleettags, "fleet-tags", "", "Comma separated list of tags.")
		flag.StringVar(&ns1apikey, "ns1-api-key", "", "NS1. API key.")
		flag.StringVar(&role, "role", "slave", "Choose one of [ master | slave | edge]")
		flag.StringVar(&domain, "domain", "cell-1.dc-1.demo.lan", "Domain name as in 'hostname -d'")
		flag.StringVar(&hostname, "hostname", "core-1", "Short host name as in 'hostname -s'")
		flag.Usage = usage
		flag.Parse()
	case "start":
		flag.StringVar(&pktApiKey, "packet-api-key", "", "Packet private API key.")
		flag.StringVar(&pktProjectName, "packet-project-name", "", "Packet project name.")
		flag.StringVar(&pktDevicePrefix, "packet-device-prefix", "", "Packet device prefix.")
		flag.StringVar(&pktPlan, "packet-plan", "", "Packet plan.")
		flag.StringVar(&pktFacility, "packet-facility", "", "Packet facility.")
		flag.StringVar(&pktOS, "packet-os", "", "Packet OS.")
		flag.Usage = usage
		flag.Parse()
	default:
		usage()
	}
	usage()
}

//----------------------------------------------------------------------------
// func usage() reports the correct commandline usage:
//----------------------------------------------------------------------------

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [ udata | start ] [arg...]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

//----------------------------------------------------------------------------
//
//----------------------------------------------------------------------------

func main() {

	udata := Udata{
		Hostname: hostname,
	}

	switch role {
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
	default:
		usage()
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
