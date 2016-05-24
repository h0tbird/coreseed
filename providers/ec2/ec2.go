package ec2

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"

	// Community:
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/h0tbird/kato/katool"
)

//-----------------------------------------------------------------------------
// Typedefs:
//-----------------------------------------------------------------------------

// AWS API endpoints:
type svc struct {
	ec2 *ec2.EC2
	iam *iam.IAM
}

// Deployment state data:
type State struct {
	MasterCount      int    //  ec2:deploy |           |       |         |
	NodeCount        int    //  ec2:deploy |           |       |         |
	EdgeCount        int    //  ec2:deploy |           |       |         |
	MasterType       string //  ec2:deploy |           |       |         |
	NodeType         string //  ec2:deploy |           |       |         |
	EdgeType         string //  ec2:deploy |           |       |         |
	Channel          string //  ec2:deploy |           |       |         |
	EtcdToken        string //  ec2:deploy |           | udata |         |
	Ns1ApiKey        string //  ec2:deploy |           | udata |         |
	CaCert           string //  ec2:deploy |           | udata |         |
	FlannelNetwork   string //  ec2:deploy |           | udata |         |
	FlannelSubnetLen string //  ec2:deploy |           | udata |         |
	FlannelSubnetMin string //  ec2:deploy |           | udata |         |
	FlannelSubnetMax string //  ec2:deploy |           | udata |         |
	FlannelBackend   string //  ec2:deploy |           | udata |         |
	Domain           string //  ec2:deploy | ec2:setup | udata |         |
	ClusterID        string //  ec2:deploy | ec2:setup |       |         |
	Region           string //  ec2:deploy | ec2:setup |       | ec2:run |
	Zone             string //  ec2:deploy | ec2:setup |       | ec2:run |
	command          string //  ec2:deploy | ec2:setup |       | ec2:run |
	VpcCidrBlock     string //  ec2:deploy | ec2:setup |       |         |
	IntSubnetCidr    string //  ec2:deploy | ec2:setup |       |         |
	ExtSubnetCidr    string //  ec2:deploy | ec2:setup |       |         |
	vpcID            string //             | ec2:setup |       |         |
	mainRouteTableID string //             | ec2:setup |       |         |
	inetGatewayID    string //             | ec2:setup |       |         |
	natGatewayID     string //             | ec2:setup |       |         |
	routeTableID     string //             | ec2:setup |       |         |
	masterRoleID     string //             | ec2:setup |       |         |
	nodeRoleID       string //             | ec2:setup |       |         |
	edgeRoleID       string //             | ec2:setup |       |         |
	rexrayPolicyARN  string //             | ec2:setup |       |         |
	masterSecGrp     string //             | ec2:setup |       |         |
	nodeSecGrp       string //             | ec2:setup |       |         |
	edgeSecGrp       string //             | ec2:setup |       |         |
	IntSubnetID      string //             | ec2:setup |       |         |
	ExtSubnetID      string //             | ec2:setup |       |         |
	allocationID     string //             | ec2:setup |       | ec2:run |
	instanceID       string //             |           |       | ec2:run |
	SubnetID         string //             |           |       | ec2:run |
	SecGrpID         string //             |           |       | ec2:run |
	ImageID          string //             |           |       | ec2:run |
	KeyPair          string //             |           |       | ec2:run |
	InstanceType     string //             |           |       | ec2:run |
	Hostname         string //             |           |       | ec2:run |
	PublicIP         string //             |           |       | ec2:run |
	IAMRole          string //             |           |       | ec2:run |
	SrcDstCheck      string //             |           |       | ec2:run |
	interfaceID      string //             |           |       | ec2:run |
	Role             string //             |           |       |         | ec2:add
	ID               string //             |           |       |         | ec2:add
}

// Service endpoints and deployment state.
type Data struct {
	svc
	State
}

//-----------------------------------------------------------------------------
// func: Deploy
//-----------------------------------------------------------------------------

// Deploy Kato's infrastructure on Amazon EC2.
func (d *Data) Deploy() error {

	// Set command to deploy:
	d.command = "deploy"

	// Setup a wait group:
	var wg sync.WaitGroup

	// Setup the environment:
	wg.Add(3)
	go d.environmentSetup(&wg)
	go d.retrieveEtcdToken(&wg)
	go d.retrieveCoreosAmiID(&wg)
	wg.Wait()

	// Whether or not to source-dest-check:
	if d.FlannelBackend == "host-gw" {
		d.SrcDstCheck = "false"
	} else {
		d.SrcDstCheck = "true"
	}

	// Dump state to file:
	if err := d.dumpState(); err != nil {
		return err
	}

	// Deploy all the nodes:
	wg.Add(3)
	go d.deployNodes("master", d.MasterCount, &wg)
	go d.deployNodes("node", d.NodeCount, &wg)
	go d.deployNodes("edge", d.EdgeCount, &wg)
	wg.Wait()

	return nil
}

//-----------------------------------------------------------------------------
// func: Add
//-----------------------------------------------------------------------------

