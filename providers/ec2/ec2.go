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
	Region        string
	SubnetIds     string
	ImageID       string
	KeyPair       string
	InsType       string
	HostName      string
	ElasticIP     string
	VpcCidrBlock  string
	VpcId         string
	VpcNameTag    string
	IntSubnetCidr string
	ExtSubnetCidr string
	IntSubnetID   string
	ExtSubnetID   string
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

	// Return on success:
	return nil
}

//-------------------------------------------------------------------------
// func: createVpc
//-------------------------------------------------------------------------

func (d *Data) createVpc(svc ec2.EC2) error {

	// Forge the VPC request:
	prmsVpc := &ec2.CreateVpcInput {
		CidrBlock:       aws.String(d.VpcCidrBlock),
		DryRun:          aws.Bool(false),
		InstanceTenancy: aws.String("default"),
	}

	// Send the VPC request:
	log.Printf("[setup-ec2] INFO Creating the VPC\n")
	rspVpc, err := svc.CreateVpc(prmsVpc)
	if err != nil {
		return err
	}

	d.VpcId = *rspVpc.Vpc.VpcId
	log.Printf("[setup-ec2] INFO VpcId: %s\n", d.VpcId)

	// Forge the tag request:
	prmsTag := &ec2.CreateTagsInput{
		Resources: []*string{
			aws.String(*rspVpc.Vpc.VpcId),
		},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(d.VpcNameTag),
			},
		},
		DryRun: aws.Bool(false),
	}

	// Send the tag request:
	_, err = svc.CreateTags(prmsTag)
	if err != nil {
		return err
	}

	// Return on success:
	log.Printf("[setup-ec2] INFO %s tagged as %s\n", d.VpcId, d.VpcNameTag)
	return nil
}

//-------------------------------------------------------------------------
// func: createSubnets
//-------------------------------------------------------------------------

func (d *Data) createSubnets(svc ec2.EC2) error {

	// Forge the internal subnet request:
	prmsInt := &ec2.CreateSubnetInput{
		CidrBlock: aws.String(d.IntSubnetCidr),
		VpcId:     aws.String(d.VpcId),
		DryRun:    aws.Bool(false),
	}

	// Send the subnet request:
	log.Printf("[setup-ec2] INFO Creating the internal subnet\n")
	rspInt, err := svc.CreateSubnet(prmsInt)
	if err != nil {
		return err
	}

	d.IntSubnetID = *rspInt.Subnet.SubnetId
	log.Printf("[setup-ec2] INFO Internal subnet ID: %s\n", d.IntSubnetID)

	// Forge the external subnet request:
	prmsExt := &ec2.CreateSubnetInput{
		CidrBlock: aws.String(d.ExtSubnetCidr),
		VpcId:     aws.String(d.VpcId),
		DryRun:    aws.Bool(false),
	}

	// Send the subnet request:
	log.Printf("[setup-ec2] INFO Creating the external subnet\n")
	rspExt, err := svc.CreateSubnet(prmsExt)
	if err != nil {
		return err
	}

	d.ExtSubnetID = *rspExt.Subnet.SubnetId
	log.Printf("[setup-ec2] INFO External subnet ID: %s\n", d.ExtSubnetID)

	// Return on success:
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
	subnetIds := strings.Split(d.SubnetIds, ",")

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
		ImageId:           aws.String(d.ImageID),
		MinCount:          aws.Int64(1),
		MaxCount:          aws.Int64(1),
		KeyName:           aws.String(d.KeyPair),
		InstanceType:      aws.String(d.InsType),
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
				Value: aws.String(d.HostName),
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
