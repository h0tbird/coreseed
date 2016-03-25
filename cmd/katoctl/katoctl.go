//-----------------------------------------------------------------------------
// Package membership:
//-----------------------------------------------------------------------------

package main

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Standard library:
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	// Community:
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/packethost/packngo"
	"gopkg.in/alecthomas/kingpin.v2"
)

//----------------------------------------------------------------------------
// Typedefs:
//----------------------------------------------------------------------------

type userData struct {
	HostId           string
	Domain           string
	Role             string
	Ns1ApiKey        string
	FleetTags        string
	CaCert           string
	EtcdToken        string
	FlannelNetwork   string
	FlannelSubnetLen string
	FlannelSubnetMin string
	FlannelSubnetMax string
	FlannelBackend   string
}

//-----------------------------------------------------------------------------
// Package variable declarations factored into a block:
//-----------------------------------------------------------------------------

var (

	//----------------------------
	// katoctl: top level command
	//----------------------------

	app = kingpin.New("katoctl", "Katoctl defines and deploys CoreOS clusters.")

	//-----------------------
	// udata: nested command
	//-----------------------

	cmdUdata = app.Command("udata", "Generate CoreOS cloud-config user-data.")

	flHostID = cmdUdata.Flag("hostid", "hostname = role-id").
			Required().PlaceHolder("CS_HOSTID").
			OverrideDefaultFromEnvar("CS_HOSTID").
			Short('i').String()

	flDomain = cmdUdata.Flag("domain", "Domain name as in (hostname -d)").
			Required().PlaceHolder("CS_DOMAIN").
			OverrideDefaultFromEnvar("CS_DOMAIN").
			Short('d').String()

	flRole = cmdUdata.Flag("role", "Choose one of [ master | node | edge ]").
			Required().PlaceHolder("CS_ROLE").
			OverrideDefaultFromEnvar("CS_ROLE").
			Short('r').String()

	flNs1Apikey = cmdUdata.Flag("ns1-api-key", "NS1 private API key.").
			Required().PlaceHolder("CS_NS1_API_KEY").
			OverrideDefaultFromEnvar("CS_NS1_API_KEY").
			Short('k').String()

	flFleetTags = cmdUdata.Flag("fleet-tags", "Comma separated list of fleet tags.").
			PlaceHolder("CS_FLEET_TAGS").
			OverrideDefaultFromEnvar("CS_FLEET_TAGS").
			Short('t').String()

	flCAcert = cmdUdata.Flag("ca-cert", "Path to CA certificate.").
			PlaceHolder("CS_CA_CERT").
			OverrideDefaultFromEnvar("CS_CA_CERT").
			Short('c').String()

	flEtcdToken = cmdUdata.Flag("etcd-token", "Provide an etcd discovery token.").
			PlaceHolder("CS_ETCD_TOKEN").
			OverrideDefaultFromEnvar("CS_ETCD_TOKEN").
			Short('e').String()

	flFlannelNetwork = cmdUdata.Flag("flannel-network", "Flannel entire overlay network.").
			PlaceHolder("CS_FLANNEL_NETWORK").
			OverrideDefaultFromEnvar("CS_FLANNEL_NETWORK").
			Short('n').String()

	flFlannelSubnetLen = cmdUdata.Flag("flannel-subnet-len", "Subnet len to llocate to each host.").
			PlaceHolder("CS_FLANNEL_SUBNET_LEN").
			OverrideDefaultFromEnvar("CS_FLANNEL_SUBNET_LEN").
			Short('s').String()

	flFlannelSubnetMin = cmdUdata.Flag("flannel-subnet-min", "Minimum subnet IP addresses.").
			PlaceHolder("CS_FLANNEL_SUBNET_MIN").
			OverrideDefaultFromEnvar("CS_FLANNEL_SUBNET_MIN").
			Short('m').String()

	flFlannelSubnetMax = cmdUdata.Flag("flannel-subnet-max", "Maximum subnet IP addresses.").
			PlaceHolder("CS_FLANNEL_SUBNET_MAX").
			OverrideDefaultFromEnvar("CS_FLANNEL_SUBNET_MAX").
			Short('x').String()

	flFlannelBackend = cmdUdata.Flag("flannel-backend", "Flannel backend type: [ udp | vxlan | host-gw | gce | aws-vpc | alloc ]").
			PlaceHolder("CS_FLANNEL_SUBNET_MAX").
			OverrideDefaultFromEnvar("CS_FLANNEL_SUBNET_MAX").
			Short('b').String()

	//----------------------------
	// run-packet: nested command
	//----------------------------

	cmdRunPacket = app.Command("run-packet", "Starts a CoreOS instance on Packet.net")

	flPktAPIKey = cmdRunPacket.Flag("api-key", "Packet API key.").
			Required().PlaceHolder("PKT_APIKEY").
			OverrideDefaultFromEnvar("PKT_APIKEY").
			Short('k').String()

	flPktHostName = cmdRunPacket.Flag("hostname", "For the Packet.net dashboard.").
			Required().PlaceHolder("PKT_HOSTNAME").
			OverrideDefaultFromEnvar("PKT_HOSTNAME").
			Short('h').String()

	flPktProjID = cmdRunPacket.Flag("project-id", "Format: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee").
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

	//-------------------------
	// run-ec2: nested command
	//-------------------------

	cmdRunEc2 = app.Command("run-ec2", "Starts a CoreOS instance on Amazon EC2")

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
// Entry point:
//----------------------------------------------------------------------------

func main() {

	// Sub-command selector:
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	// katoctl udata ...
	case cmdUdata.FullCommand():
		udata()

	// katoctl run-packet ...
	case cmdRunPacket.FullCommand():
		runPacket()

	// katoctl run-ec2 ...
	case cmdRunEc2.FullCommand():
		runEc2()
	}
}