// Adds a new instance to the cluster.
func (d *Data) Add() error {

	// Set command to add:
	d.command = "add"

	// Load data from state file:
	dat, err := katool.LoadState(d.ClusterID)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Store the values:
	d.vpcID = dat["VpcID"].(string)
	d.MasterCount = int(dat["MasterCount"].(float64))
	d.VpcCidrBlock = dat["VpcCidrBlock"].(string)
	d.IntSubnetCidr = dat["IntSubnetCidr"].(string)
	d.ExtSubnetCidr = dat["ExtSubnetCidr"].(string)
	d.mainRouteTableID = dat["MainRouteTableID"].(string)
	d.IntSubnetID = dat["IntSubnetID"].(string)
	d.ExtSubnetID = dat["ExtSubnetID"].(string)
	d.inetGatewayID = dat["InetGatewayID"].(string)
	d.allocationID = dat["AllocationID"].(string)
	d.natGatewayID = dat["NatGatewayID"].(string)
	d.routeTableID = dat["RouteTableID"].(string)
	d.masterSecGrp = dat["MasterSecGrp"].(string)
	d.nodeSecGrp = dat["NodeSecGrp"].(string)
	d.edgeSecGrp = dat["EdgeSecGrp"].(string)
	d.Domain = dat["Domain"].(string)
	d.Ns1ApiKey = dat["Ns1ApiKey"].(string)
	d.CaCert = dat["CaCert"].(string)
	d.EtcdToken = dat["EtcdToken"].(string)
	d.FlannelNetwork = dat["FlannelNetwork"].(string)
	d.FlannelSubnetLen = dat["FlannelSubnetLen"].(string)
	d.FlannelSubnetMin = dat["FlannelSubnetMin"].(string)
	d.FlannelSubnetMax = dat["FlannelSubnetMax"].(string)
	d.FlannelBackend = dat["FlannelBackend"].(string)
	d.Region = dat["Region"].(string)
	d.Zone = dat["Zone"].(string)
	d.ImageID = dat["ImageID"].(string)
	d.MasterType = dat["MasterType"].(string)
	d.NodeType = dat["NodeType"].(string)
	d.EdgeType = dat["EdgeType"].(string)
	d.KeyPair = dat["KeyPair"].(string)
	d.SrcDstCheck = dat["SrcDstCheck"].(string)

	// Forge the udata command:
	cmdUdata := exec.Command("katoctl", "udata",
		"--role", d.Role,
		"--master-count", strconv.Itoa(d.MasterCount),
		"--hostid", d.ID,
		"--domain", d.Domain,
		"--ns1-api-key", d.Ns1ApiKey,
		"--ca-cert", d.CaCert,
		"--etcd-token", d.EtcdToken,
		"--flannel-network", d.FlannelNetwork,
		"--flannel-subnet-len", d.FlannelSubnetLen,
		"--flannel-subnet-min", d.FlannelSubnetMin,
		"--flannel-subnet-max", d.FlannelSubnetMax,
		"--flannel-backend", d.FlannelBackend,
		"--rexray-storage-driver", "ec2",
		"--gzip-udata")

	// Forge the run command:
	var cmdRun *exec.Cmd

	switch d.Role {
	case "master":
		cmdRun = exec.Command("katoctl", "ec2", "run",
			"--hostname", d.Role+"-"+d.ID+"."+d.Domain,
			"--region", d.Region,
			"--zone", d.Zone,
			"--image-id", d.ImageID,
			"--instance-type", d.MasterType,
			"--key-pair", d.KeyPair,
			"--subnet-id", d.IntSubnetID,
			"--security-group-id", d.masterSecGrp,
			"--iam-role", d.Role,
			"--source-dest-check", d.SrcDstCheck,
			"--public-ip", "false")

	case "node":
		cmdRun = exec.Command("katoctl", "ec2", "run",
			"--hostname", d.Role+"-"+d.ID+"."+d.Domain,
			"--region", d.Region,
			"--zone", d.Zone,
			"--image-id", d.ImageID,
			"--instance-type", d.NodeType,
			"--key-pair", d.KeyPair,
			"--subnet-id", d.ExtSubnetID,
			"--security-group-id", d.nodeSecGrp,
			"--iam-role", d.Role,
			"--source-dest-check", d.SrcDstCheck,
			"--public-ip", "true")

	case "edge":
		cmdRun = exec.Command("katoctl", "ec2", "run",
			"--hostname", d.Role+"-"+d.ID+"."+d.Domain,
			"--region", d.Region,
			"--zone", d.Zone,
			"--image-id", d.ImageID,
			"--instance-type", d.EdgeType,
			"--key-pair", d.KeyPair,
			"--subnet-id", d.ExtSubnetID,
			"--security-group-id", d.edgeSecGrp,
			"--iam-role", d.Role,
			"--source-dest-check", d.SrcDstCheck,
			"--public-ip", "true")
	}

	// Execute the pipeline:
	if err := katool.ExecutePipeline(cmdUdata, cmdRun); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: Run
//-----------------------------------------------------------------------------

// Run uses EC2 API to launch a new instance.
func (d *Data) Run(udata []byte) error {

	// Set command to run:
	d.command = "run"

	// Connect and authenticate to the API endpoint:
	d.svc.ec2 = ec2.New(session.New(&aws.Config{Region: aws.String(d.Region)}))

	// Run the EC2 instance:
	if err := d.runInstance(udata); err != nil {
		return err
	}

	// Modify instance attributes:
	if err := d.modifyInstanceAttribute(); err != nil {
		return err
	}

	if d.PublicIP == "elastic" {

		// Allocate an elastic IP address:
		if err := d.allocateElasticIP(); err != nil {
			return err
		}

		// Associate the elastic IP:
		if err := d.associateElasticIP(); err != nil {
			return err
		}
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: Setup
//-----------------------------------------------------------------------------

// Setup VPC, IAM and EC2 components.
func (d *Data) Setup() error {

	// Set current command:
	d.command = "setup"

	// Connect and authenticate to the API endpoints:
	log.WithField("cmd", "ec2:"+d.command).
		Info("Connecting to region " + d.Region)
	d.svc.ec2 = ec2.New(session.New(&aws.Config{Region: aws.String(d.Region)}))
	d.svc.iam = iam.New(session.New())

	// Create the VPC:
	if err := d.createVpc(); err != nil {
		return err
	}

	// Setup a wait group:
	var wg sync.WaitGroup
	wg.Add(3)

	// Setup VPC, IAM and EC2:
	go d.setupVPCNetwork(&wg)
	go d.setupIAMSecurity(&wg)
	go d.setupEC2Firewall(&wg)

	// Wait to proceed:
	wg.Wait()

	// Dump state to file:
	if err := d.dumpState(); err != nil {
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: environmentSetup
//-----------------------------------------------------------------------------

func (d *Data) environmentSetup(wg *sync.WaitGroup) {

	// Decrement:
	defer wg.Done()

	// Forge the setup command:
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.Domain}).
		Info("Setup the EC2 environment")

	cmdSetup := exec.Command("katoctl", "ec2", "setup",
		"--cluster-id", d.ClusterID,
		"--domain", d.Domain,
		"--region", d.Region,
		"--zone", d.Zone,
		"--vpc-cidr-block", d.VpcCidrBlock,
		"--internal-subnet-cidr", d.IntSubnetCidr,
		"--external-subnet-cidr", d.ExtSubnetCidr)

	// Execute the setup command:
	cmdSetup.Stderr = os.Stderr
	if err := cmdSetup.Run(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		os.Exit(1)
	}

	// Load data from state file:
	dat, err := katool.LoadState(d.ClusterID)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		os.Exit(1)
	}

	// Store the values:
	d.vpcID = dat["VpcID"].(string)
	d.VpcCidrBlock = dat["VpcCidrBlock"].(string)
	d.IntSubnetCidr = dat["IntSubnetCidr"].(string)
	d.ExtSubnetCidr = dat["ExtSubnetCidr"].(string)
	d.mainRouteTableID = dat["MainRouteTableID"].(string)
	d.IntSubnetID = dat["IntSubnetID"].(string)
	d.ExtSubnetID = dat["ExtSubnetID"].(string)
	d.inetGatewayID = dat["InetGatewayID"].(string)
	d.allocationID = dat["AllocationID"].(string)
	d.natGatewayID = dat["NatGatewayID"].(string)
	d.routeTableID = dat["RouteTableID"].(string)
	d.masterSecGrp = dat["MasterSecGrp"].(string)
	d.nodeSecGrp = dat["NodeSecGrp"].(string)
	d.edgeSecGrp = dat["EdgeSecGrp"].(string)
}

//-----------------------------------------------------------------------------
// func: retrieveEtcdToken
//-----------------------------------------------------------------------------

func (d *Data) retrieveEtcdToken(wg *sync.WaitGroup) {

	// Decrement:
	defer wg.Done()
	var err error

	if d.EtcdToken == "auto" {
		if d.EtcdToken, err = katool.EtcdToken(d.MasterCount); err != nil {
			log.WithField("cmd", "ec2:"+d.command).Error(err)
			os.Exit(1)
		}
		log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.EtcdToken}).
			Info("New etcd bootstrap token requested")
	}
}

