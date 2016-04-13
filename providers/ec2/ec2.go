package ec2

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	// Community:
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/h0tbird/kato/katool"
)

//-----------------------------------------------------------------------------
// Typedefs:
//-----------------------------------------------------------------------------

// Data contains variables used by this EC2 provider.
type Data struct {
	MasterCount        int    //  deploy:ec2 |           |       |
	NodeCount          int    //  deploy:ec2 |           |       |
	EdgeCount          int    //  deploy:ec2 |           |       |
	MasterType         string //  deploy:ec2 |           |       |
	NodeType           string //  deploy:ec2 |           |       |
	EdgeType           string //  deploy:ec2 |           |       |
	Channel            string //  deploy:ec2 |           |       |
	EtcdToken          string //  deploy:ec2 |           | udata |
	Ns1ApiKey          string //  deploy:ec2 |           | udata |
	CaCert             string //  deploy:ec2 |           | udata |
	Domain             string //  deploy:ec2 | setup:ec2 | udata |
	Region             string //  deploy:ec2 | setup:ec2 |       | run:ec2
	command            string //  deploy:ec2 | setup:ec2 |       | run:ec2
	VpcCidrBlock       string //             | setup:ec2 |       |
	vpcID              string //             | setup:ec2 |       |
	mainRouteTableID   string //             | setup:ec2 |       |
	InternalSubnetCidr string //             | setup:ec2 |       |
	ExternalSubnetCidr string //             | setup:ec2 |       |
	internetGatewayID  string //             | setup:ec2 |       |
	natGatewayID       string //             | setup:ec2 |       |
	routeTableID       string //             | setup:ec2 |       |
	masterIntSecGrp    string //             | setup:ec2 |       |
	nodeIntSecGrp      string //             | setup:ec2 |       |
	nodeExtSecGrp      string //             | setup:ec2 |       |
	edgeIntSecGrp      string //             | setup:ec2 |       |
	edgeExtSecGrp      string //             | setup:ec2 |       |
	allocationID       string //             | setup:ec2 |       | run:ec2
	IntSubnetID        string //             | setup:ec2 |       | run:ec2
	ExtSubnetID        string //             | setup:ec2 |       | run:ec2
	IntSecGrpID        string //             | setup:ec2 |       | run:ec2
	ExtSecGrpID        string //             | setup:ec2 |       | run:ec2
	instanceID         string //             |           |       | run:ec2
	SubnetIDs          string //             |           |       | run:ec2
	ImageID            string //             |           |       | run:ec2
	KeyPair            string //             |           |       | run:ec2
	InstanceType       string //             |           |       | run:ec2
	Hostname           string //             |           |       | run:ec2
	ElasticIP          string //             |           |       | run:ec2
	intIfaceID         string //             |           |       | run:ec2
	extIfaceID         string //             |           |       | run:ec2
}

//-----------------------------------------------------------------------------
// func: Deploy
//-----------------------------------------------------------------------------

