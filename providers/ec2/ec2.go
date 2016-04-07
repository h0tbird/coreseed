package ec2

//----------------------------------------------------------------------------
// Package factored import statement:
//----------------------------------------------------------------------------

import (

	// Stdlib:
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"

	// Community:
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

//----------------------------------------------------------------------------
// Typedefs:
//----------------------------------------------------------------------------

// Data contains variables used by EC2 API.
type Data struct {
	MasterCount        int
	NodeCount          int
	EdgeCount          int
	EtcdToken          string
	Ns1ApiKey          string
	CaCert             string
	Region             string
	SubnetIDs          string
	ImageID            string
	KeyPair            string
	InstanceType       string
	Hostname           string
	ElasticIP          string
	VpcCidrBlock       string
	VpcID              string
	MainRouteTableID   string
	Domain             string
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

//--------------------------------------------------------------------------
// func: Deploy
//--------------------------------------------------------------------------

// Deploy Kato's infrastructure on Amazon EC2.
func (d *Data) Deploy() error {

	//------------------------
	// Setup the environment:
	//------------------------

	log.WithField("cmd", "deploy-ec2").Info("Setup the EC2 environment")
	cmd := exec.Command("katoctl", "setup-ec2",
		"--domain", d.Domain,
		"--region", d.Region)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.WithField("cmd", "deploy-ec2").Error(err)
		return err
	}

	//----------------------
	// Deploy master nodes:
	//----------------------

	for i := 1; i <= d.MasterCount; i++ {
		log.WithField("cmd", "deploy-ec2").Info("Deploy master ", i)
	}

	//----------------------
	// Deploy worker nodes:
	//----------------------

	for i := 1; i <= d.NodeCount; i++ {
		log.WithField("cmd", "deploy-ec2").Info("Deploy node ", i)
	}

	//--------------------
	// Deploy edge nodes:
	//--------------------

	for i := 1; i <= d.EdgeCount; i++ {
		log.WithField("cmd", "deploy-ec2").Info("Deploy edge ", i)
	}

	return nil
}

//--------------------------------------------------------------------------
// func: Setup
//--------------------------------------------------------------------------

// Setup an EC2 VPC and all the related components.
func (d *Data) Setup() error {

	// Connect and authenticate to the API endpoint:
	log.WithField("cmd", "setup-ec2").Info("Connecting to region " + d.Region)
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

	// Allocate a new eIP:
	if err := d.allocateAddress(*svc); err != nil {
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

	// Return on success:
	return nil
}

//-------------------------------------------------------------------------
// func: createVpc
//-------------------------------------------------------------------------

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
		log.WithField("cmd", "setup-ec2").Error(err)
		return err
	}

	// Store the VPC ID:
	d.VpcID = *resp.Vpc.VpcId
	log.WithFields(log.Fields{"cmd": "setup-ec2", "id": d.VpcID}).
		Info("A new VPC has been created")

	// Tag the VPC:
	if err = tag(d.VpcID, "Name", d.Domain, svc); err != nil {
		return err
	}

	return nil
}

//-------------------------------------------------------------------------
// func: retrieveMainRouteTableID
//-------------------------------------------------------------------------

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
					aws.String(d.VpcID),
				},
			},
		},
	}

	// Send the description request:
	resp, err := svc.DescribeRouteTables(params)
	if err != nil {
		log.WithField("cmd", "setup-ec2").Error(err)
		return err
	}

	// Store the main route table ID:
	d.MainRouteTableID = *resp.RouteTables[0].RouteTableId
	log.WithFields(log.Fields{"cmd": "setup-ec2", "id": d.MainRouteTableID}).
		Info("New main route table added")

	return nil
}

//-------------------------------------------------------------------------
// func: createSubnets
//-------------------------------------------------------------------------