//-----------------------------------------------------------------------------
// func: retrieveCoreosAmiID
//-----------------------------------------------------------------------------

func (d *Data) retrieveCoreosAmiID(wg *sync.WaitGroup) {

	// Decrement:
	defer wg.Done()

	// Send the request:
	res, err := http.
		Get("https://coreos.com/dist/aws/aws-" + d.Channel + ".json")
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		os.Exit(1)
	}

	// Retrieve the data:
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		os.Exit(1)
	}

	// Close the handler:
	if err = res.Body.Close(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		os.Exit(1)
	}

	// Decode JSON into Go values:
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		os.Exit(1)
	}

	// Store the AMI ID:
	amis := jsonData[d.Region].(map[string]interface{})
	d.ImageID = amis["hvm"].(string)

	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.ImageID}).
		Info("Latest CoreOS " + d.Channel + " AMI located")
}

//-----------------------------------------------------------------------------
// func: deployNodes
//-----------------------------------------------------------------------------

func (d *Data) deployNodes(role string, count int, wg *sync.WaitGroup) {

	// Decrement:
	defer wg.Done()
	var wgInt sync.WaitGroup

	log.WithField("cmd", "ec2:"+d.command).
		Info("Deploying " + strconv.Itoa(count) + " " + role + " nodes")

	for i := 1; i <= count; i++ {
		wgInt.Add(1)

		go func(id int) {
			defer wgInt.Done()

			// Forge the add command:
			cmdAdd := exec.Command("katoctl", "ec2", "add",
				"--cluster-id", d.ClusterID,
				"--role", role,
				"--id", strconv.Itoa(id))

			// Execute the add command:
			cmdAdd.Stderr = os.Stderr
			if err := cmdAdd.Run(); err != nil {
				log.WithField("cmd", "ec2:"+d.command).Error(err)
				os.Exit(1)
			}
		}(i)
	}

	// Wait:
	wgInt.Wait()
}

//-----------------------------------------------------------------------------
// func: forgeNetworkInterfaces
//-----------------------------------------------------------------------------

func (d *Data) forgeNetworkInterfaces() []*ec2.
	InstanceNetworkInterfaceSpecification {

	var networkInterfaces []*ec2.InstanceNetworkInterfaceSpecification
	var securityGroupIds []*string

	securityGroupIds = append(securityGroupIds, aws.String(d.SecGrpID))

	iface := ec2.InstanceNetworkInterfaceSpecification{
		DeleteOnTermination: aws.Bool(true),
		DeviceIndex:         aws.Int64(int64(0)),
		Groups:              securityGroupIds,
		SubnetId:            aws.String(d.SubnetID),
	}

	if d.PublicIP == "true" {
		iface.AssociatePublicIpAddress = aws.Bool(true)
	}

	networkInterfaces = append(networkInterfaces, &iface)

	return networkInterfaces
}

//-----------------------------------------------------------------------------
// func: runInstance
//-----------------------------------------------------------------------------

