package main

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"io/ioutil"
	"os"

	// Local:
	"github.com/h0tbird/kato/providers/ec2"
	"github.com/h0tbird/kato/providers/pkt"
	"github.com/h0tbird/kato/udata"

	// Community:
	log "github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

//--------------------------------------------------------------------------
// Typedefs:
//--------------------------------------------------------------------------

type cloudProvider interface {
	Deploy() error
	Setup() error
	Run(udata []byte) error
}

//-----------------------------------------------------------------------------
// Package variable declarations factored into a block:
//-----------------------------------------------------------------------------

var (

	//-----------------------
	// katoctl: root command
	//-----------------------

	app = kingpin.New("katoctl", "Katoctl defines and deploys Kato's infrastructure.")

	//---------------------------
	// deploy: top level command
	//---------------------------

	cmdDeploy = app.Command("deploy", "Deploy Kato's infrastructure.")

	//--------------------------
	// setup: top level command
	//--------------------------

	cmdSetup = app.Command("setup", "Setup the IaaS provider.")

	//--------------------------
	// udata: top level command
	//--------------------------

	cmdUdata = app.Command("udata", "Generate CoreOS cloud-config user-data.")

	flUdataMasterCount = cmdUdata.Flag("master-count", "Number of master nodes [ 1 | 3 | 5 ]").
				Default("3").PlaceHolder("KATO_UDATA_MASTER_COUNT").
				OverrideDefaultFromEnvar("KATO_UDATA_MASTER_COUNT").
				Int()

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

	flUdataGzipUdata = cmdUdata.Flag("gzip-udata", "Enable udata compression.").
				Default("false").PlaceHolder("KATO_UDATA_GZIP_UDATA").
				OverrideDefaultFromEnvar("KATO_UDATA_GZIP_UDATA").
				Short('g').Bool()

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

	//------------------------
	// run: top level command
	//------------------------

	cmdRun = app.Command("run", "Starts a CoreOS instance.")

	flRunUserData = cmdRun.Flag("user-data", "Path to file containing user data.").
			PlaceHolder("KATO_RUN_USER_DATA").
			OverrideDefaultFromEnvar("KATO_RUN_USER_DATA").
			Short('u').String()

	//-------------------------------
	// deploy packet: nested command
	//-------------------------------

	cmdDeployPacket = cmdDeploy.Command("packet", "Deploy Kato's infrastructure on Packet.net")

	//------------------------------
	// setup packet: nested command
	//------------------------------

	cmdSetupPacket = cmdSetup.Command("packet", "Setup a Packet.net project to be used by katoctl.")

	//----------------------------
	// run packet: nested command
	//----------------------------

	cmdRunPacket = cmdRun.Command("packet", "Starts a CoreOS instance on Packet.net.")

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

	//----------------------------
	// deploy ec2: nested command
	//----------------------------

	cmdDeployEc2 = cmdDeploy.Command("ec2", "Deploy Kato's infrastructure on Amazon EC2.")

	flDeployEc2MasterCount = cmdDeployEc2.Flag("master-count", "Number of master nodes to deploy [ 1 | 3 ]").
				Required().PlaceHolder("KATO_DEPLOY_EC2_MASTER_COUNT").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_MASTER_COUNT").
				Short('m').Int()

	flDeployEc2NodeCount = cmdDeployEc2.Flag("node-count", "Number of worker nodes to deploy.").
				Required().PlaceHolder("KATO_DEPLOY_EC2_NODE_COUNT").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_NODE_COUNT").
				Short('n').Int()

	flDeployEc2EdgeCount = cmdDeployEc2.Flag("edge-count", "Number of edge nodes to deploy.").
				Required().PlaceHolder("KATO_DEPLOY_EC2_EDGE_COUNT").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_EDGE_COUNT").
				Short('e').Int()

	flDeployEc2MasterType = cmdDeployEc2.Flag("master-type", "EC2 master instance type.").
				Default("t2.medium").PlaceHolder("KATO_DEPLOY_EC2_MASTER_TYPE").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_MASTER_TYPE").
				String()

	flDeployEc2NodeType = cmdDeployEc2.Flag("node-type", "EC2 node instance type.").
				Default("t2.medium").PlaceHolder("KATO_DEPLOY_EC2_NODE_TYPE").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_NODE_TYPE").
				String()

	flDeployEc2EdgeType = cmdDeployEc2.Flag("edge-type", "EC2 edge instance type.").
				Default("t2.medium").PlaceHolder("KATO_DEPLOY_EC2_EDGE_TYPE").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_EDGE_TYPE").
				String()

	flDeployEc2EtcdToken = cmdDeployEc2.Flag("etcd-token", "Etcd bootstrap token [ auto | <token> ]").
				Required().PlaceHolder("KATO_DEPLOY_EC2_ETCD_TOKEN").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_ETCD_TOKEN").
				Short('t').String()

	flDeployEc2Ns1ApiKey = cmdDeployEc2.Flag("ns1-api-key", "NS1 private API key.").
				Required().PlaceHolder("KATO_DEPLOY_EC2_NS1_API_KEY").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_NS1_API_KEY").
				String()

	flDeployEc2CaCert = cmdDeployEc2.Flag("ca-cert", "Path to CA certificate.").
				Required().PlaceHolder("KATO_DEPLOY_EC2_CA_CET").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_CA_CET").
				Short('c').String()

	flDeployEc2Region = cmdDeployEc2.Flag("region", "Amazon EC2 region.").
				Required().PlaceHolder("KATO_DEPLOY_EC2_REGION").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_REGION").
				Short('r').String()

	flDeployEc2Domain = cmdDeployEc2.Flag("domain", "Used to identify the VPC.").
				Required().PlaceHolder("KATO_DEPLOY_EC2_DOMAIN").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_DOMAIN").
				Short('d').String()

	flDeployEc2KeyPair = cmdDeployEc2.Flag("key-pair", "EC2 key pair.").
				Required().PlaceHolder("KATO_DEPLOY_EC2_KEY_PAIR").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_KEY_PAIR").
				Short('k').String()

	//---------------------------
	// setup ec2: nested command
	//---------------------------

	cmdSetupEc2 = cmdSetup.Command("ec2", "Setup an EC2 VPC and all the related components.")

	flSetupEc2Region = cmdSetupEc2.Flag("region", "EC2 region.").
				Required().PlaceHolder("KATO_SETUP_EC2_REGION").
				OverrideDefaultFromEnvar("KATO_SETUP_EC2_REGION").
				Short('r').String()

	flSetupEc2VpcCidrBlock = cmdSetupEc2.Flag("vpc-cidr-block", "IPs to be used by the VPC (default: 10.0.0.0/16).").
				Default("10.0.0.0/16").PlaceHolder("KATO_SETUP_EC2_VPC_CIDR_BLOCK").
				OverrideDefaultFromEnvar("KATO_SETUP_EC2_VPC_CIDR_BLOCK").
				Short('c').String()

	flSetupEc2Domain = cmdSetupEc2.Flag("domain", "Used to identify the VPC..").
				Required().PlaceHolder("KATO_SETUP_EC2_DOMAIN").
				OverrideDefaultFromEnvar("KATO_SETUP_EC2_DOMAIN").
				Short('t').String()

	flSetupEc2IntSubnetCidr = cmdSetupEc2.Flag("internal-subnet-cidr", "CIDR for the internal subnet (default: 10.0.1.0/24).").
				Default("10.0.1.0/24").PlaceHolder("KATO_SETUP_EC2_INTERNAL_SUBNET_CIDR").
				OverrideDefaultFromEnvar("KATO_SETUP_EC2_INTERNAL_SUBNET_CIDR").
				Short('i').String()

	flSetupEc2ExtSubnetCidr = cmdSetupEc2.Flag("external-subnet-cidr", "CIDR for the external subnet (default: 10.0.0.0/24).").
				Default("10.0.0.0/24").PlaceHolder("KATO_SETUP_EC2_EXTERNAL_SUBNET_CIDR").
				OverrideDefaultFromEnvar("KATO_SETUP_EC2_EXTERNAL_SUBNET_CIDR").
				Short('e').String()

	//-------------------------
	// run ec2: nested command
	//-------------------------

	cmdRunEc2 = cmdRun.Command("ec2", "Starts a CoreOS instance on Amazon EC2.")

	flRunEc2Hostname = cmdRunEc2.Flag("hostname", "For the EC2 dashboard.").
				PlaceHolder("KATO_RUN_EC2_HOSTNAME").
				OverrideDefaultFromEnvar("KATO_RUN_EC2_HOSTNAME").
				Short('h').String()

	flRunEc2Region = cmdRunEc2.Flag("region", "EC2 region.").
			Required().PlaceHolder("KATO_RUN_EC2_REGION").
			OverrideDefaultFromEnvar("KATO_RUN_EC2_REGION").
			Short('r').String()

	flRunEc2ImageID = cmdRunEc2.Flag("image-id", "EC2 image id.").
			Required().PlaceHolder("KATO_RUN_EC2_IMAGE_ID").
			OverrideDefaultFromEnvar("KATO_RUN_EC2_IMAGE_ID").
			Short('i').String()

	flRunEc2InsType = cmdRunEc2.Flag("instance-type", "EC2 instance type.").
			Required().PlaceHolder("KATO_RUN_EC2_INSTANCE_TYPE").
			OverrideDefaultFromEnvar("KATO_RUN_EC2_INSTANCE_TYPE").
			Short('t').String()

	flRunEc2KeyPair = cmdRunEc2.Flag("key-pair", "EC2 key pair.").
			Required().PlaceHolder("KATO_RUN_EC2_KEY_PAIR").
			OverrideDefaultFromEnvar("KATO_RUN_EC2_KEY_PAIR").
			Short('k').String()

	flRunEc2SubnetIDs = cmdRunEc2.Flag("subnet-ids", "EC2 subnet ids.").
				Required().PlaceHolder("KATO_RUN_EC2_SUBNET_ID").
				OverrideDefaultFromEnvar("KATO_RUN_EC2_SUBNET_ID").
				Short('s').String()

	flRunEc2ElasticIP = cmdRunEc2.Flag("elastic-ip", "Allocate an elastic IP [ true | false ]").
				Default("false").PlaceHolder("KATO_RUN_EC2_ELASTIC_IP").
				OverrideDefaultFromEnvar("KATO_RUN_EC2_ELASTIC_IP").
				Short('e').String()
)

