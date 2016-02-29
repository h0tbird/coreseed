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
	"log"
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
	role     = flag.String("role", "slave", "Choose one of [master|slave|edge]")
	domain   = flag.String("domain", "cell-1.dc-1.demo.lan", "Domain name as in `hostname -d`")
	hostname = flag.String("hostname", "core-1", "Short host name as in `hostname -s`")
)

//-----------------------------------------------------------------------------
// func init() is called after all the variable declarations in the Package
// have evaluated their initializers, and those are evaluated only after all
// the imported packages have been initialized:
//-----------------------------------------------------------------------------

func init() {

	// Check for mandatory argc:
	if len(os.Args) < 2 {
		usage()
	}

	// Change the flags on the default logger:
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Parse commandline flags:
	flag.Usage = usage
	flag.Parse()
}

//----------------------------------------------------------------------------
// func usage() reports the correct commandline usage:
//----------------------------------------------------------------------------

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

//----------------------------------------------------------------------------
//
//----------------------------------------------------------------------------

func main() {

	udata := Udata{
		Hostname: *hostname,
	}

	switch *role {
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