func (d *Data) runInstance(udata []byte) error {

	// Send the instance request:
	runResult, err := d.svc.ec2.RunInstances(&ec2.RunInstancesInput{
		ImageId:           aws.String(d.ImageID),
		MinCount:          aws.Int64(1),
		MaxCount:          aws.Int64(1),
		KeyName:           aws.String(d.KeyPair),
		InstanceType:      aws.String(d.InstanceType),
		NetworkInterfaces: d.forgeNetworkInterfaces(),
		Placement: &ec2.Placement{
			AvailabilityZone: aws.String(d.Zone),
		},
		UserData: aws.String(base64.StdEncoding.EncodeToString([]byte(udata))),
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Name: aws.String(d.IAMRole),
		},
	})

	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Store the instance ID:
	d.instanceID = *runResult.Instances[0].InstanceId
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.instanceID}).
		Info("New " + d.InstanceType + " EC2 instance requested")

	// Store the interface ID:
	d.interfaceID = *runResult.Instances[0].
		NetworkInterfaces[0].NetworkInterfaceId

	// Tag the instance:
	if err := d.tag(d.instanceID, "Name", d.Hostname); err != nil {
		return err
	}

	// Pretty-print to stderr:
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.Hostname}).
		Info("New EC2 instance tagged")

	return nil
}

//-----------------------------------------------------------------------------
// func: modifyInstanceAttribute
//-----------------------------------------------------------------------------