func (d *Data) createSubnets(svc ec2.EC2) error {

	// Map to iterate:
	nets := map[string]map[string]string{
		"internal": map[string]string{"SubnetCidr": d.InternalSubnetCidr, "SubnetID": ""},
		"external": map[string]string{"SubnetCidr": d.ExternalSubnetCidr, "SubnetID": ""},
	}

	// For each subnet:
	for k, v := range nets {

		// Forge the subnet request:
		params := &ec2.CreateSubnetInput{
			CidrBlock: aws.String(v["SubnetCidr"]),
			VpcId:     aws.String(d.VpcID),
			DryRun:    aws.Bool(false),
		}

		// Send the subnet request:
		resp, err := svc.CreateSubnet(params)
		if err != nil {
			log.WithField("cmd", "setup-ec2").Error(err)
			return err
		}

		// Locally store the subnet ID:
		v["SubnetID"] = *resp.Subnet.SubnetId
		log.WithFields(log.Fields{"cmd": "setup-ec2", "id": v["SubnetID"]}).
			Info("New " + k + " subnet")

		// Tag the subnet:
		if err = tag(v["SubnetID"], "Name", k, svc); err != nil {
			return err
		}
	}

	// Store subnet IDs:
	d.InternalSubnetID = nets["internal"]["SubnetID"]
	d.ExternalSubnetID = nets["external"]["SubnetID"]

	return nil
}

//-------------------------------------------------------------------------
// func: createRouteTable
//-------------------------------------------------------------------------

func (d *Data) createRouteTable(svc ec2.EC2) error {

	// Forge the route table request:
	params := &ec2.CreateRouteTableInput{
		VpcId:  aws.String(d.VpcID),
		DryRun: aws.Bool(false),
	}

	// Send the route table request:
	resp, err := svc.CreateRouteTable(params)
	if err != nil {
		log.WithField("cmd", "setup-ec2").Error(err)
		return err
	}

	// Store the route table ID:
	d.RouteTableID = *resp.RouteTable.RouteTableId
	log.WithFields(log.Fields{"cmd": "setup-ec2", "id": d.RouteTableID}).
		Info("New route table added")

	return nil
}

//-------------------------------------------------------------------------
// func: associateRouteTable
//-------------------------------------------------------------------------

func (d *Data) associateRouteTable(svc ec2.EC2) error {

	// Forge the association request:
	params := &ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(d.RouteTableID),
		SubnetId:     aws.String(d.ExternalSubnetID),
		DryRun:       aws.Bool(false),
	}

	// Send the association request:
	resp, err := svc.AssociateRouteTable(params)
	if err != nil {
		log.WithField("cmd", "setup-ec2").Error(err)
		return err
	}

	log.WithFields(log.Fields{"cmd": "setup-ec2", "id": *resp.AssociationId}).
		Info("New route table association")

	return nil
}

//-------------------------------------------------------------------------
// func: createInternetGateway
//-------------------------------------------------------------------------

func (d *Data) createInternetGateway(svc ec2.EC2) error {

	// Forge the internet gateway request:
	params := &ec2.CreateInternetGatewayInput{
		DryRun: aws.Bool(false),
	}

	// Send the internet gateway request:
	resp, err := svc.CreateInternetGateway(params)
	if err != nil {
		log.WithField("cmd", "setup-ec2").Error(err)
		return err
	}

	// Store the internet gateway ID:
	d.InternetGatewayID = *resp.InternetGateway.InternetGatewayId
	log.WithFields(log.Fields{"cmd": "setup-ec2", "id": d.InternetGatewayID}).
		Info("New internet gateway")

	return nil
}

//-------------------------------------------------------------------------
// func: attachInternetGateway
//-------------------------------------------------------------------------

func (d *Data) attachInternetGateway(svc ec2.EC2) error {

	// Forge the attachement request:
	params := &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(d.InternetGatewayID),
		VpcId:             aws.String(d.VpcID),
		DryRun:            aws.Bool(false),
	}

	// Send the attachement request:
	if _, err := svc.AttachInternetGateway(params); err != nil {
		log.WithField("cmd", "setup-ec2").Error(err)
		return err
	}

	log.WithField("cmd", "setup-ec2").Info("Internet gateway attached to VPC")

	return nil
}

//-------------------------------------------------------------------------
// func: createInternetGatewayRoute
//-------------------------------------------------------------------------