//----------------------------------------------------------------------------
// func init() is called after all the variable declarations in the package
// have evaluated their initializers, and those are evaluated only after all
// the imported packages have been initialized:
//----------------------------------------------------------------------------

func init() {

	// Customize the default logger:
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetOutput(os.Stderr)
	log.SetLevel(log.InfoLevel)
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

		udata := udata.Data{
			MasterCount:      *flUdataMasterCount,
			HostID:           *flUdataHostID,
			Domain:           *flUdataDomain,
			Role:             *flUdataRole,
			Ns1ApiKey:        *flUdataNs1Apikey,
			CaCert:           *flUdataCaCert,
			EtcdToken:        *flUdataEtcdToken,
			GzipUdata:        *flUdataGzipUdata,
			FlannelNetwork:   *flUdataFlannelNetwork,
			FlannelSubnetLen: *flUdataFlannelSubnetLen,
			FlannelSubnetMin: *flUdataFlannelSubnetMin,
			FlannelSubnetMax: *flUdataFlannelSubnetMax,
			FlannelBackend:   *flUdataFlannelBackend,
		}

		err := udata.Render()
		checkError(err)

	//-----------------------
	// katoctl deploy packet
	//-----------------------

	case cmdDeployPacket.FullCommand():

		pkt := pkt.Data{}
		err := pkt.Deploy()
		checkError(err)

	//----------------------
	// katoctl setup packet
	//----------------------

	case cmdSetupPacket.FullCommand():

		pkt := pkt.Data{}
		err := pkt.Setup()
		checkError(err)

	//--------------------
	// katoctl run packet
	//--------------------

	case cmdRunPacket.FullCommand():

		pkt := pkt.Data{
			APIKey:    *flRunPktAPIKey,
			HostName:  *flRunPktHostname,
			ProjectID: *flRunPktProjectID,
			Plan:      *flRunPktPlan,
			OS:        *flRunPktOS,
			Facility:  *flRunPktFacility,
			Billing:   *flRunPktBilling,
		}

		udata, err := readUdata()
		checkError(err)
		err = pkt.Run(udata)
		checkError(err)

	//--------------------
	// katoctl deploy ec2
	//--------------------

	case cmdDeployEc2.FullCommand():

		ec2 := ec2.Data{
			MasterCount: *flDeployEc2MasterCount,
			NodeCount:   *flDeployEc2NodeCount,
			EdgeCount:   *flDeployEc2EdgeCount,
			MasterType:  *flDeployEc2MasterType,
			NodeType:    *flDeployEc2NodeType,
			EdgeType:    *flDeployEc2EdgeType,
			EtcdToken:   *flDeployEc2EtcdToken,
			Ns1ApiKey:   *flDeployEc2Ns1ApiKey,
			CaCert:      *flDeployEc2CaCert,
			Domain:      *flDeployEc2Domain,
			Region:      *flDeployEc2Region,
			KeyPair:     *flDeployEc2KeyPair,
		}

		err := ec2.Deploy()
		checkError(err)

	//-------------------
	// katoctl setup ec2
	//-------------------

	case cmdSetupEc2.FullCommand():

		ec2 := ec2.Data{
			Domain:             *flSetupEc2Domain,
			Region:             *flSetupEc2Region,
			VpcCidrBlock:       *flSetupEc2VpcCidrBlock,
			InternalSubnetCidr: *flSetupEc2IntSubnetCidr,
			ExternalSubnetCidr: *flSetupEc2ExtSubnetCidr,
		}

		err := ec2.Setup()
		checkError(err)

	//-----------------
	// katoctl run ec2
	//-----------------

	case cmdRunEc2.FullCommand():

		ec2 := ec2.Data{
			Region:       *flRunEc2Region,
			SubnetIDs:    *flRunEc2SubnetIDs,
			ImageID:      *flRunEc2ImageID,
			KeyPair:      *flRunEc2KeyPair,
			InstanceType: *flRunEc2InsType,
			Hostname:     *flRunEc2Hostname,
			ElasticIP:    *flRunEc2ElasticIP,
		}

		udata, err := readUdata()
		checkError(err)
		err = ec2.Run(udata)
		checkError(err)
	}
}

//--------------------------------------------------------------------------
// func: readUdata
//--------------------------------------------------------------------------

func readUdata() ([]byte, error) {

	// Read data from file:
	if *flRunUserData != "" {
		udata, err := ioutil.ReadFile(*flRunUserData)
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
		log.WithField("cmd", "katoctl").Error(err)
		os.Exit(1)
	}
}
