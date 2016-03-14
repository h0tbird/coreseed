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
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	// Community:
	"github.com/packethost/packngo"
	"gopkg.in/alecthomas/kingpin.v2"
)

//----------------------------------------------------------------------------
// Typedefs:
//----------------------------------------------------------------------------

type Udata struct {
	Hostname  string
	Hostid    string
	Domain    string
	Role      string
	Ns1apikey string
	Fleettags string
	CAcert    string
	EtcdTkn   string
}

//-----------------------------------------------------------------------------
// Package variable declarations factored into a block:
//-----------------------------------------------------------------------------

var (

	//-----------------------------
	// coreseed: top level command
	//-----------------------------

	app = kingpin.New("coreseed", "Coreseed defines and deploys CoreOS clusters.")

	//----------------------
	// data: nested command
	//----------------------

	cmdData = app.Command("data", "Generate CoreOS cloud-config user-data.")

	flHostName = cmdData.Flag("hostname", "Short host name as in (hostname -s).").
			Required().PlaceHolder("CS_HOSTNAME").
			OverrideDefaultFromEnvar("CS_HOSTNAME").
			Short('h').String()

	flDomain = cmdData.Flag("domain", "Domain name as in (hostname -d).").
			Required().PlaceHolder("CS_DOMAIN").
			OverrideDefaultFromEnvar("CS_DOMAIN").
			Short('d').String()

	flHostRole = cmdData.Flag("role", "Choose one of [ master | slave | edge ].").
			Required().PlaceHolder("CS_ROLE").
			OverrideDefaultFromEnvar("CS_ROLE").
			Short('r').String()

	flNs1Apikey = cmdData.Flag("ns1-api-key", "NS1 private API key.").
			Required().PlaceHolder("CS_NS1_KEY").
			OverrideDefaultFromEnvar("CS_NS1_KEY").
			Short('k').String()

	flFleetTags = cmdData.Flag("tags", "Comma separated list of fleet tags.").
			PlaceHolder("CS_TAGS").
			OverrideDefaultFromEnvar("CS_TAGS").
			Short('t').String()

	flCAcert = cmdData.Flag("ca-cert", "Path to CA certificate.").
			PlaceHolder("CS_CA_CERT").
			OverrideDefaultFromEnvar("CS_CA_CERT").
			Short('c').String()

	flEtcdTkn = cmdData.Flag("etcd-token", "Provide an etcd discovery token.").
			PlaceHolder("CS_ETCD_TKN").
			OverrideDefaultFromEnvar("CS_ETCD_TKN").
			Short('e').String()

	//----------------------------
	// run-packet: nested command
	//----------------------------

	cmdRunPacket = app.Command("run-packet", "Starts a CoreOS instance on Packet.net")

	flPktApiKey = cmdRunPacket.Flag("api-key", "Packet API key.").
			Required().PlaceHolder("PKT_APIKEY").
			OverrideDefaultFromEnvar("PKT_APIKEY").
			Short('k').String()

	flPktHostName = cmdRunPacket.Flag("hostname", "For the Packet.net dashboard.").
			Required().PlaceHolder("PKT_HOSTNAME").
			OverrideDefaultFromEnvar("PKT_HOSTNAME").
			Short('h').String()

	flPktProjId = cmdRunPacket.Flag("project-id", "Format: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee").
			Required().PlaceHolder("PKT_PROJID").
			OverrideDefaultFromEnvar("PKT_PROJID").
			Short('i').String()

	flPktPlan = cmdRunPacket.Flag("plan", "One of [ baremetal_0 | baremetal_1 | baremetal_2 | baremetal_3 ]").
			Required().PlaceHolder("PKT_PLAN").
			OverrideDefaultFromEnvar("PKT_PLAN").
			Short('p').String()

	flPktOsys = cmdRunPacket.Flag("os", "One of [ coreos_stable | coreos_beta | coreos_alpha ]").
			Required().PlaceHolder("PKT_OS").
			OverrideDefaultFromEnvar("PKT_OS").
			Short('o').String()

	flPktFacility = cmdRunPacket.Flag("facility", "One of [ ewr1 | ams1 ]").
			Required().PlaceHolder("PKT_FACILITY").
			OverrideDefaultFromEnvar("PKT_FACILITY").
			Short('f').String()

	flPktBilling = cmdRunPacket.Flag("billing", "One of [ hourly | monthly ]").
			Required().PlaceHolder("PKT_BILLING").
			OverrideDefaultFromEnvar("PKT_BILLING").
			Short('b').String()
)

//----------------------------------------------------------------------------
// Entry point:
//----------------------------------------------------------------------------

func main() {

	// Sub-command selector:
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	// coreseed data ...
	case cmdData.FullCommand():
		cmd_data()

	// coreseed run-packet ...
	case cmdRunPacket.FullCommand():
		cmd_run_packet()
	}
}

//--------------------------------------------------------------------------
// func: cmd_data
//--------------------------------------------------------------------------

func cmd_data() {

	// Template data structure:
	udata := Udata{
		Hostname:  *flHostName,
		Hostid:    string((*flHostName)[strings.LastIndex(*flHostName, "-")+1:]),
		Domain:    *flDomain,
		Role:      *flHostRole,
		Ns1apikey: *flNs1Apikey,
		Fleettags: *flFleetTags,
		EtcdTkn:   *flEtcdTkn,
	}

	// Read the CA certificate:
	if *flCAcert != "" {
		dat, err := ioutil.ReadFile(*flCAcert)
		checkError(err)
		udata.CAcert = strings.TrimSpace(strings.Replace(string(dat), "\n", "\n    ", -1))
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
// func: cmd_run_packet
//--------------------------------------------------------------------------

func cmd_run_packet() {

	// Read user-data from stdin:
	udata, err := ioutil.ReadAll(os.Stdin)
	checkError(err)

	// Connect and authenticate to the API endpoint:
	client := packngo.NewClient("", *flPktApiKey, nil)

	// Forge the request:
	createRequest := &packngo.DeviceCreateRequest{
		HostName:     *flPktHostName,
		Plan:         *flPktPlan,
		Facility:     *flPktFacility,
		OS:           *flPktOsys,
		BillingCycle: *flPktBilling,
		ProjectID:    *flPktProjId,
		UserData:     string(udata),
	}

	// Send the request:
	newDevice, _, err := client.Devices.Create(createRequest)
	checkError(err)

	// Response output:
	fmt.Println(newDevice)
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
