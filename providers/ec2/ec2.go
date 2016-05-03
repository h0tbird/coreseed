package ec2

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"encoding/base64"
	"encoding/json"
	"fmt"
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

// Data contains variables used by this EC2 provider.
type Data struct {

	// AWS API endpoints:
	svcEC2 *ec2.EC2
	svcIAM *iam.IAM

	MasterCount       int    //  deploy:ec2 |           |       |
	NodeCount         int    //  deploy:ec2 |           |       |
	EdgeCount         int    //  deploy:ec2 |           |       |
	MasterType        string //  deploy:ec2 |           |       |
	NodeType          string //  deploy:ec2 |           |       |
	EdgeType          string //  deploy:ec2 |           |       |
	Channel           string //  deploy:ec2 |           |       |
	EtcdToken         string //  deploy:ec2 |           | udata |
	Ns1ApiKey         string //  deploy:ec2 |           | udata |
	CaCert            string //  deploy:ec2 |           | udata |
	FlannelNetwork    string //  deploy:ec2 |           | udata |
	FlannelSubnetLen  string //  deploy:ec2 |           | udata |
	FlannelSubnetMin  string //  deploy:ec2 |           | udata |
	FlannelSubnetMax  string //  deploy:ec2 |           | udata |
	FlannelBackend    string //  deploy:ec2 |           | udata |
	Domain            string //  deploy:ec2 | setup:ec2 | udata |
	Region            string //  deploy:ec2 | setup:ec2 |       | run:ec2
	command           string //  deploy:ec2 | setup:ec2 |       | run:ec2
	VpcCidrBlock      string //  deploy:ec2 | setup:ec2 |       |
	IntSubnetCidr     string //  deploy:ec2 | setup:ec2 |       |
	ExtSubnetCidr     string //  deploy:ec2 | setup:ec2 |       |
	vpcID             string //             | setup:ec2 |       |
	mainRouteTableID  string //             | setup:ec2 |       |
	internetGatewayID string //             | setup:ec2 |       |
	natGatewayID      string //             | setup:ec2 |       |
	routeTableID      string //             | setup:ec2 |       |
	masterRoleID      string //             | setup:ec2 |       |
	nodeRoleID        string //             | setup:ec2 |       |
	edgeRoleID        string //             | setup:ec2 |       |
	rexrayPolicyARN   string //             | setup:ec2 |       |
	masterSecGrp      string //             | setup:ec2 |       |
	nodeSecGrp        string //             | setup:ec2 |       |
	edgeSecGrp        string //             | setup:ec2 |       |
	IntSubnetID       string //             | setup:ec2 |       |
	ExtSubnetID       string //             | setup:ec2 |       |
	allocationID      string //             | setup:ec2 |       | run:ec2
	instanceID        string //             |           |       | run:ec2
	SubnetID          string //             |           |       | run:ec2
	SecGrpID          string //             |           |       | run:ec2
	ImageID           string //             |           |       | run:ec2
	KeyPair           string //             |           |       | run:ec2
	InstanceType      string //             |           |       | run:ec2
	Hostname          string //             |           |       | run:ec2
	PublicIP          string //             |           |       | run:ec2
	interfaceID       string //             |           |       | run:ec2
}

//-----------------------------------------------------------------------------
// func: Deploy
//-----------------------------------------------------------------------------