func (d *Data) modifyInstanceAttribute() error {

	// Variable transformation:
	SrcDstCheck, err := strconv.ParseBool(d.SrcDstCheck)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Forge the attribute modification request:
	params := &ec2.ModifyInstanceAttributeInput{
		InstanceId: aws.String(d.instanceID),
		SourceDestCheck: &ec2.AttributeBooleanValue{
			Value: aws.Bool(SrcDstCheck),
		},
	}

	// Send the attribute modification request:
	_, err = d.svc.ec2.ModifyInstanceAttribute(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: setupVPCNetwork
//-----------------------------------------------------------------------------

func (d *Data) setupVPCNetwork(wg *sync.WaitGroup) {

	// Decrement:
	defer wg.Done()

	// Retrieve the main route table ID:
	if err := d.retrieveMainRouteTableID(); err != nil {
		os.Exit(1)
	}

	// Create the external and internal subnets:
	if err := d.createSubnets(); err != nil {
		os.Exit(1)
	}

	// Create a route table (ext):
	if err := d.createRouteTable(); err != nil {
		os.Exit(1)
	}

	// Associate the route table to the external subnet:
	if err := d.associateRouteTable(); err != nil {
		os.Exit(1)
	}

	// Create the internet gateway:
	if err := d.createInternetGateway(); err != nil {
		os.Exit(1)
	}

	// Attach internet gateway to VPC:
	if err := d.attachInternetGateway(); err != nil {
		os.Exit(1)
	}

	// Create a default route via internet GW (ext):
	if err := d.createInternetGatewayRoute(); err != nil {
		os.Exit(1)
	}

	// Allocate a new elastic IP:
	if err := d.allocateElasticIP(); err != nil {
		os.Exit(1)
	}

	// Create a NAT gateway:
	if err := d.createNatGateway(); err != nil {
		os.Exit(1)
	}

	// Create a default route via NAT GW (int):
	if err := d.createNatGatewayRoute(); err != nil {
		os.Exit(1)
	}
}

//-----------------------------------------------------------------------------
// func: setupIAMSecurity
//-----------------------------------------------------------------------------

func (d *Data) setupIAMSecurity(wg *sync.WaitGroup) {

	// Decrement:
	defer wg.Done()

	// Create REX-Ray policy:
	if err := d.createRexrayPolicy(); err != nil {
		os.Exit(1)
	}

	// Create IAM roles:
	if err := d.createIAMRoles(); err != nil {
		os.Exit(1)
	}

	// Create instance profiles:
	d.createInstanceProfiles()

	// Attach REX-Ray policy to IAM role:
	if err := d.attachRexrayPolicy(); err != nil {
		os.Exit(1)
	}

	// Add IAM roles to instance profiles:
	if err := d.addIAMRolesToInstanceProfiles(); err != nil {
		os.Exit(1)
	}
}

//-----------------------------------------------------------------------------
// func: setupEC2Firewall
//-----------------------------------------------------------------------------

func (d *Data) setupEC2Firewall(wg *sync.WaitGroup) {

	// Decrement:
	defer wg.Done()

	// Create EC2 security groups:
	if err := d.createSecurityGroups(); err != nil {
		os.Exit(1)
	}

	// Setup master nodes firewall:
	if err := d.masterFirewall(); err != nil {
		os.Exit(1)
	}

	// Setup worker nodes firewall:
	if err := d.nodeFirewall(); err != nil {
		os.Exit(1)
	}

	// Setup edge nodes firewall:
	if err := d.edgeFirewall(); err != nil {
		os.Exit(1)
	}
}

//-----------------------------------------------------------------------------
// func: createVpc
//-----------------------------------------------------------------------------

func (d *Data) createVpc() error {

	// Forge the VPC request:
	params := &ec2.CreateVpcInput{
		CidrBlock:       aws.String(d.VpcCidrBlock),
		DryRun:          aws.Bool(false),
		InstanceTenancy: aws.String("default"),
	}

	// Send the VPC request:
	resp, err := d.svc.ec2.CreateVpc(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Store the VPC ID:
	d.vpcID = *resp.Vpc.VpcId
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.vpcID}).
		Info("New EC2 VPC created")

	// Tag the VPC:
	if err = d.tag(d.vpcID, "Name", d.Domain); err != nil {
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: retrieveMainRouteTableID
//-----------------------------------------------------------------------------

func (d *Data) retrieveMainRouteTableID() error {

	// Forge the description request:
	params := &ec2.DescribeRouteTablesInput{
		DryRun: aws.Bool(false),
		Filters: []*ec2.Filter{
			{
				Name: aws.String("association.main"),
				Values: []*string{
					aws.String("true"),
				},
			},
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(d.vpcID),
				},
			},
		},
	}

	// Send the description request:
	resp, err := d.svc.ec2.DescribeRouteTables(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Store the main route table ID:
	d.mainRouteTableID = *resp.RouteTables[0].RouteTableId
	log.WithFields(log.Fields{
		"cmd": "ec2:" + d.command, "id": d.mainRouteTableID}).
		Info("New main route table added")

	return nil
}

//-----------------------------------------------------------------------------
// func: createSubnets
//-----------------------------------------------------------------------------

func (d *Data) createSubnets() error {

	// Map to iterate:
	nets := map[string]map[string]string{
		"internal": map[string]string{
			"SubnetCidr": d.IntSubnetCidr, "SubnetID": ""},
		"external": map[string]string{
			"SubnetCidr": d.ExtSubnetCidr, "SubnetID": ""},
	}

	// For each subnet:
	for k, v := range nets {

		// Forge the subnet request:
		params := &ec2.CreateSubnetInput{
			CidrBlock:        aws.String(v["SubnetCidr"]),
			VpcId:            aws.String(d.vpcID),
			AvailabilityZone: aws.String(d.Zone),
			DryRun:           aws.Bool(false),
		}

		// Send the subnet request:
		resp, err := d.svc.ec2.CreateSubnet(params)
		if err != nil {
			log.WithField("cmd", "ec2:"+d.command).Error(err)
			return err
		}

		// Locally store the subnet ID:
		v["SubnetID"] = *resp.Subnet.SubnetId
		log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": v["SubnetID"]}).
			Info("New " + k + " subnet")

		// Tag the subnet:
		if err = d.tag(v["SubnetID"], "Name", k); err != nil {
			return err
		}
	}

	// Store subnet IDs:
	d.IntSubnetID = nets["internal"]["SubnetID"]
	d.ExtSubnetID = nets["external"]["SubnetID"]

	return nil
}

//-----------------------------------------------------------------------------
// func: createRouteTable
//-----------------------------------------------------------------------------

func (d *Data) createRouteTable() error {

	// Forge the route table request:
	params := &ec2.CreateRouteTableInput{
		VpcId:  aws.String(d.vpcID),
		DryRun: aws.Bool(false),
	}

	// Send the route table request:
	resp, err := d.svc.ec2.CreateRouteTable(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Store the route table ID:
	d.routeTableID = *resp.RouteTable.RouteTableId
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.routeTableID}).
		Info("New route table added")

	return nil
}

//-----------------------------------------------------------------------------
// func: associateRouteTable
//-----------------------------------------------------------------------------

func (d *Data) associateRouteTable() error {

	// Forge the association request:
	params := &ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(d.routeTableID),
		SubnetId:     aws.String(d.ExtSubnetID),
		DryRun:       aws.Bool(false),
	}

	// Send the association request:
	resp, err := d.svc.ec2.AssociateRouteTable(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	log.WithFields(log.Fields{
		"cmd": "ec2:" + d.command, "id": *resp.AssociationId}).
		Info("New route table association")

	return nil
}

//-----------------------------------------------------------------------------
// func: createInternetGateway
//-----------------------------------------------------------------------------

func (d *Data) createInternetGateway() error {

	// Forge the internet gateway request:
	params := &ec2.CreateInternetGatewayInput{
		DryRun: aws.Bool(false),
	}

	// Send the internet gateway request:
	resp, err := d.svc.ec2.CreateInternetGateway(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Store the internet gateway ID:
	d.inetGatewayID = *resp.InternetGateway.InternetGatewayId
	log.WithFields(log.Fields{
		"cmd": "ec2:" + d.command, "id": d.inetGatewayID}).
		Info("New internet gateway")

	return nil
}

//-----------------------------------------------------------------------------
// func: attachInternetGateway
//-----------------------------------------------------------------------------

func (d *Data) attachInternetGateway() error {

	// Forge the attachement request:
	params := &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(d.inetGatewayID),
		VpcId:             aws.String(d.vpcID),
		DryRun:            aws.Bool(false),
	}

	// Send the attachement request:
	if _, err := d.svc.ec2.AttachInternetGateway(params); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	log.WithField("cmd", "ec2:"+d.command).
		Info("Internet gateway attached to VPC")

	return nil
}

//-----------------------------------------------------------------------------
// func: createInternetGatewayRoute
//-----------------------------------------------------------------------------

func (d *Data) createInternetGatewayRoute() error {

	// Forge the route request:
	params := &ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		RouteTableId:         aws.String(d.routeTableID),
		DryRun:               aws.Bool(false),
		GatewayId:            aws.String(d.inetGatewayID),
	}

	// Send the route request:
	if _, err := d.svc.ec2.CreateRoute(params); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	log.WithField("cmd", "ec2:"+d.command).
		Info("New default route added via internet GW")

	return nil
}

//-----------------------------------------------------------------------------
// func: allocateElasticIP
//-----------------------------------------------------------------------------

func (d *Data) allocateElasticIP() error {

	// Forge the allocation request:
	params := &ec2.AllocateAddressInput{
		Domain: aws.String("vpc"),
		DryRun: aws.Bool(false),
	}

	// Send the allocation request:
	resp, err := d.svc.ec2.AllocateAddress(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Store the EIP ID:
	d.allocationID = *resp.AllocationId
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.allocationID}).
		Info("New elastic IP allocated")

	return nil
}

//-----------------------------------------------------------------------------
// func: associateElasticIP
//-----------------------------------------------------------------------------