//--------------------------------------------------------------------------
// func: udata
//--------------------------------------------------------------------------

func udata() {

	// Template udata structure:
	udata := userData {
		HostId:           *flHostID,
		Domain:           *flDomain,
		Role:             *flRole,
		Ns1ApiKey:        *flNs1Apikey,
		FleetTags:        *flFleetTags,
		EtcdToken:        *flEtcdToken,
		FlannelNetwork:   *flFlannelNetwork,
		FlannelSubnetLen: *flFlannelSubnetLen,
		FlannelSubnetMin: *flFlannelSubnetMin,
		FlannelSubnetMax: *flFlannelSubnetMax,
		FlannelBackend:   *flFlannelBackend,
	}

	// Read the CA certificate:
	if *flCAcert != "" {
		dat, err := ioutil.ReadFile(*flCAcert)
		checkError(err)
		udata.CaCert = strings.TrimSpace(strings.Replace(string(dat), "\n", "\n    ", -1))
	}

	// Render the template for the selected role:
	switch *flRole {
	case "master":
		t := template.New("master_udata")
		t, err := t.Parse(templMaster)
		err = t.Execute(os.Stdout, udata)
		checkError(err)
	case "node":
		t := template.New("node_udata")
		t, err := t.Parse(templNode)
		err = t.Execute(os.Stdout, udata)
		checkError(err)
	case "edge":
		t := template.New("edge_udata")
		t, err := t.Parse(templEdge)
		err = t.Execute(os.Stdout, udata)
		checkError(err)
	}
}

//--------------------------------------------------------------------------
// func: runPacket
//--------------------------------------------------------------------------

func runPacket() {

	// Read user-data from stdin:
	udata, err := ioutil.ReadAll(os.Stdin)
	checkError(err)

	// Connect and authenticate to the API endpoint:
	client := packngo.NewClient("", *flPktAPIKey, nil)

	// Forge the request:
	createRequest := &packngo.DeviceCreateRequest{
		HostName:     *flPktHostName,
		Plan:         *flPktPlan,
		Facility:     *flPktFacility,
		OS:           *flPktOsys,
		BillingCycle: *flPktBilling,
		ProjectID:    *flPktProjID,
		UserData:     string(udata),
	}

	// Send the request:
	newDevice, _, err := client.Devices.Create(createRequest)
	checkError(err)

	// Pretty-print the response data:
	fmt.Println(newDevice)
}

//--------------------------------------------------------------------------
// func: runEc2
//--------------------------------------------------------------------------

func runEc2() {

	// Read user-data from stdin:
	udata, err := ioutil.ReadAll(os.Stdin)
	checkError(err)

	// Connect and authenticate to the API endpoint:
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(*flEc2Region)}))

	// Forge the network interfaces:
	var networkInterfaces []*ec2.InstanceNetworkInterfaceSpecification
	subnetIds := strings.Split(*flEc2SubnetIds, ",")

	for i := 0; i < len(subnetIds); i++ {

		// Forge the security group ids:
		var securityGroupIds []*string
		for _, gid := range strings.Split(subnetIds[i], ":")[1:] {
			securityGroupIds = append(securityGroupIds, aws.String(gid))
		}

		iface := ec2.InstanceNetworkInterfaceSpecification{
			DeleteOnTermination: aws.Bool(true),
			DeviceIndex:         aws.Int64(int64(i)),
			Groups:              securityGroupIds,
			SubnetId:            aws.String(strings.Split(subnetIds[i], ":")[0]),
		}

		networkInterfaces = append(networkInterfaces, &iface)
	}

	// Send the request:
	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:           aws.String(*flEc2ImageID),
		MinCount:          aws.Int64(1),
		MaxCount:          aws.Int64(1),
		KeyName:           aws.String(*flEc2KeyPair),
		InstanceType:      aws.String(*flEc2InsType),
		NetworkInterfaces: networkInterfaces,
		UserData:          aws.String(base64.StdEncoding.EncodeToString([]byte(udata))),
	})

	checkError(err)

	// Pretty-print the response data:
	fmt.Println("Created instance", *runResult.Instances[0].InstanceId)

	// Add tags to the created instance:
	_, err = svc.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{runResult.Instances[0].InstanceId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(*flEc2HostName),
			},
		},
	})

	checkError(err)

	// Allocate an elastic IP address:
	if *flEc2ElasticIP == "true" {

		params := &ec2.AllocateAddressInput{
			Domain: aws.String("vpc"),
			DryRun: aws.Bool(false),
		}

		// Send the request:
		resp, err := svc.AllocateAddress(params)
		checkError(err)

		// Pretty-print the response data:
		fmt.Println(resp)
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
