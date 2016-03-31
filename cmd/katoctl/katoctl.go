package main

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"fmt"
	"io/ioutil"
	"log"
	"os"

	// Local:
	"github.com/h0tbird/kato/udata"
	"github.com/h0tbird/kato/providers/pkt"
	"github.com/h0tbird/kato/providers/ec2"

	// Community:
	"gopkg.in/alecthomas/kingpin.v2"
)

//-----------------------------------------------------------------------------
// Package variable declarations factored into a block:
//-----------------------------------------------------------------------------

var (

	//----------------------------
	// katoctl: top level command
	//----------------------------

	app = kingpin.New("katoctl", "Katoctl defines and deploys CoreOS clusters.")

	flUdataFile = app.Flag("user-data", "Path to file containing user data.").
			PlaceHolder("KATO_USER_DATA").
			OverrideDefaultFromEnvar("KATO_USER_DATA").
			Short('u').String()

	//-----------------------
	// udata: nested command
	//-----------------------

	cmdUdata = app.Command("udata", "Generate CoreOS cloud-config user-data.")

	flUdataHostID = cmdUdata.Flag("hostid", "Must be a number: hostname = <role>-<hostid>").
			Required().PlaceHolder("KATO_UDATA_HOSTID").
			OverrideDefaultFromEnvar("KATO_UDATA_HOSTID").
			Short('i').String()

	flUdataDomain = cmdUdata.Flag("domain", "Domain name as in (hostname -d)").
			Required().PlaceHolder("KATO_UDATA_DOMAIN").
			OverrideDefaultFromEnvar("KATO_UDATA_DOMAIN").
			Short('d').String()

	flUdataRole = cmdUdata.Flag("role", "Choose one of [ master | node | edge ]").
			Required().PlaceHolder("KATO_UDATA_ROLE").
			OverrideDefaultFromEnvar("KATO_UDATA_ROLE").
			Short('r').String()

	flUdataNs1Apikey = cmdUdata.Flag("ns1-api-key", "NS1 private API key.").
			Required().PlaceHolder("KATO_UDATA_NS1_API_KEY").
			OverrideDefaultFromEnvar("KATO_UDATA_NS1_API_KEY").
			Short('k').String()

	flUdataCaCert = cmdUdata.Flag("ca-cert", "Path to CA certificate.").
			PlaceHolder("KATO_UDATA_CA_CERT").
			OverrideDefaultFromEnvar("KATO_UDATA_CA_CERT").
			Short('c').String()

	flUdataEtcdToken = cmdUdata.Flag("etcd-token", "Provide an etcd discovery token.").
			PlaceHolder("KATO_UDATA_ETCD_TOKEN").
			OverrideDefaultFromEnvar("KATO_UDATA_ETCD_TOKEN").
			Short('e').String()

	flUdataFlannelNetwork = cmdUdata.Flag("flannel-network", "Flannel entire overlay network.").
			PlaceHolder("KATO_UDATA_FLANNEL_NETWORK").
			OverrideDefaultFromEnvar("KATO_UDATA_FLANNEL_NETWORK").
			Short('n').String()

	flUdataFlannelSubnetLen = cmdUdata.Flag("flannel-subnet-len", "Subnet len to llocate to each host.").
			PlaceHolder("KATO_UDATA_FLANNEL_SUBNET_LEN").
			OverrideDefaultFromEnvar("KATO_UDATA_FLANNEL_SUBNET_LEN").
			Short('s').String()

	flUdataFlannelSubnetMin = cmdUdata.Flag("flannel-subnet-min", "Minimum subnet IP addresses.").
			PlaceHolder("KATO_UDATA_FLANNEL_SUBNET_MIN").
			OverrideDefaultFromEnvar("KATO_UDATA_FLANNEL_SUBNET_MIN").
			Short('m').String()

	flUdataFlannelSubnetMax = cmdUdata.Flag("flannel-subnet-max", "Maximum subnet IP addresses.").
			PlaceHolder("KATO_UDATA_FLANNEL_SUBNET_MAX").
			OverrideDefaultFromEnvar("KATO_UDATA_FLANNEL_SUBNET_MAX").
			Short('x').String()

	flUdataFlannelBackend = cmdUdata.Flag("flannel-backend", "Flannel backend type: [ udp | vxlan | host-gw | gce | aws-vpc | alloc ]").
			PlaceHolder("KATO_UDATA_FLANNEL_BACKEND").
			OverrideDefaultFromEnvar("KATO_UDATA_FLANNEL_BACKEND").
			Short('b').String()

	//------------------------------
	// setup-packet: nested command
	//------------------------------

	cmdSetupPacket = app.Command("setup-packet", "Setup a Packet.net project to be used by katoctl.")

	//----------------------------
	// run-packet: nested command
	//----------------------------

	cmdRunPacket = app.Command("run-packet", "Starts a CoreOS instance on Packet.net.")

	flRunPktAPIKey = cmdRunPacket.Flag("api-key", "Packet API key.").
			Required().PlaceHolder("KATO_RUN_PKT_APIKEY").
			OverrideDefaultFromEnvar("KATO_RUN_PKT_APIKEY").
			Short('k').String()

	flRunPktHostname = cmdRunPacket.Flag("hostname", "Used in the Packet.net dashboard.").
			Required().PlaceHolder("KATO_RUN_PKT_HOSTNAME").
			OverrideDefaultFromEnvar("KATO_RUN_PKT_HOSTNAME").
			Short('h').String()

	flRunPktProjectID = cmdRunPacket.Flag("project-id", "Format: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee").
			Required().PlaceHolder("KATO_RUN_PKT_PROJECT_ID").
			OverrideDefaultFromEnvar("KATO_RUN_PKT_PROJECT_ID").
			Short('i').String()

	flRunPktPlan = cmdRunPacket.Flag("plan", "One of [ baremetal_0 | baremetal_1 | baremetal_2 | baremetal_3 ]").
			Required().PlaceHolder("KATO_RUN_PKT_PLAN").
			OverrideDefaultFromEnvar("KATO_RUN_PKT_PLAN").
			Short('p').String()

	flRunPktOS = cmdRunPacket.Flag("os", "One of [ coreos_stable | coreos_beta | coreos_alpha ]").
			Required().PlaceHolder("KATO_RUN_PKT_OS").
			OverrideDefaultFromEnvar("KATO_RUN_PKT_OS").
			Short('o').String()

	flRunPktFacility = cmdRunPacket.Flag("facility", "One of [ ewr1 | ams1 ]").
			Required().PlaceHolder("KATO_RUN_PKT_FACILITY").
			OverrideDefaultFromEnvar("KATO_RUN_PKT_FACILITY").
			Short('f').String()

	flRunPktBilling = cmdRunPacket.Flag("billing", "One of [ hourly | monthly ]").
			Required().PlaceHolder("KATO_RUN_PKT_BILLING").
			OverrideDefaultFromEnvar("KATO_RUN_PKT_BILLING").
			Short('b').String()

	//---------------------------
	// setup-ec2: nested command
	//---------------------------

	cmdSetupEc2 = app.Command("setup-ec2", "Setup an EC2 VPC and all the related components.")

	flSetupEc2Region = cmdSetupEc2.Flag("region", "EC2 region.").
			Required().PlaceHolder("KATO_SETUP_EC2_REGION").
			OverrideDefaultFromEnvar("KATO_SETUP_EC2_REGION").
			Short('r').String()

	flSetupEc2VpcCidrBlock = cmdSetupEc2.Flag("vpc-cidr-block", "The range of IPs to be used by the VPC.").
			Required().PlaceHolder("KATO_SETUP_EC2_VPC_CIDR_BLOCK").
			OverrideDefaultFromEnvar("KATO_SETUP_EC2_VPC_CIDR_BLOCK").
			Short('c').String()

	flSetupEc2VpcNameTag = cmdSetupEc2.Flag("vpc-name-tag", "Used as Name tag for the VPC.").
			Required().PlaceHolder("KATO_SETUP_EC2_VPC_NAME_TAG").
			OverrideDefaultFromEnvar("KATO_SETUP_EC2_VPC_NAME_TAG").
			Short('n').String()

	flSetupEc2IntSubnetCidr = cmdSetupEc2.Flag("internal-subnet-cidr", "CIDR for the internal subnet.").
			Required().PlaceHolder("KATO_SETUP_EC2_INTERNAL_SUBNET_CIDR").
			OverrideDefaultFromEnvar("KATO_SETUP_EC2_INTERNAL_SUBNET_CIDR").
			Short('i').String()

	flSetupEc2ExtSubnetCidr = cmdSetupEc2.Flag("external-subnet-cidr", "CIDR for the external subnet.").
			Required().PlaceHolder("KATO_SETUP_EC2_EXTERNAL_SUBNET_CIDR").
			OverrideDefaultFromEnvar("KATO_SETUP_EC2_EXTERNAL_SUBNET_CIDR").
			Short('e').String()

	//-------------------------
	// run-ec2: nested command
	//-------------------------

	cmdRunEc2 = app.Command("run-ec2", "Starts a CoreOS instance on Amazon EC2.")

	flEc2HostName = cmdRunEc2.Flag("hostname", "For the EC2 dashboard.").
			PlaceHolder("EC2_HOSTNAME").
			OverrideDefaultFromEnvar("EC2_HOSTNAME").
			Short('h').String()

	flEc2Region = cmdRunEc2.Flag("region", "EC2 region.").
			Required().PlaceHolder("EC2_REGION").
			OverrideDefaultFromEnvar("EC2_REGION").
			Short('r').String()

	flEc2ImageID = cmdRunEc2.Flag("image-id", "EC2 image id.").
			Required().PlaceHolder("EC2_IMAGE_ID").
			OverrideDefaultFromEnvar("EC2_IMAGE_ID").
			Short('i').String()

	flEc2InsType = cmdRunEc2.Flag("instance-type", "EC2 instance type.").
			Required().PlaceHolder("EC2_INSTANCE_TYPE").
			OverrideDefaultFromEnvar("EC2_INSTANCE_TYPE").
			Short('t').String()

	flEc2KeyPair = cmdRunEc2.Flag("key-pair", "EC2 key pair.").
			Required().PlaceHolder("EC2_KEY_PAIR").
			OverrideDefaultFromEnvar("EC2_KEY_PAIR").
			Short('k').String()

	flEc2VpcID = cmdRunEc2.Flag("vpc-id", "EC2 VPC id.").
			Required().PlaceHolder("EC2_VPC_ID").
			OverrideDefaultFromEnvar("EC2_VPC_ID").
			Short('v').String()

	flEc2SubnetIds = cmdRunEc2.Flag("subnet-ids", "EC2 subnet ids.").
			Required().PlaceHolder("EC2_SUBNET_ID").
			OverrideDefaultFromEnvar("EC2_SUBNET_ID").
			Short('s').String()

	flEc2ElasticIP = cmdRunEc2.Flag("elastic-ip", "Allocate an elastic IP [ true | false ]").
			Default("false").PlaceHolder("EC2_ELASTIC_IP").
			OverrideDefaultFromEnvar("EC2_ELASTIC_IP").
			Short('e').String()
)