// Deploy Kato's infrastructure on Amazon EC2.
func (d *Data) Deploy() error {

	// Set command to deploy:
	d.command = "deploy"

	// Setup the EC2 environment:
	if err := d.deploySetup(); err != nil {
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

	// Deploy the master nodes:
	if err := d.deployMasterNodes(); err != nil {
		return err
	}

	// Deploy the worker nodes:
	if err := d.deployWorkerNodes(); err != nil {
		return err
	}

	// Deploy the edge nodes:
	if err := d.deployEdgeNodes(); err != nil {
		return err
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
	log.WithField("cmd", d.command+":ec2").Info("- Connecting to region " + d.Region)
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(d.Region)}))

	// Run the EC2 instance:
	if err := d.runInstance(udata, *svc); err != nil {
		return err
	}

	if d.ElasticIP == "true" {

		// Allocate an elastic IP address:
		if err := d.allocateElasticIP(*svc); err != nil {
			return err
		}

		// Associate the elastic IP:
		if err := d.associateElasticIP(*svc); err != nil {
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

	// Connect and authenticate to the API endpoint:
	log.WithField("cmd", d.command+":ec2").
		Info("- Connecting to region " + d.Region)
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(d.Region)}))

	// Create the VPC:
	if err := d.createVpc(*svc); err != nil {
		return err
	}

	// Retrieve the main route table ID:
	if err := d.retrieveMainRouteTableID(*svc); err != nil {
		return err
	}

	// Create the external and internal subnets:
	if err := d.createSubnets(*svc); err != nil {
		return err
	}

	// Create a route table (ext):
	if err := d.createRouteTable(*svc); err != nil {
		return err
	}

	// Associate the route table to the external subnet:
	if err := d.associateRouteTable(*svc); err != nil {
		return err
	}

	// Create the internet gateway:
	if err := d.createInternetGateway(*svc); err != nil {
		return err
	}

	// Attach internet gateway to VPC:
	if err := d.attachInternetGateway(*svc); err != nil {
		return err
	}

	// Create a default route via internet GW (ext):
	if err := d.createInternetGatewayRoute(*svc); err != nil {
		return err
	}

	// Allocate a new elastic IP:
	if err := d.allocateElasticIP(*svc); err != nil {
		return err
	}

	// Create a NAT gateway:
	if err := d.createNatGateway(*svc); err != nil {
		return err
	}

	// Create a default route via NAT GW (int):
	if err := d.createNatGatewayRoute(*svc); err != nil {
		return err
	}

	// Create security groups:
	if err := d.createSecurityGroups(*svc); err != nil {
		return err
	}

	// Setup master nodes firewall:

	// Setup worker nodes firewall:

	// Setup edge nodes firewall:

	// Expose identifiers to stdout:
	if err := d.exposeIdentifiers(); err != nil {
		return err
	}

	// Return on success:
	return nil
}

//-----------------------------------------------------------------------------
// func: deploySetup
//-----------------------------------------------------------------------------

func (d *Data) deploySetup() error {

	// Forge the setup command:
	log.WithField("cmd", d.command+":ec2").Info("Setup the EC2 environment")
	cmdSetup := exec.Command("katoctl", "setup", "ec2",
		"--domain", d.Domain,
		"--region", d.Region)

	// Execute the setup command:
	cmdSetup.Stderr = os.Stderr
	out, err := cmdSetup.Output()
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Decode JSON data from setup:
	var dat map[string]interface{}
	if err := json.Unmarshal(out, &dat); err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Store the values:
	d.VpcCidrBlock = dat["VpcCidrBlock"].(string)
	d.vpcID = dat["VpcID"].(string)
	d.mainRouteTableID = dat["MainRouteTableID"].(string)
	d.InternalSubnetCidr = dat["InternalSubnetCidr"].(string)
	d.ExternalSubnetCidr = dat["ExternalSubnetCidr"].(string)
	d.IntSubnetID = dat["InternalSubnetID"].(string)
	d.ExtSubnetID = dat["ExternalSubnetID"].(string)
	d.internetGatewayID = dat["InternetGatewayID"].(string)
	d.allocationID = dat["AllocationID"].(string)
	d.natGatewayID = dat["NatGatewayID"].(string)
	d.routeTableID = dat["RouteTableID"].(string)
	d.masterIntSecGrp = dat["MasterIntSecGrp"].(string)
	d.nodeIntSecGrp = dat["NodeIntSecGrp"].(string)
	d.nodeExtSecGrp = dat["NodeExtSecGrp"].(string)
	d.edgeIntSecGrp = dat["EdgeIntSecGrp"].(string)
	d.edgeExtSecGrp = dat["EdgeExtSecGrp"].(string)

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

	d.ImageID = "ami-f4199987"
	return nil
}

//-----------------------------------------------------------------------------
// func: deployMasterNodes
//-----------------------------------------------------------------------------