func (d *Data) createInternetGatewayRoute(svc ec2.EC2) error {

	// Forge the route request:
	params := &ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		RouteTableId:         aws.String(d.RouteTableID),
		DryRun:               aws.Bool(false),
		GatewayId:            aws.String(d.InternetGatewayID),
	}

	// Send the route request:
	if _, err := svc.CreateRoute(params); err != nil {
		log.WithField("cmd", "setup-ec2").Error(err)
		return err
	}

	log.WithField("cmd", "setup-ec2").
		Info("New default route added via internet GW")

	return nil
}

//-------------------------------------------------------------------------
// func: allocateAddress
//-------------------------------------------------------------------------

func (d *Data) allocateAddress(svc ec2.EC2) error {

	// Forge the allocation request:
	params := &ec2.AllocateAddressInput{
		Domain: aws.String("vpc"),
		DryRun: aws.Bool(false),
	}

	// Send the allocation request:
	resp, err := svc.AllocateAddress(params)
	if err != nil {
		log.WithField("cmd", "setup-ec2").Error(err)
		return err
	}

	// Store the EIP ID:
	d.AllocationID = *resp.AllocationId
	log.WithFields(log.Fields{"cmd": "setup-ec2", "id": d.AllocationID}).
		Info("New elastic IP allocated")

	return nil
}

//-------------------------------------------------------------------------
// func: createNatGateway
//-------------------------------------------------------------------------

func (d *Data) createNatGateway(svc ec2.EC2) error {

	// Forge the NAT gateway request:
	params := &ec2.CreateNatGatewayInput{
		AllocationId: aws.String(d.AllocationID),
		SubnetId:     aws.String(d.ExternalSubnetID),
		ClientToken:  aws.String("kato"),
	}

	// Send the NAT gateway request:
	resp, err := svc.CreateNatGateway(params)
	if err != nil {
		log.WithField("cmd", "setup-ec2").Error(err)
		return err
	}

	// Store the NAT gateway ID:
	d.NatGatewayID = *resp.NatGateway.NatGatewayId
	log.WithFields(log.Fields{"cmd": "setup-ec2", "id": d.NatGatewayID}).
		Info("New NAT gateway requested")

	// Wait until the NAT gateway is available:
	log.WithField("cmd", "setup-ec2").
		Info("Wait until the NAT gateway is available...")
	if err := svc.WaitUntilNatGatewayAvailable(&ec2.DescribeNatGatewaysInput{
		NatGatewayIds: []*string{aws.String(d.NatGatewayID)},
	}); err != nil {
		log.WithField("cmd", "setup-ec2").Error(err)
		return err
	}

	return nil
}

//-------------------------------------------------------------------------
// func: createNatGatewayRoute
//-------------------------------------------------------------------------

func (d *Data) createNatGatewayRoute(svc ec2.EC2) error {

	// Forge the route request:
	params := &ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		RouteTableId:         aws.String(d.MainRouteTableID),
		DryRun:               aws.Bool(false),
		NatGatewayId:         aws.String(d.NatGatewayID),
	}

	// Send the route request:
	if _, err := svc.CreateRoute(params); err != nil {
		log.WithField("cmd", "setup-ec2").Error(err)
		return err
	}

	log.WithField("cmd", "setup-ec2").
		Info("New default route added via NAT gateway")

	return nil
}

//-------------------------------------------------------------------------
// func: createSecurityGroups
//-------------------------------------------------------------------------