func (d *Data) associateElasticIP() error {

	// Wait until instance is running:
	if err := d.svc.ec2.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(d.instanceID)},
	}); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Forge the association request:
	params := &ec2.AssociateAddressInput{
		AllocationId:       aws.String(d.allocationID),
		AllowReassociation: aws.Bool(true),
		DryRun:             aws.Bool(false),
		NetworkInterfaceId: aws.String(d.interfaceID),
	}

	// Send the association request:
	resp, err := d.svc.ec2.AssociateAddress(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	log.WithFields(log.Fields{
		"cmd": "ec2:" + d.command, "id": *resp.AssociationId}).
		Info("New elastic IP association")

	return nil
}

//-----------------------------------------------------------------------------
// func: createNatGateway
//-----------------------------------------------------------------------------

func (d *Data) createNatGateway() error {

	// Forge the NAT gateway request:
	params := &ec2.CreateNatGatewayInput{
		AllocationId: aws.String(d.allocationID),
		SubnetId:     aws.String(d.ExtSubnetID),
		ClientToken:  aws.String(d.Domain),
	}

	// Send the NAT gateway request:
	resp, err := d.svc.ec2.CreateNatGateway(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Store the NAT gateway ID:
	d.natGatewayID = *resp.NatGateway.NatGatewayId
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.natGatewayID}).
		Info("New NAT gateway requested")

	// Wait until the NAT gateway is available:
	log.WithField("cmd", "ec2:"+d.command).
		Info("Waiting until NAT gateway is available")
	if err := d.svc.ec2.WaitUntilNatGatewayAvailable(&ec2.DescribeNatGatewaysInput{
		NatGatewayIds: []*string{aws.String(d.natGatewayID)},
	}); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: createNatGatewayRoute
//-----------------------------------------------------------------------------

func (d *Data) createNatGatewayRoute() error {

	// Forge the route request:
	params := &ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		RouteTableId:         aws.String(d.mainRouteTableID),
		DryRun:               aws.Bool(false),
		NatGatewayId:         aws.String(d.natGatewayID),
	}

	// Send the route request:
	if _, err := d.svc.ec2.CreateRoute(params); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	log.WithField("cmd", "ec2:"+d.command).
		Info("New default route added via NAT gateway")

	return nil
}

//-----------------------------------------------------------------------------
// func: createRexrayPolicy
//-----------------------------------------------------------------------------

func (d *Data) createRexrayPolicy() error {

	// Forge the listing request:
	listPrms := &iam.ListPoliciesInput{
		MaxItems:     aws.Int64(100),
		OnlyAttached: aws.Bool(false),
		PathPrefix:   aws.String("/kato/"),
		Scope:        aws.String("Local"),
	}

	// Send the listing request:
	listRsp, err := d.svc.iam.ListPolicies(listPrms)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Check whether the policy exists:
	for _, v := range listRsp.Policies {
		if *v.PolicyName == "REX-Ray" {
			d.rexrayPolicyARN = *listRsp.Policies[0].Arn
			log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": *listRsp.
				Policies[0].PolicyId}).Info("Using existing REX-Ray security policy")
			return nil
		}
	}

	// REX-Ray IAM policy:
	policy := `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "RexRayMin",
            "Effect": "Allow",
            "Action": [
                "ec2:AttachVolume",
                "ec2:CreateVolume",
                "ec2:CreateSnapshot",
                "ec2:CreateTags",
                "ec2:DeleteVolume",
                "ec2:DeleteSnapshot",
                "ec2:DescribeAvailabilityZones",
                "ec2:DescribeInstances",
                "ec2:DescribeVolumes",
                "ec2:DescribeVolumeAttribute",
                "ec2:DescribeVolumeStatus",
                "ec2:DescribeSnapshots",
                "ec2:CopySnapshot",
                "ec2:DescribeSnapshotAttribute",
                "ec2:DetachVolume",
                "ec2:ModifySnapshotAttribute",
                "ec2:ModifyVolumeAttribute",
                "ec2:DescribeTags"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
	}`

	// Forge the policy request:
	policyPrms := &iam.CreatePolicyInput{
		PolicyDocument: aws.String(policy),
		PolicyName:     aws.String("REX-Ray"),
		Description:    aws.String("Enables necessary functionality for REX-Ray"),
		Path:           aws.String("/kato/"),
	}

	// Send the policy request:
	policyRsp, err := d.svc.iam.CreatePolicy(policyPrms)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Store the policy ARN:
	d.rexrayPolicyARN = *policyRsp.Policy.Arn
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": *policyRsp.Policy.
		PolicyId}).Info("Setup REX-Ray security policy")

	return nil
}

//-----------------------------------------------------------------------------
// func: atachRexrayPolicy
//-----------------------------------------------------------------------------