// Deploy Kato's infrastructure on Amazon EC2.
func (d *Data) Deploy() error {

	// Set command to deploy:
	d.command = "deploy"

	// Setup the EC2 environment:
	if err := d.environmentSetup(); err != nil {
		return err
	}

	// Retrieve the etcd bootstrap token:
	if err := d.retrieveEtcdToken(); err != nil {
		return err
	}

	// Retrieve the CoreOS AMI ID:
	if err := d.retrieveCoreosAmiID(); err != nil {
		return err
	}

	// Setup a wait group:
	var wg sync.WaitGroup
	wg.Add(3)

	// Deploy all the nodes:
	go d.deployMasterNodes(&wg)
	go d.deployWorkerNodes(&wg)
	go d.deployEdgeNodes(&wg)

	// Wait and return:
	wg.Wait()
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
	d.svcEC2 = ec2.New(session.New(&aws.Config{Region: aws.String(d.Region)}))

	// Run the EC2 instance:
	if err := d.runInstance(udata); err != nil {
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

// Setup an EC2 VPC and all the related components.
func (d *Data) Setup() error {

	// Set command to setup:
	d.command = "setup"

	// Connect and authenticate to the API endpoints:
	log.WithField("cmd", d.command+":ec2").
		Info("- Connecting to region " + d.Region)

	d.svcEC2 = ec2.New(session.New(&aws.Config{Region: aws.String(d.Region)}))
	d.svcIAM = iam.New(session.New())

	// Setup the network:
	if err := d.setupNetwork(); err != nil {
		return err
	}

	// Setup security items:
	if err := d.setupSecurity(); err != nil {
		return err
	}

	// Expose identifiers to stdout:
	if err := d.exposeIdentifiers(); err != nil {
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: environmentSetup
//-----------------------------------------------------------------------------

func (d *Data) environmentSetup() error {

	// Forge the setup command:
	log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": d.Domain}).
		Info("Setup the EC2 environment")

	cmdSetup := exec.Command("katoctl", "setup", "ec2",
		"--domain", d.Domain,
		"--region", d.Region,
		"--vpc-cidr-block", d.VpcCidrBlock,
		"--internal-subnet-cidr", d.IntSubnetCidr,
		"--external-subnet-cidr", d.ExtSubnetCidr)

	// Execute the setup command:
	cmdSetup.Stderr = os.Stderr
	out, err := cmdSetup.Output()
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Decode JSON data returned from katoctl-setup:
	var dat map[string]interface{}
	if err := json.Unmarshal(out, &dat); err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Store the values:
	d.VpcCidrBlock = dat["VpcCidrBlock"].(string)
	d.IntSubnetCidr = dat["IntSubnetCidr"].(string)
	d.ExtSubnetCidr = dat["ExtSubnetCidr"].(string)
	d.vpcID = dat["VpcID"].(string)
	d.mainRouteTableID = dat["MainRouteTableID"].(string)
	d.IntSubnetID = dat["InternalSubnetID"].(string)
	d.ExtSubnetID = dat["ExternalSubnetID"].(string)
	d.internetGatewayID = dat["InternetGatewayID"].(string)
	d.allocationID = dat["AllocationID"].(string)
	d.natGatewayID = dat["NatGatewayID"].(string)
	d.routeTableID = dat["RouteTableID"].(string)
	d.masterSecGrp = dat["MasterSecGrp"].(string)
	d.nodeSecGrp = dat["NodeSecGrp"].(string)
	d.edgeSecGrp = dat["EdgeSecGrp"].(string)

	return nil
}

//-----------------------------------------------------------------------------
// func: retrieveEtcdToken
//-----------------------------------------------------------------------------

func (d *Data) retrieveEtcdToken() error {

	var err error

	if d.EtcdToken == "auto" {
		if d.EtcdToken, err = katool.EtcdToken(d.MasterCount); err != nil {
			log.WithField("cmd", d.command+":ec2").Error(err)
			return err
		}
		log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": d.EtcdToken}).
			Info("New etcd bootstrap token requested")
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: retrieveCoreosAmiID
//-----------------------------------------------------------------------------

func (d *Data) retrieveCoreosAmiID() error {

	// Send the request:
	res, err := http.
		Get("https://coreos.com/dist/aws/aws-" + d.Channel + ".json")
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Retrieve the data:
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Close the handler:
	if err = res.Body.Close(); err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Decode JSON into Go values:
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Store the AMI ID:
	amis := jsonData[d.Region].(map[string]interface{})
	d.ImageID = amis["hvm"].(string)

	log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": d.ImageID}).
		Info("Latest CoreOS " + d.Channel + " AMI located")

	return nil
}

//-----------------------------------------------------------------------------
// func: deployMasterNodes
//-----------------------------------------------------------------------------

func (d *Data) deployMasterNodes(wg *sync.WaitGroup) {

	// Decrement:
	defer wg.Done()
	var wgInt sync.WaitGroup

	log.WithField("cmd", d.command+":ec2").
		Info("Deploying " + strconv.Itoa(d.MasterCount) + " master nodes")

	for i := 1; i <= d.MasterCount; i++ {

		// Increment:
		wgInt.Add(1)

		go func(id int) {

			// Decrement:
			defer wgInt.Done()

			// Forge the udata command:
			cmdUdata := exec.Command("katoctl", "udata",
				"--role", "master",
				"--master-count", strconv.Itoa(d.MasterCount),
				"--hostid", strconv.Itoa(id),
				"--domain", d.Domain,
				"--ns1-api-key", d.Ns1ApiKey,
				"--ca-cert", d.CaCert,
				"--etcd-token", d.EtcdToken,
				"--gzip-udata")

			// Forge the run command:
			cmdRun := exec.Command("katoctl", "run", "ec2",
				"--hostname", "master-"+strconv.Itoa(id)+"."+d.Domain,
				"--region", d.Region,
				"--image-id", d.ImageID,
				"--instance-type", d.MasterType,
				"--key-pair", d.KeyPair,
				"--subnet-id", d.IntSubnetID,
				"--security-group-id", d.masterSecGrp)

			// Execute the pipeline:
			if err := katool.ExecutePipeline(cmdUdata, cmdRun); err != nil {
				log.WithField("cmd", d.command+":ec2").Error(err)
			}
		}(i)
	}

	// Wait:
	wgInt.Wait()
}

//-----------------------------------------------------------------------------
// func: deployWorkerNodes
//-----------------------------------------------------------------------------

func (d *Data) deployWorkerNodes(wg *sync.WaitGroup) {

	// Decrement:
	defer wg.Done()
	var wgInt sync.WaitGroup

	log.WithField("cmd", d.command+":ec2").
		Info("Deploying " + strconv.Itoa(d.NodeCount) + " worker nodes")

	for i := 1; i <= d.NodeCount; i++ {

		// Increment:
		wgInt.Add(1)

		go func(id int) {

			// Decrement:
			defer wgInt.Done()

			// Forge the udata command:
			cmdUdata := exec.Command("katoctl", "udata",
				"--role", "node",
				"--master-count", strconv.Itoa(d.MasterCount),
				"--hostid", strconv.Itoa(id),
				"--domain", d.Domain,
				"--ns1-api-key", d.Ns1ApiKey,
				"--ca-cert", d.CaCert,
				"--etcd-token", d.EtcdToken,
				"--gzip-udata",
				"--flannel-network", d.FlannelNetwork,
				"--flannel-subnet-len", d.FlannelSubnetLen,
				"--flannel-subnet-min", d.FlannelSubnetMin,
				"--flannel-subnet-max", d.FlannelSubnetMax,
				"--flannel-backend", d.FlannelBackend,
				"--rexray-storage-driver", "ec2")

			// Forge the run command:
			cmdRun := exec.Command("katoctl", "run", "ec2",
				"--hostname", "node-"+strconv.Itoa(id)+"."+d.Domain,
				"--region", d.Region,
				"--image-id", d.ImageID,
				"--instance-type", d.NodeType,
				"--key-pair", d.KeyPair,
				"--subnet-id", d.ExtSubnetID,
				"--security-group-id", d.nodeSecGrp,
				"--public-ip", "true")

			// Execute the pipeline:
			if err := katool.ExecutePipeline(cmdUdata, cmdRun); err != nil {
				log.WithField("cmd", d.command+":ec2").Error(err)
			}
		}(i)
	}

	// Wait:
	wgInt.Wait()
}

//-----------------------------------------------------------------------------
// func: deployEdgeNodes
//-----------------------------------------------------------------------------

func (d *Data) deployEdgeNodes(wg *sync.WaitGroup) {

	// Decrement:
	defer wg.Done()
	var wgInt sync.WaitGroup

	log.WithField("cmd", d.command+":ec2").
		Info("Deploying " + strconv.Itoa(d.EdgeCount) + " edge nodes")

	for i := 1; i <= d.EdgeCount; i++ {

		// Increment:
		wgInt.Add(1)

		go func(id int) {

			// Decrement:
			defer wgInt.Done()

			// Forge the udata command:
			cmdUdata := exec.Command("katoctl", "udata",
				"--role", "edge",
				"--master-count", strconv.Itoa(d.MasterCount),
				"--hostid", strconv.Itoa(id),
				"--domain", d.Domain,
				"--ns1-api-key", d.Ns1ApiKey,
				"--ca-cert", d.CaCert,
				"--etcd-token", d.EtcdToken,
				"--gzip-udata")

			// Forge the run command:
			cmdRun := exec.Command("katoctl", "run", "ec2",
				"--hostname", "edge-"+strconv.Itoa(id)+"."+d.Domain,
				"--region", d.Region,
				"--image-id", d.ImageID,
				"--instance-type", d.EdgeType,
				"--key-pair", d.KeyPair,
				"--subnet-id", d.ExtSubnetID,
				"--security-group-id", d.edgeSecGrp,
				"--public-ip", "true")

			// Execute the pipeline:
			if err := katool.ExecutePipeline(cmdUdata, cmdRun); err != nil {
				log.WithField("cmd", d.command+":ec2").Error(err)
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
	runResult, err := d.svcEC2.RunInstances(&ec2.RunInstancesInput{
		ImageId:           aws.String(d.ImageID),
		MinCount:          aws.Int64(1),
		MaxCount:          aws.Int64(1),
		KeyName:           aws.String(d.KeyPair),
		InstanceType:      aws.String(d.InstanceType),
		NetworkInterfaces: d.forgeNetworkInterfaces(),
		UserData:          aws.String(base64.StdEncoding.EncodeToString([]byte(udata))),
	})

	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Store the instance ID:
	d.instanceID = *runResult.Instances[0].InstanceId
	log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": d.instanceID}).
		Info("- New " + d.InstanceType + " EC2 instance requested")

	// Store the interface ID:
	d.interfaceID = *runResult.Instances[0].
		NetworkInterfaces[0].NetworkInterfaceId

	// Tag the instance:
	if err := d.tag(d.instanceID, "Name", d.Hostname); err != nil {
		return err
	}

	// Pretty-print to stderr:
	log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": d.Hostname}).
		Info("- New EC2 instance tagged")

	return nil
}

//-----------------------------------------------------------------------------
// func: setupNetwork
//-----------------------------------------------------------------------------

func (d *Data) setupNetwork() error {

	// Create the VPC:
	if err := d.createVpc(); err != nil {
		return err
	}

	// Retrieve the main route table ID:
	if err := d.retrieveMainRouteTableID(); err != nil {
		return err
	}

	// Create the external and internal subnets:
	if err := d.createSubnets(); err != nil {
		return err
	}

	// Create a route table (ext):
	if err := d.createRouteTable(); err != nil {
		return err
	}

	// Associate the route table to the external subnet:
	if err := d.associateRouteTable(); err != nil {
		return err
	}

	// Create the internet gateway:
	if err := d.createInternetGateway(); err != nil {
		return err
	}

	// Attach internet gateway to VPC:
	if err := d.attachInternetGateway(); err != nil {
		return err
	}

	// Create a default route via internet GW (ext):
	if err := d.createInternetGatewayRoute(); err != nil {
		return err
	}

	// Allocate a new elastic IP:
	if err := d.allocateElasticIP(); err != nil {
		return err
	}

	// Create a NAT gateway:
	if err := d.createNatGateway(); err != nil {
		return err
	}

	// Create a default route via NAT GW (int):
	if err := d.createNatGatewayRoute(); err != nil {
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: setupSecurity
//-----------------------------------------------------------------------------

func (d *Data) setupSecurity() error {

	// Create security roles:
	if err := d.createSecurityRoles(); err != nil {
		return err
	}

	// Create REX-Ray policy:
	if err := d.createRexrayPolicy(); err != nil {
		return err
	}

	// Attach REX-Ray policy:
	if err := d.attachRexrayPolicy(); err != nil {
		return err
	}

	// Create security groups:
	if err := d.createSecurityGroups(); err != nil {
		return err
	}

	// Setup master nodes firewall:
	if err := d.masterFirewall(); err != nil {
		return err
	}

	// Setup worker nodes firewall:
	if err := d.nodeFirewall(); err != nil {
		return err
	}

	// Setup edge nodes firewall:
	if err := d.edgeFirewall(); err != nil {
		return err
	}

	return nil
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
	resp, err := d.svcEC2.CreateVpc(params)
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Store the VPC ID:
	d.vpcID = *resp.Vpc.VpcId
	log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": d.vpcID}).
		Info("- New EC2 VPC created")

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
	resp, err := d.svcEC2.DescribeRouteTables(params)
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Store the main route table ID:
	d.mainRouteTableID = *resp.RouteTables[0].RouteTableId
	log.WithFields(log.Fields{
		"cmd": d.command + ":ec2", "id": d.mainRouteTableID}).
		Info("- New main route table added")

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
			CidrBlock: aws.String(v["SubnetCidr"]),
			VpcId:     aws.String(d.vpcID),
			DryRun:    aws.Bool(false),
		}

		// Send the subnet request:
		resp, err := d.svcEC2.CreateSubnet(params)
		if err != nil {
			log.WithField("cmd", d.command+":ec2").Error(err)
			return err
		}

		// Locally store the subnet ID:
		v["SubnetID"] = *resp.Subnet.SubnetId
		log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": v["SubnetID"]}).
			Info("- New " + k + " subnet")

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
	resp, err := d.svcEC2.CreateRouteTable(params)
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Store the route table ID:
	d.routeTableID = *resp.RouteTable.RouteTableId
	log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": d.routeTableID}).
		Info("- New route table added")

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
	resp, err := d.svcEC2.AssociateRouteTable(params)
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	log.WithFields(log.Fields{
		"cmd": d.command + ":ec2", "id": *resp.AssociationId}).
		Info("- New route table association")

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
	resp, err := d.svcEC2.CreateInternetGateway(params)
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Store the internet gateway ID:
	d.internetGatewayID = *resp.InternetGateway.InternetGatewayId
	log.WithFields(log.Fields{
		"cmd": d.command + ":ec2", "id": d.internetGatewayID}).
		Info("- New internet gateway")

	return nil
}

//-----------------------------------------------------------------------------
// func: attachInternetGateway
//-----------------------------------------------------------------------------

func (d *Data) attachInternetGateway() error {

	// Forge the attachement request:
	params := &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(d.internetGatewayID),
		VpcId:             aws.String(d.vpcID),
		DryRun:            aws.Bool(false),
	}

	// Send the attachement request:
	if _, err := d.svcEC2.AttachInternetGateway(params); err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	log.WithField("cmd", d.command+":ec2").
		Info("- Internet gateway attached to VPC")

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
		GatewayId:            aws.String(d.internetGatewayID),
	}

	// Send the route request:
	if _, err := d.svcEC2.CreateRoute(params); err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	log.WithField("cmd", d.command+":ec2").
		Info("- New default route added via internet GW")

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
	resp, err := d.svcEC2.AllocateAddress(params)
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Store the EIP ID:
	d.allocationID = *resp.AllocationId
	log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": d.allocationID}).
		Info("- New elastic IP allocated")

	return nil
}

//-----------------------------------------------------------------------------
// func: associateElasticIP
//-----------------------------------------------------------------------------

func (d *Data) associateElasticIP() error {

	// Wait until instance is running:
	if err := d.svcEC2.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(d.instanceID)},
	}); err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
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
	resp, err := d.svcEC2.AssociateAddress(params)
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	log.WithFields(log.Fields{
		"cmd": d.command + ":ec2", "id": *resp.AssociationId}).
		Info("- New elastic IP association")

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
	resp, err := d.svcEC2.CreateNatGateway(params)
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Store the NAT gateway ID:
	d.natGatewayID = *resp.NatGateway.NatGatewayId
	log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": d.natGatewayID}).
		Info("- New NAT gateway requested")

	// Wait until the NAT gateway is available:
	log.WithField("cmd", d.command+":ec2").
		Info("- Waiting until NAT gateway is available")
	if err := d.svcEC2.WaitUntilNatGatewayAvailable(&ec2.DescribeNatGatewaysInput{
		NatGatewayIds: []*string{aws.String(d.natGatewayID)},
	}); err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
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
	if _, err := d.svcEC2.CreateRoute(params); err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	log.WithField("cmd", d.command+":ec2").
		Info("- New default route added via NAT gateway")

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
	listRsp, err := d.svcIAM.ListPolicies(listPrms)
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Check whether the policy exists:
	for _, v := range listRsp.Policies {
		if *v.PolicyName == "REX-Ray" {
			d.rexrayPolicyARN = *listRsp.Policies[0].Arn
			log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": *listRsp.
				Policies[0].PolicyId}).Info("- Using existing REX-Ray security policy")
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
	policyRsp, err := d.svcIAM.CreatePolicy(policyPrms)
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Store the policy ARN:
	d.rexrayPolicyARN = *policyRsp.Policy.Arn
	log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": *policyRsp.Policy.
		PolicyId}).Info("- Setup REX-Ray security policy")

	return nil
}

//-----------------------------------------------------------------------------
// func: atachRexrayPolicy
//-----------------------------------------------------------------------------

func (d *Data) attachRexrayPolicy() error {

	// Forge the attachment request:
	params := &iam.AttachRolePolicyInput{
		PolicyArn: aws.String(d.rexrayPolicyARN),
		RoleName:  aws.String("katoNode"),
	}

	// Send the attachement request:
	_, err := d.svcIAM.AttachRolePolicy(params)
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	log.WithField("cmd", d.command+":ec2").
		Info("- REX-Ray policy attached to katoNode")

	return nil
}

//-----------------------------------------------------------------------------
// func: createSecurityRoles
//-----------------------------------------------------------------------------

func (d *Data) createSecurityRoles() error {

	// Map to iterate:
	grps := map[string]map[string]string{
		"master": map[string]string{"name": "katoMaster", "roleID": ""},
		"node":   map[string]string{"name": "katoNode", "roleID": ""},
		"edge":   map[string]string{"name": "katoEdge", "roleID": ""},
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
	for _, v := range grps {

		// Forge the role request:
		params := &iam.CreateRoleInput{
			AssumeRolePolicyDocument: aws.String(policy),
			RoleName:                 aws.String(v["name"]),
			Path:                     aws.String("/kato/"),
		}

		// Send the role request:
		resp, err := d.svcIAM.CreateRole(params)
		if err != nil {
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				if reqErr.StatusCode() == 409 {
					continue
				}
			}
			log.WithField("cmd", d.command+":ec2").Error(err)
			return err
		}

		// Locally store the role ID:
		v["roleID"] = *resp.Role.RoleId
		log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": v["roleID"]}).
			Info("- New " + v["name"] + " IAM role")
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
// func: createSecurityGroups
//-----------------------------------------------------------------------------

func (d *Data) createSecurityGroups() error {

	// Map to iterate:
	grps := map[string]map[string]string{
		"master": map[string]string{"Desc": "master", "SecGrpID": ""},
		"node":   map[string]string{"Desc": "node", "SecGrpID": ""},
		"edge":   map[string]string{"Desc": "edge", "SecGrpID": ""},
	}

	// For each security group:
	for k, v := range grps {

		// Forge the group request:
		params := &ec2.CreateSecurityGroupInput{
			Description: aws.String(d.Domain + " " + v["Desc"]),
			GroupName:   aws.String(k),
			DryRun:      aws.Bool(false),
			VpcId:       aws.String(d.vpcID),
		}

		// Send the group request:
		resp, err := d.svcEC2.CreateSecurityGroup(params)
		if err != nil {
			log.WithField("cmd", d.command+":ec2").Error(err)
			return err
		}

		// Locally store the group ID:
		v["SecGrpID"] = *resp.GroupId
		log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": v["SecGrpID"]}).
			Info("- New " + k + " security group")

		// Tag the group:
		if err = d.tag(v["SecGrpID"], "Name", d.Domain+" "+k); err != nil {
			return err
		}
	}

	// Store security groups IDs:
	d.masterSecGrp = grps["master"]["SecGrpID"]
	d.nodeSecGrp = grps["node"]["SecGrpID"]
	d.edgeSecGrp = grps["edge"]["SecGrpID"]

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
	_, err := d.svcEC2.AuthorizeSecurityGroupIngress(params)
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": "master"}).
		Info("- New firewall rules defined")

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
	_, err := d.svcEC2.AuthorizeSecurityGroupIngress(params)
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": "node"}).
		Info("- New firewall rules defined")

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
	_, err := d.svcEC2.AuthorizeSecurityGroupIngress(params)
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": "edge"}).
		Info("- New firewall rules defined")

	return nil
}

//-----------------------------------------------------------------------------
// func: exposeIdentifiers
//-----------------------------------------------------------------------------

func (d *Data) exposeIdentifiers() error {

	type identifiers struct {
		VpcCidrBlock      string
		VpcID             string
		MainRouteTableID  string
		IntSubnetCidr     string
		ExtSubnetCidr     string
		InternalSubnetID  string
		ExternalSubnetID  string
		InternetGatewayID string
		AllocationID      string
		NatGatewayID      string
		RouteTableID      string
		MasterSecGrp      string
		NodeSecGrp        string
		EdgeSecGrp        string
	}

	ids := identifiers{
		VpcCidrBlock:      d.VpcCidrBlock,
		VpcID:             d.vpcID,
		MainRouteTableID:  d.mainRouteTableID,
		IntSubnetCidr:     d.IntSubnetCidr,
		ExtSubnetCidr:     d.ExtSubnetCidr,
		InternalSubnetID:  d.IntSubnetID,
		ExternalSubnetID:  d.ExtSubnetID,
		InternetGatewayID: d.internetGatewayID,
		AllocationID:      d.allocationID,
		NatGatewayID:      d.natGatewayID,
		RouteTableID:      d.routeTableID,
		MasterSecGrp:      d.masterSecGrp,
		NodeSecGrp:        d.nodeSecGrp,
		EdgeSecGrp:        d.edgeSecGrp,
	}

	// Marshal the data:
	idsJSON, err := json.Marshal(ids)
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Return on success:
	fmt.Println(string(idsJSON))
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
	if _, err := d.svcEC2.CreateTags(params); err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	return nil
}