func (d *Data) createSecurityGroups(svc ec2.EC2) error {

	// Map to iterate:
	grps := map[string]map[string]string{
		"master-int": map[string]string{"Description": "master internal", "SecGrpID": ""},
		"node-int":   map[string]string{"Description": "node internal", "SecGrpID": ""},
		"node-ext":   map[string]string{"Description": "node external", "SecGrpID": ""},
		"edge-int":   map[string]string{"Description": "edge internal", "SecGrpID": ""},
		"edge-ext":   map[string]string{"Description": "edge external", "SecGrpID": ""},
	}

	// For each security group:
	for k, v := range grps {

		// Forge the group request:
		params := &ec2.CreateSecurityGroupInput{
			Description: aws.String(d.Domain + " " + v["Description"]),
			GroupName:   aws.String(k),
			DryRun:      aws.Bool(false),
			VpcId:       aws.String(d.VpcID),
		}

		// Send the group request:
		resp, err := svc.CreateSecurityGroup(params)
		if err != nil {
			log.WithField("cmd", "setup-ec2").Error(err)
			return err
		}

		// Locally store the group ID:
		v["SecGrpID"] = *resp.GroupId
		log.WithFields(log.Fields{"cmd": "setup-ec2", "id": v["SecGrpID"]}).
			Info("New " + k + " security group")

		// Tag the group:
		if err = tag(v["SecGrpID"], "Name", d.Domain+" "+k, svc); err != nil {
			return err
		}
	}

	// Store security groups IDs:
	d.MasterIntSecGrp = grps["master-int"]["SecGrpID"]
	d.NodeIntSecGrp = grps["node-int"]["SecGrpID"]
	d.NodeExtSecGrp = grps["node-ext"]["SecGrpID"]
	d.EdgeIntSecGrp = grps["edge-int"]["SecGrpID"]
	d.EdgeExtSecGrp = grps["edge-ext"]["SecGrpID"]

	return nil
}

//-------------------------------------------------------------------------
// func: tag
//-------------------------------------------------------------------------

func tag(resource, key, value string, svc ec2.EC2) error {

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
		log.WithField("cmd", "setup-ec2").Error(err)
		return err
	}

	return nil
}

//--------------------------------------------------------------------------
// func: Run
//--------------------------------------------------------------------------

// Run uses EC2 API to launch a new instance.
func (d *Data) Run(udata []byte) error {

	// Connect and authenticate to the API endpoint:
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(d.Region)}))

	// Forge the network interfaces:
	var networkInterfaces []*ec2.InstanceNetworkInterfaceSpecification
	subnetIDs := strings.Split(d.SubnetIDs, ",")

	for i := 0; i < len(subnetIDs); i++ {

		// Forge the security group ids:
		var securityGroupIds []*string
		for _, gid := range strings.Split(subnetIDs[i], ":")[1:] {
			securityGroupIds = append(securityGroupIds, aws.String(gid))
		}

		iface := ec2.InstanceNetworkInterfaceSpecification{
			DeleteOnTermination: aws.Bool(true),
			DeviceIndex:         aws.Int64(int64(i)),
			Groups:              securityGroupIds,
			SubnetId:            aws.String(strings.Split(subnetIDs[i], ":")[0]),
		}

		networkInterfaces = append(networkInterfaces, &iface)
	}

	// Send the request:
	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:           aws.String(d.ImageID),
		MinCount:          aws.Int64(1),
		MaxCount:          aws.Int64(1),
		KeyName:           aws.String(d.KeyPair),
		InstanceType:      aws.String(d.InstanceType),
		NetworkInterfaces: networkInterfaces,
		UserData:          aws.String(base64.StdEncoding.EncodeToString([]byte(udata))),
	})

	if err != nil {
		log.WithField("cmd", "run-ec2").Error(err)
		return err
	}

	// Pretty-print the response data:
	fmt.Println("Created instance", *runResult.Instances[0].InstanceId)

	// Add tags to the created instance:
	_, err = svc.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{runResult.Instances[0].InstanceId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(d.Hostname),
			},
		},
	})

	if err != nil {
		log.WithField("cmd", "run-ec2").Error(err)
		return err
	}

	// Allocate an elastic IP address:
	if d.ElasticIP == "true" {

		params := &ec2.AllocateAddressInput{
			Domain: aws.String("vpc"),
			DryRun: aws.Bool(false),
		}

		// Send the request:
		resp, err := svc.AllocateAddress(params)
		if err != nil {
			log.WithField("cmd", "run-ec2").Error(err)
			return err
		}

		// Pretty-print the response data:
		fmt.Println(resp)
	}

	// Return on success:
	return nil
}