func (d *Data) attachRexrayPolicy() error {

	// Forge the attachment request:
	params := &iam.AttachRolePolicyInput{
		PolicyArn: aws.String(d.rexrayPolicyARN),
		RoleName:  aws.String("node"),
	}

	// Send the attachement request:
	_, err := d.svc.iam.AttachRolePolicy(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	log.WithField("cmd", "ec2:"+d.command).
		Info("REX-Ray policy attached to node")

	return nil
}

//-----------------------------------------------------------------------------
// func: createIAMRoles
//-----------------------------------------------------------------------------

func (d *Data) createIAMRoles() error {

	// Map to iterate:
	grps := map[string]map[string]string{
		"master": map[string]string{"roleID": ""},
		"node":   map[string]string{"roleID": ""},
		"edge":   map[string]string{"roleID": ""},
	}

	// IAM Role type:
	policy := `{
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": ["ec2.amazonaws.com"]
            },
            "Action": ["sts:AssumeRole"]
        }]
  }`

	// For each security role:
	for k, v := range grps {

		// Forge the role request:
		params := &iam.CreateRoleInput{
			AssumeRolePolicyDocument: aws.String(policy),
			RoleName:                 aws.String(k),
			Path:                     aws.String("/kato/"),
		}

		// Send the role request:
		resp, err := d.svc.iam.CreateRole(params)
		if err != nil {
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				if reqErr.StatusCode() == 409 {
					continue
				}
			}
			log.WithField("cmd", "ec2:"+d.command).Error(err)
			return err
		}

		// Locally store the role ID:
		v["roleID"] = *resp.Role.RoleId
		log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": v["roleID"]}).
			Info("New " + k + " IAM role")
	}

	// Store security role IDs:
	if grps["master"]["roleID"] != "" {
		d.masterRoleID = grps["master"]["roleID"]
	}
	if grps["node"]["roleID"] != "" {
		d.nodeRoleID = grps["node"]["roleID"]
	}
	if grps["edge"]["roleID"] != "" {
		d.edgeRoleID = grps["edge"]["roleID"]
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: createInstanceProfiles
//-----------------------------------------------------------------------------

func (d *Data) createInstanceProfiles() {

	// Setup a wait group:
	var wg sync.WaitGroup

	// For each instance profile:
	for _, v := range [3]string{"master", "node", "edge"} {

		// Increment wait group:
		wg.Add(1)

		go func(role string) {

			// Decrement wait group:
			defer wg.Done()

			// Forge the profile request:
			params := &iam.CreateInstanceProfileInput{
				InstanceProfileName: aws.String(role),
				Path:                aws.String("/kato/"),
			}

			// Send the profile request:
			resp, err := d.svc.iam.CreateInstanceProfile(params)
			if err != nil {
				if reqErr, ok := err.(awserr.RequestFailure); ok {
					if reqErr.StatusCode() == 409 {
						return
					}
				}
				log.WithField("cmd", "ec2:"+d.command).Error(err)
				os.Exit(1)
			}

			// Wait until the instance profile exists:
			log.WithFields(log.Fields{"cmd": "ec2:" + d.command,
				"id": *resp.InstanceProfile.InstanceProfileId}).
				Info("Waiting until " + role + " profile exists")
			if err := d.svc.iam.WaitUntilInstanceProfileExists(
				&iam.GetInstanceProfileInput{
					InstanceProfileName: aws.String(role),
				}); err != nil {
				log.WithField("cmd", "ec2:"+d.command).Error(err)
				os.Exit(1)
			}
		}(v)
	}

	// Wait:
	wg.Wait()
}

//-----------------------------------------------------------------------------
// func: addIAMRolesToInstanceProfiles
//-----------------------------------------------------------------------------

func (d *Data) addIAMRolesToInstanceProfiles() error {

	// For each instance profile:
	for _, v := range [3]string{"master", "node", "edge"} {

		// Forge the addition request:
		params := &iam.AddRoleToInstanceProfileInput{
			InstanceProfileName: aws.String(v),
			RoleName:            aws.String(v),
		}

		// Send the addition request:
		if _, err := d.svc.iam.AddRoleToInstanceProfile(params); err != nil {
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				if reqErr.StatusCode() == 409 {
					continue
				}
			}
			log.WithField("cmd", "ec2:"+d.command).Error(err)
			return err
		}

		// Log the addition request:
		log.WithField("cmd", "ec2:"+d.command).
			Info("New " + v + " IAM role added to profile")
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: createSecurityGroups
//-----------------------------------------------------------------------------

func (d *Data) createSecurityGroups() error {

	// Map to iterate:
	grps := map[string]map[string]string{
		"master": map[string]string{"secGrpID": ""},
		"node":   map[string]string{"secGrpID": ""},
		"edge":   map[string]string{"secGrpID": ""},
	}

	// For each security group:
	for k, v := range grps {

		// Forge the group request:
		params := &ec2.CreateSecurityGroupInput{
			Description: aws.String(d.Domain + " " + k),
			GroupName:   aws.String(k),
			DryRun:      aws.Bool(false),
			VpcId:       aws.String(d.vpcID),
		}

		// Send the group request:
		resp, err := d.svc.ec2.CreateSecurityGroup(params)
		if err != nil {
			log.WithField("cmd", "ec2:"+d.command).Error(err)
			return err
		}

		// Locally store the group ID:
		v["secGrpID"] = *resp.GroupId
		log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": v["secGrpID"]}).
			Info("New EC2 " + k + " security group")

		// Tag the group:
		if err = d.tag(v["secGrpID"], "Name", d.Domain+" "+k); err != nil {
			return err
		}
	}

	// Store security groups IDs:
	d.masterSecGrp = grps["master"]["secGrpID"]
	d.nodeSecGrp = grps["node"]["secGrpID"]
	d.edgeSecGrp = grps["edge"]["secGrpID"]

	return nil
}

//-----------------------------------------------------------------------------
// func: masterFirewall
//-----------------------------------------------------------------------------

func (d *Data) masterFirewall() error {

	// Forge the rule request:
	params := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(d.masterSecGrp),
		IpPermissions: []*ec2.IpPermission{
			{
				IpProtocol: aws.String("-1"),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						GroupId: aws.String(d.masterSecGrp),
					},
					{
						GroupId: aws.String(d.nodeSecGrp),
					},
					{
						GroupId: aws.String(d.edgeSecGrp),
					},
				},
			},
		},
	}

	// Send the rule request:
	_, err := d.svc.ec2.AuthorizeSecurityGroupIngress(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": "master"}).
		Info("New firewall rules defined")

	return nil
}

//-----------------------------------------------------------------------------
// func: nodeFirewall
//-----------------------------------------------------------------------------

func (d *Data) nodeFirewall() error {

	// Forge the rule request:
	params := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(d.nodeSecGrp),
		IpPermissions: []*ec2.IpPermission{
			{
				IpProtocol: aws.String("-1"),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						GroupId: aws.String(d.masterSecGrp),
					},
					{
						GroupId: aws.String(d.nodeSecGrp),
					},
					{
						GroupId: aws.String(d.edgeSecGrp),
					},
				},
			},
			{
				FromPort:   aws.Int64(80),
				ToPort:     aws.Int64(80),
				IpProtocol: aws.String("tcp"),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp: aws.String("0.0.0.0/0"),
					},
				},
			},
			{
				FromPort:   aws.Int64(443),
				ToPort:     aws.Int64(443),
				IpProtocol: aws.String("tcp"),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp: aws.String("0.0.0.0/0"),
					},
				},
			},
		},
	}

	// Send the rule request:
	_, err := d.svc.ec2.AuthorizeSecurityGroupIngress(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": "node"}).
		Info("New firewall rules defined")

	return nil
}