func (d *Data) deployMasterNodes() error {

	log.WithField("cmd", d.command+":ec2").
		Info("Deploy " + strconv.Itoa(d.MasterCount) + " master nodes")
	for i := 1; i <= d.MasterCount; i++ {

		// Forge the udata command:
		cmdUdata := exec.Command("katoctl", "udata",
			"--role", "master",
			"--master-count", strconv.Itoa(d.MasterCount),
			"--hostid", strconv.Itoa(i),
			"--domain", d.Domain,
			"--ns1-api-key", d.Ns1ApiKey,
			"--ca-cert", d.CaCert,
			"--etcd-token", d.EtcdToken,
			"--gzip-udata")

		// Forge the run command:
		cmdRun := exec.Command("katoctl", "run", "ec2",
			"--hostname", "master-"+strconv.Itoa(i)+"."+d.Domain,
			"--region", d.Region,
			"--image-id", d.ImageID,
			"--instance-type", d.MasterType,
			"--key-pair", d.KeyPair,
			"--internal-subnet-id", d.IntSubnetID,
			"--internal-security-group-id", d.masterIntSecGrp)

		// Execute the pipeline:
		if err := katool.ExecutePipeline(cmdUdata, cmdRun); err != nil {
			log.WithField("cmd", d.command+":ec2").Error(err)
			return err
		}
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: deployWorkerNodes
//-----------------------------------------------------------------------------

func (d *Data) deployWorkerNodes() error {

	log.WithField("cmd", d.command+":ec2").
		Info("Deploy " + strconv.Itoa(d.NodeCount) + " worker nodes")
	for i := 1; i <= d.NodeCount; i++ {

		// Forge the udata command:
		cmdUdata := exec.Command("katoctl", "udata",
			"--role", "node",
			"--master-count", strconv.Itoa(d.MasterCount),
			"--hostid", strconv.Itoa(i),
			"--domain", d.Domain,
			"--ns1-api-key", d.Ns1ApiKey,
			"--ca-cert", d.CaCert,
			"--gzip-udata",
			"--flannel-network", "10.128.0.0/21",
			"--flannel-subnet-len", "27",
			"--flannel-subnet-min", "10.128.0.192",
			"--flannel-subnet-max", "10.128.7.224",
			"--flannel-backend", "vxlan")

		// Forge the run command:
		cmdRun := exec.Command("katoctl", "run", "ec2",
			"--hostname", "node-"+strconv.Itoa(i)+"."+d.Domain,
			"--region", d.Region,
			"--image-id", d.ImageID,
			"--instance-type", d.NodeType,
			"--key-pair", d.KeyPair,
			"--internal-subnet-id", d.IntSubnetID,
			"--external-subnet-id", d.ExtSubnetID,
			"--internal-security-group-id", d.nodeIntSecGrp,
			"--external-security-group-id", d.nodeExtSecGrp,
			"--elastic-ip", "true")

		// Execute the pipeline:
		if err := katool.ExecutePipeline(cmdUdata, cmdRun); err != nil {
			log.WithField("cmd", d.command+":ec2").Error(err)
			return err
		}
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: deployEdgeNodes
//-----------------------------------------------------------------------------

func (d *Data) deployEdgeNodes() error {

	log.WithField("cmd", d.command+":ec2").
		Info("Deploy " + strconv.Itoa(d.EdgeCount) + " edge nodes")
	for i := 1; i <= d.EdgeCount; i++ {

		// Forge the udata command:
		cmdUdata := exec.Command("katoctl", "udata",
			"--role", "edge",
			"--master-count", strconv.Itoa(d.MasterCount),
			"--hostid", strconv.Itoa(i),
			"--domain", d.Domain,
			"--ns1-api-key", d.Ns1ApiKey,
			"--ca-cert", d.CaCert,
			"--gzip-udata")

		// Forge the run command:
		cmdRun := exec.Command("katoctl", "run", "ec2",
			"--hostname", "edge-"+strconv.Itoa(i)+"."+d.Domain,
			"--region", d.Region,
			"--image-id", d.ImageID,
			"--instance-type", d.NodeType,
			"--key-pair", d.KeyPair,
			"--internal-subnet-id", d.IntSubnetID,
			"--external-subnet-id", d.ExtSubnetID,
			"--internal-security-group-id", d.edgeIntSecGrp,
			"--external-security-group-id", d.edgeExtSecGrp,
			"--elastic-ip", "true")

		// Execute the pipeline:
		if err := katool.ExecutePipeline(cmdUdata, cmdRun); err != nil {
			log.WithField("cmd", d.command+":ec2").Error(err)
			return err
		}
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: forgeNetworkInterfaces
//-----------------------------------------------------------------------------

func (d *Data) forgeNetworkInterfaces(svc ec2.EC2) []*ec2.
	InstanceNetworkInterfaceSpecification {

	var networkInterfaces []*ec2.InstanceNetworkInterfaceSpecification

	// Internal interface:
	if d.IntSubnetID != "" {

		var securityGroupIds []*string

		if d.IntSecGrpID != "" {
			securityGroupIds = append(securityGroupIds, aws.String(d.IntSecGrpID))
		}

		iface := ec2.InstanceNetworkInterfaceSpecification{
			DeleteOnTermination: aws.Bool(true),
			DeviceIndex:         aws.Int64(int64(0)),
			Groups:              securityGroupIds,
			SubnetId:            aws.String(d.IntSubnetID),
		}

		networkInterfaces = append(networkInterfaces, &iface)
	}

	// External interface:
	if d.ExtSubnetID != "" {

		var securityGroupIds []*string

		if d.ExtSecGrpID != "" {
			securityGroupIds = append(securityGroupIds, aws.String(d.ExtSecGrpID))
		}

		iface := ec2.InstanceNetworkInterfaceSpecification{
			DeleteOnTermination: aws.Bool(true),
			DeviceIndex:         aws.Int64(int64(0)),
			Groups:              securityGroupIds,
			SubnetId:            aws.String(d.IntSubnetID),
		}

		networkInterfaces = append(networkInterfaces, &iface)
	}

	return networkInterfaces
}

//-----------------------------------------------------------------------------
// func: runInstance
//-----------------------------------------------------------------------------

func (d *Data) runInstance(udata []byte, svc ec2.EC2) error {

	// Send the instance request:
	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:           aws.String(d.ImageID),
		MinCount:          aws.Int64(1),
		MaxCount:          aws.Int64(1),
		KeyName:           aws.String(d.KeyPair),
		InstanceType:      aws.String(d.InstanceType),
		NetworkInterfaces: d.forgeNetworkInterfaces(svc),
		UserData: aws.String(base64.StdEncoding.
			EncodeToString([]byte(udata))),
	})

	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Locally store the instance ID:
	d.instanceID = *runResult.Instances[0].InstanceId
	log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": d.instanceID}).
		Info("- New " + d.InstanceType + " EC2 instance requested")

	// Locally store the interface IDs:
	d.intIfaceID = *runResult.Instances[0].NetworkInterfaces[1].
		NetworkInterfaceId
	d.extIfaceID = *runResult.Instances[0].NetworkInterfaces[0].
		NetworkInterfaceId

	// Tag the instance:
	if err := d.tag(d.instanceID, "Name", d.Hostname, svc); err != nil {
		return err
	}

	// Pretty-print to stderr:
	log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": d.Hostname}).
		Info("- New EC2 instance tagged")

	return nil
}

//-----------------------------------------------------------------------------
// func: createVpc
//-----------------------------------------------------------------------------

func (d *Data) createVpc(svc ec2.EC2) error {

	// Forge the VPC request:
	params := &ec2.CreateVpcInput{
		CidrBlock:       aws.String(d.VpcCidrBlock),
		DryRun:          aws.Bool(false),
		InstanceTenancy: aws.String("default"),
	}

	// Send the VPC request:
	resp, err := svc.CreateVpc(params)
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Store the VPC ID:
	d.vpcID = *resp.Vpc.VpcId
	log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": d.vpcID}).
		Info("- New EC2 VPC created")

	// Tag the VPC:
	if err = d.tag(d.vpcID, "Name", d.Domain, svc); err != nil {
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: retrieveMainRouteTableID
//-----------------------------------------------------------------------------

func (d *Data) retrieveMainRouteTableID(svc ec2.EC2) error {

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
	resp, err := svc.DescribeRouteTables(params)
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

func (d *Data) createSubnets(svc ec2.EC2) error {

	// Map to iterate:
	nets := map[string]map[string]string{
		"internal": map[string]string{
			"SubnetCidr": d.InternalSubnetCidr, "SubnetID": ""},
		"external": map[string]string{
			"SubnetCidr": d.ExternalSubnetCidr, "SubnetID": ""},
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
		resp, err := svc.CreateSubnet(params)
		if err != nil {
			log.WithField("cmd", d.command+":ec2").Error(err)
			return err
		}

		// Locally store the subnet ID:
		v["SubnetID"] = *resp.Subnet.SubnetId
		log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": v["SubnetID"]}).
			Info("- New " + k + " subnet")

		// Tag the subnet:
		if err = d.tag(v["SubnetID"], "Name", k, svc); err != nil {
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

func (d *Data) createRouteTable(svc ec2.EC2) error {

	// Forge the route table request:
	params := &ec2.CreateRouteTableInput{
		VpcId:  aws.String(d.vpcID),
		DryRun: aws.Bool(false),
	}

	// Send the route table request:
	resp, err := svc.CreateRouteTable(params)
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

func (d *Data) associateRouteTable(svc ec2.EC2) error {

	// Forge the association request:
	params := &ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(d.routeTableID),
		SubnetId:     aws.String(d.ExtSubnetID),
		DryRun:       aws.Bool(false),
	}

	// Send the association request:
	resp, err := svc.AssociateRouteTable(params)
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

func (d *Data) createInternetGateway(svc ec2.EC2) error {

	// Forge the internet gateway request:
	params := &ec2.CreateInternetGatewayInput{
		DryRun: aws.Bool(false),
	}

	// Send the internet gateway request:
	resp, err := svc.CreateInternetGateway(params)
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

func (d *Data) attachInternetGateway(svc ec2.EC2) error {

	// Forge the attachement request:
	params := &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(d.internetGatewayID),
		VpcId:             aws.String(d.vpcID),
		DryRun:            aws.Bool(false),
	}

	// Send the attachement request:
	if _, err := svc.AttachInternetGateway(params); err != nil {
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

func (d *Data) createInternetGatewayRoute(svc ec2.EC2) error {

	// Forge the route request:
	params := &ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		RouteTableId:         aws.String(d.routeTableID),
		DryRun:               aws.Bool(false),
		GatewayId:            aws.String(d.internetGatewayID),
	}

	// Send the route request:
	if _, err := svc.CreateRoute(params); err != nil {
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

func (d *Data) allocateElasticIP(svc ec2.EC2) error {

	// Forge the allocation request:
	params := &ec2.AllocateAddressInput{
		Domain: aws.String("vpc"),
		DryRun: aws.Bool(false),
	}

	// Send the allocation request:
	resp, err := svc.AllocateAddress(params)
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

func (d *Data) associateElasticIP(svc ec2.EC2) error {

	// Forge the association request:
	params := &ec2.AssociateAddressInput{
		AllocationId:       aws.String(d.allocationID),
		AllowReassociation: aws.Bool(true),
		DryRun:             aws.Bool(false),
		NetworkInterfaceId: aws.String(d.extIfaceID),
	}

	// Send the association request:
	resp, err := svc.AssociateAddress(params)
	if err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	// Pretty-print the response data.
	fmt.Println(resp)

	return nil
}

//-----------------------------------------------------------------------------
// func: createNatGateway
//-----------------------------------------------------------------------------

func (d *Data) createNatGateway(svc ec2.EC2) error {

	// Forge the NAT gateway request:
	params := &ec2.CreateNatGatewayInput{
		AllocationId: aws.String(d.allocationID),
		SubnetId:     aws.String(d.ExtSubnetID),
		ClientToken:  aws.String(d.Domain),
	}

	// Send the NAT gateway request:
	resp, err := svc.CreateNatGateway(params)
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
	if err := svc.WaitUntilNatGatewayAvailable(&ec2.DescribeNatGatewaysInput{
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

func (d *Data) createNatGatewayRoute(svc ec2.EC2) error {

	// Forge the route request:
	params := &ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		RouteTableId:         aws.String(d.mainRouteTableID),
		DryRun:               aws.Bool(false),
		NatGatewayId:         aws.String(d.natGatewayID),
	}

	// Send the route request:
	if _, err := svc.CreateRoute(params); err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	log.WithField("cmd", d.command+":ec2").
		Info("- New default route added via NAT gateway")

	return nil
}

//-----------------------------------------------------------------------------
// func: createSecurityGroups
//-----------------------------------------------------------------------------

func (d *Data) createSecurityGroups(svc ec2.EC2) error {

	// Map to iterate:
	grps := map[string]map[string]string{
		"master-int": map[string]string{"Desc": "master internal", "SecGrpID": ""},
		"node-int":   map[string]string{"Desc": "node internal", "SecGrpID": ""},
		"node-ext":   map[string]string{"Desc": "node external", "SecGrpID": ""},
		"edge-int":   map[string]string{"Desc": "edge internal", "SecGrpID": ""},
		"edge-ext":   map[string]string{"Desc": "edge external", "SecGrpID": ""},
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
		resp, err := svc.CreateSecurityGroup(params)
		if err != nil {
			log.WithField("cmd", d.command+":ec2").Error(err)
			return err
		}

		// Locally store the group ID:
		v["SecGrpID"] = *resp.GroupId
		log.WithFields(log.Fields{"cmd": d.command + ":ec2", "id": v["SecGrpID"]}).
			Info("- New " + k + " security group")

		// Tag the group:
		if err = d.tag(v["SecGrpID"], "Name", d.Domain+" "+k, svc); err != nil {
			return err
		}
	}

	// Store security groups IDs:
	d.masterIntSecGrp = grps["master-int"]["SecGrpID"]
	d.nodeIntSecGrp = grps["node-int"]["SecGrpID"]
	d.nodeExtSecGrp = grps["node-ext"]["SecGrpID"]
	d.edgeIntSecGrp = grps["edge-int"]["SecGrpID"]
	d.edgeExtSecGrp = grps["edge-ext"]["SecGrpID"]

	return nil
}

//-----------------------------------------------------------------------------
// func: exposeIdentifiers
//-----------------------------------------------------------------------------

func (d *Data) exposeIdentifiers() error {

	type identifiers struct {
		VpcCidrBlock       string
		VpcID              string
		MainRouteTableID   string
		InternalSubnetCidr string
		ExternalSubnetCidr string
		InternalSubnetID   string
		ExternalSubnetID   string
		InternetGatewayID  string
		AllocationID       string
		NatGatewayID       string
		RouteTableID       string
		MasterIntSecGrp    string
		NodeIntSecGrp      string
		NodeExtSecGrp      string
		EdgeIntSecGrp      string
		EdgeExtSecGrp      string
	}

	ids := identifiers{
		VpcCidrBlock:       d.VpcCidrBlock,
		VpcID:              d.vpcID,
		MainRouteTableID:   d.mainRouteTableID,
		InternalSubnetCidr: d.InternalSubnetCidr,
		ExternalSubnetCidr: d.ExternalSubnetCidr,
		InternalSubnetID:   d.IntSubnetID,
		ExternalSubnetID:   d.ExtSubnetID,
		InternetGatewayID:  d.internetGatewayID,
		AllocationID:       d.allocationID,
		NatGatewayID:       d.natGatewayID,
		RouteTableID:       d.routeTableID,
		MasterIntSecGrp:    d.masterIntSecGrp,
		NodeIntSecGrp:      d.nodeIntSecGrp,
		NodeExtSecGrp:      d.nodeExtSecGrp,
		EdgeIntSecGrp:      d.edgeIntSecGrp,
		EdgeExtSecGrp:      d.edgeExtSecGrp,
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

func (d *Data) tag(resource, key, value string, svc ec2.EC2) error {

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
	if _, err := svc.CreateTags(params); err != nil {
		log.WithField("cmd", d.command+":ec2").Error(err)
		return err
	}

	return nil
}
