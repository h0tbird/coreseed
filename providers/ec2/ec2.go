package ec2

//----------------------------------------------------------------------------
// Package factored import statement:
//----------------------------------------------------------------------------

import (

	// Stdlib:
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	// Community:
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

//----------------------------------------------------------------------------
// Typedefs:
//----------------------------------------------------------------------------

// Data contains variables used by EC2 API.
type Data struct {
	Region             string
	SubnetIDs          string
	ImageID            string
	KeyPair            string
	InstanceType       string
	Hostname           string
	ElasticIP          string
	VpcCidrBlock       string
	VpcID              string
	VpcNameTag         string
	InternalSubnetCidr string
	ExternalSubnetCidr string
	InternalSubnetID   string
	ExternalSubnetID   string
	InternetGatewayID  string
	AllocationID       string
	NatGatewayID       string
}

//--------------------------------------------------------------------------
// func: Setup
//--------------------------------------------------------------------------

// Setup an EC2 VPC and all the related components.
func (d *Data) Setup() error {

	// Connect and authenticate to the API endpoint:
	log.Printf("[setup-ec2] INFO Connecting to %s\n", d.Region)
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(d.Region)}))

	// Create the VPC:
	if err := d.createVpc(*svc); err != nil {
		return err
	}

	// Create external and internal subnets:
	if err := d.createSubnets(*svc); err != nil {
		return err
	}

	// Create the internet gateway:
	if err := d.createInternetGateway(*svc); err != nil {
		return err
	}

	// Allocate a new EIP:
	if err := d.allocateAddress(*svc); err != nil {
		return err
	}

	// Create a NAT gateway:
	if err := d.createNatGateway(*svc); err != nil {
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
	prmsVpc := &ec2.CreateVpcInput{
		CidrBlock:       aws.String(d.VpcCidrBlock),
		DryRun:          aws.Bool(false),
		InstanceTenancy: aws.String("default"),
	}

	// Send the VPC request:
	rspVpc, err := svc.CreateVpc(prmsVpc)
	if err != nil {
		return err
	}

	// Store the VPC ID:
	d.VpcID = *rspVpc.Vpc.VpcId
	log.Printf("[setup-ec2] INFO New VPC %s\n", d.VpcID)

	// Tag the VPC:
	if err = tag(d.VpcID, "Name", d.VpcNameTag, svc); err != nil {
		return err
	}

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
			return err
		}

		// Locally store the subnet ID:
		v["SubnetID"] = *resp.Subnet.SubnetId
		log.Printf("[setup-ec2] INFO New %s subnet %s\n", k, v["SubnetID"])

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
		return err
	}

	// Store the internet gateway ID:
	d.InternetGatewayID = *resp.InternetGateway.InternetGatewayId
	log.Printf("[setup-ec2] INFO New internet gateway %s\n", d.InternetGatewayID)

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
		return err
	}

	// Store the EIP ID:
	d.AllocationID = *resp.AllocationId
	log.Printf("[setup-ec2] INFO New elastic IP %s\n", d.AllocationID)

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
		return err
	}

	// Store the NAT gateway ID:
	d.NatGatewayID = *resp.NatGateway.NatGatewayId
	log.Printf("[setup-ec2] INFO New NAT gateway %s\n", d.NatGatewayID)

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
	_, err := svc.CreateTags(params)
	if err != nil {
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
			return err
		}

		// Pretty-print the response data:
		fmt.Println(resp)
	}

	// Return on success:
	return nil
}