//-----------------------------------------------------------------------------
// func: edgeFirewall
//-----------------------------------------------------------------------------

func (d *Data) edgeFirewall() error {

	// Forge the rule request:
	params := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(d.edgeSecGrp),
		IpPermissions: []*ec2.IpPermission{
			{
				IpProtocol: aws.String("-1"),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						GroupId: aws.String(d.masterSecGrp),
					},
					{
						GroupId: aws.String(d.nodeSecGrp),
					},
					{
						GroupId: aws.String(d.edgeSecGrp),
					},
				},
			},
			{
				FromPort:   aws.Int64(22),
				ToPort:     aws.Int64(22),
				IpProtocol: aws.String("tcp"),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp: aws.String("0.0.0.0/0"),
					},
				},
			},
			{
				FromPort:   aws.Int64(80),
				ToPort:     aws.Int64(80),
				IpProtocol: aws.String("tcp"),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp: aws.String("0.0.0.0/0"),
					},
				},
			},
			{
				FromPort:   aws.Int64(443),
				ToPort:     aws.Int64(443),
				IpProtocol: aws.String("tcp"),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp: aws.String("0.0.0.0/0"),
					},
				},
			},
			{
				FromPort:   aws.Int64(18443),
				ToPort:     aws.Int64(18443),
				IpProtocol: aws.String("udp"),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp: aws.String("0.0.0.0/0"),
					},
				},
			},
		},
	}

	// Send the rule request:
	_, err := d.svc.ec2.AuthorizeSecurityGroupIngress(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": "edge"}).
		Info("New firewall rules defined")

	return nil
}

//-----------------------------------------------------------------------------
// func: dumpState
//-----------------------------------------------------------------------------

func (d *Data) dumpState() error {

	type identifiers struct {
		MasterCount      int
		VpcCidrBlock     string
		VpcID            string
		MainRouteTableID string
		IntSubnetCidr    string
		ExtSubnetCidr    string
		IntSubnetID      string
		ExtSubnetID      string
		InetGatewayID    string
		AllocationID     string
		NatGatewayID     string
		RouteTableID     string
		MasterSecGrp     string
		NodeSecGrp       string
		EdgeSecGrp       string
		Domain           string
		Ns1ApiKey        string
		CaCert           string
		EtcdToken        string
		FlannelNetwork   string
		FlannelSubnetLen string
		FlannelSubnetMin string
		FlannelSubnetMax string
		FlannelBackend   string
		Region           string
		Zone             string
		ImageID          string
		MasterType       string
		NodeType         string
		EdgeType         string
		KeyPair          string
		SrcDstCheck      string
	}

	ids := identifiers{
		MasterCount:      d.MasterCount,
		VpcCidrBlock:     d.VpcCidrBlock,
		VpcID:            d.vpcID,
		MainRouteTableID: d.mainRouteTableID,
		IntSubnetCidr:    d.IntSubnetCidr,
		ExtSubnetCidr:    d.ExtSubnetCidr,
		IntSubnetID:      d.IntSubnetID,
		ExtSubnetID:      d.ExtSubnetID,
		InetGatewayID:    d.inetGatewayID,
		AllocationID:     d.allocationID,
		NatGatewayID:     d.natGatewayID,
		RouteTableID:     d.routeTableID,
		MasterSecGrp:     d.masterSecGrp,
		NodeSecGrp:       d.nodeSecGrp,
		EdgeSecGrp:       d.edgeSecGrp,
		Domain:           d.Domain,
		Ns1ApiKey:        d.Ns1ApiKey,
		CaCert:           d.CaCert,
		EtcdToken:        d.EtcdToken,
		FlannelNetwork:   d.FlannelNetwork,
		FlannelSubnetLen: d.FlannelSubnetLen,
		FlannelSubnetMin: d.FlannelSubnetMin,
		FlannelSubnetMax: d.FlannelSubnetMax,
		FlannelBackend:   d.FlannelBackend,
		Region:           d.Region,
		Zone:             d.Zone,
		ImageID:          d.ImageID,
		MasterType:       d.MasterType,
		NodeType:         d.NodeType,
		EdgeType:         d.EdgeType,
		KeyPair:          d.KeyPair,
		SrcDstCheck:      d.SrcDstCheck,
	}

	// Marshal the data:
	idsJSON, err := json.MarshalIndent(ids, "", "  ")
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Create the state directory:
	path := os.Getenv("HOME") + "/.kato"
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(path, 0700)
			if err != nil {
				log.WithField("cmd", "ec2:"+d.command).Error(err)
				return err
			}
		}
	}

	// Write the state file:
	err = ioutil.WriteFile(path+"/"+d.ClusterID+".json", idsJSON, 0600)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: tag
//-----------------------------------------------------------------------------

func (d *Data) tag(resource, key, value string) error {

	// Forge the tag request:
	params := &ec2.CreateTagsInput{
		Resources: []*string{
			aws.String(resource),
		},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String(key),
				Value: aws.String(value),
			},
		},
		DryRun: aws.Bool(false),
	}

	// Send the tag request:
	if _, err := d.svc.ec2.CreateTags(params); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	return nil
}