//----------------------------------------------------------------------------
// func init() is called after all the variable declarations in the package
// have evaluated their initializers, and those are evaluated only after all
// the imported packages have been initialized:
//----------------------------------------------------------------------------

func init() {

	// Change the flags on the default logger:
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

//----------------------------------------------------------------------------
// Entry point:
//----------------------------------------------------------------------------

func main() {

	// Sub-command selector:
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	//---------------
	// katoctl udata
	//---------------

	case cmdUdata.FullCommand():

		udata := udata.Data {
			HostID:           *flUdataHostID,
			Domain:           *flUdataDomain,
			Role:             *flUdataRole,
			Ns1ApiKey:        *flUdataNs1Apikey,
			CaCert:           *flUdataCaCert,
			EtcdToken:        *flUdataEtcdToken,
			FlannelNetwork:   *flUdataFlannelNetwork,
			FlannelSubnetLen: *flUdataFlannelSubnetLen,
			FlannelSubnetMin: *flUdataFlannelSubnetMin,
			FlannelSubnetMax: *flUdataFlannelSubnetMax,
			FlannelBackend:   *flUdataFlannelBackend,
		}

		err := udata.Render()
		checkError(err)

	//----------------------
	// katoctl setup-packet
	//----------------------

	case cmdSetupPacket.FullCommand():

		pkt := pkt.Data {
		}

		err := setup(&pkt)
		checkError(err)

	//--------------------
	// katoctl run-packet
	//--------------------

	case cmdRunPacket.FullCommand():

		pkt := pkt.Data {
			APIKey:    *flRunPktAPIKey,
			HostName:  *flRunPktHostname,
			ProjectID: *flRunPktProjectID,
			Plan:      *flRunPktPlan,
			OS:        *flRunPktOS,
			Facility:  *flRunPktFacility,
			Billing:   *flRunPktBilling,
		}

		err := run(&pkt)
		checkError(err)

	//-------------------
	// katoctl setup-ec2
	//-------------------

	case cmdSetupEc2.FullCommand():

		ec2 := ec2.Data {
			Region:        *flSetupEc2Region,
			VpcCidrBlock:  *flSetupEc2VpcCidrBlock,
			VpcNameTag:    *flSetupEc2VpcNameTag,
			IntSubnetCidr: *flSetupEc2IntSubnetCidr,
			ExtSubnetCidr: *flSetupEc2ExtSubnetCidr,
		}

		err := setup(&ec2)
		checkError(err)

	//-----------------
	// katoctl run-ec2
	//-----------------

	case cmdRunEc2.FullCommand():

		ec2 := ec2.Data {
			Region:    *flEc2Region,
			SubnetIds: *flEc2SubnetIds,
			ImageID:   *flEc2ImageID,
			KeyPair:   *flEc2KeyPair,
			InsType:   *flEc2InsType,
			HostName:  *flEc2HostName,
			ElasticIP: *flEc2ElasticIP,
		}

		err := run(&ec2)
		checkError(err)
	}
}

//--------------------------------------------------------------------------
// func: readUdata
//--------------------------------------------------------------------------

func readUdata() ([]byte, error) {

	// Read data from file:
	if *flUdataFile != "" {
		udata, err := ioutil.ReadFile(*flUdataFile)
		return udata, err
	}

	// Read data from stdin:
	udata, err := ioutil.ReadAll(os.Stdin)
	return udata, err
}

//---------------------------------------------------------------------------
// func: checkError
//---------------------------------------------------------------------------

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error: ", err.Error())
		os.Exit(1)
	}
}
