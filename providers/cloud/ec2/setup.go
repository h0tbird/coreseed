package ec2

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"os"
	"strings"
	"sync"
	"time"

	// Community:
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iam"
)

//-----------------------------------------------------------------------------
// func: Setup
//-----------------------------------------------------------------------------

// Setup VPC, IAM and EC2 components.
func (d *Data) Setup() {

	// Set current command:
	d.command = "setup"
	d.setupAPIEndpoints()

	// Load state from state file (if any):
	if err := d.loadState(); err != nil {
		if !strings.Contains(err.Error(), "no such file or directory") {
			log.WithField("cmd", "ec2:"+d.command).Fatal(err)
		}
	}

	// Create the VPC:
	if err := d.createVPC(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Setup a wait group:
	var wg sync.WaitGroup

	// Setup VPC and IAM:
	wg.Add(2)
	go d.setupVPCNetwork(&wg)
	go d.setupIAMSecurity(&wg)
	wg.Wait()

	// Setup the EC2 and ELB:
	wg.Add(2)
	go d.setupEC2Firewall(&wg)
	go d.setupEC2Balancer(&wg)
	wg.Wait()

	// Dump state to file:
	if err := d.dumpState(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}
}

//-----------------------------------------------------------------------------
// func: setupAPIEndpoints
//-----------------------------------------------------------------------------

func (d *Data) setupAPIEndpoints() {

	// Log this acction:
	log.WithField("cmd", "ec2:"+d.command).
		Info("Connecting to region " + d.Region)

	// Connect and authenticate to the API endpoints:
	d.ec2 = ec2.New(session.New(&aws.Config{Region: aws.String(d.Region)}))
	d.iam = iam.New(session.New(&aws.Config{Region: aws.String(d.Region)}))
	d.elb = elb.New(session.New(&aws.Config{Region: aws.String(d.Region)}))
}

//-----------------------------------------------------------------------------
// func: createVPC
//-----------------------------------------------------------------------------

func (d *Data) createVPC() error {

	// Return if already defined:
	if d.VpcID != "" {
		log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.VpcID}).
			Info("Using defined VPC")
		return nil
	}

	// Forge the VPC request:
	params := &ec2.CreateVpcInput{
		CidrBlock:       aws.String(d.VpcCidrBlock),
		InstanceTenancy: aws.String("default"),
	}

	// Send the VPC request:
	resp, err := d.ec2.CreateVpc(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Store the VPC ID:
	d.VpcID = *resp.Vpc.VpcId
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.VpcID}).
		Info("New EC2 VPC created")

	// Wait until VPC is available:
	if err := d.ec2.WaitUntilVpcAvailable(&ec2.DescribeVpcsInput{
		VpcIds: []*string{aws.String(d.VpcID)},
	}); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Tag the VPC:
	if err = d.tag(d.VpcID, "Name", d.Domain); err != nil {
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
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Create the external and internal subnets:
	if err := d.createSubnets(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Create a route table (ext):
	if err := d.createRouteTable(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Associate the route table to the external subnet:
	if err := d.associateRouteTable(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Create the internet gateway:
	if err := d.createInternetGateway(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Attach internet gateway to VPC:
	if err := d.attachInternetGateway(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Create a default route via internet GW (ext):
	if err := d.createInternetGatewayRoute(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Allocate a new elastic IP:
	if err := d.allocateElasticIP(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	if d.IntSubnetCidr != "" {

		// Create a NAT gateway:
		if err := d.createNatGateway(); err != nil {
			log.WithField("cmd", "ec2:"+d.command).Fatal(err)
		}

		// Create a default route via NAT GW (int):
		if err := d.createNatGatewayRoute(); err != nil {
			log.WithField("cmd", "ec2:"+d.command).Fatal(err)
		}
	}
}

//-----------------------------------------------------------------------------
// func: retrieveMainRouteTableID
//-----------------------------------------------------------------------------

func (d *Data) retrieveMainRouteTableID() error {

	// Forge the description request:
	params := &ec2.DescribeRouteTablesInput{
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
	resp, err := d.ec2.DescribeRouteTables(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Store the main route table ID:
	if len(resp.RouteTables) > 0 {
		d.MainRouteTableID = *resp.RouteTables[0].RouteTableId
		log.WithFields(log.Fields{
			"cmd": "ec2:" + d.command, "id": d.MainRouteTableID}).
			Info("VPC main route table acquired")
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: createSubnets
//-----------------------------------------------------------------------------

func (d *Data) createSubnets() error {

	// Map to iterate:
	nets := map[string]map[string]string{
		"internal": {
			"SubnetCidr": d.IntSubnetCidr, "SubnetID": d.IntSubnetID},
		"external": {
			"SubnetCidr": d.ExtSubnetCidr, "SubnetID": d.ExtSubnetID},
	}

	// For each subnet:
	for k, v := range nets {

		if v["SubnetCidr"] != "" && v["SubnetID"] == "" {

			// Forge the subnet request:
			params := &ec2.CreateSubnetInput{
				CidrBlock:        aws.String(v["SubnetCidr"]),
				VpcId:            aws.String(d.VpcID),
				AvailabilityZone: aws.String(d.Region + d.Zone),
			}

			// Send the subnet request:
			resp, err := d.ec2.CreateSubnet(params)
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

	// Return if already defined:
	if d.RouteTableID != "" {
		log.WithFields(
			log.Fields{"cmd": "ec2:" + d.command, "id": d.RouteTableID}).
			Info("Using defined route table")
		return nil
	}

	// Forge the route table request:
	params := &ec2.CreateRouteTableInput{
		VpcId: aws.String(d.VpcID),
	}

	// Send the route table request:
	resp, err := d.ec2.CreateRouteTable(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Store the route table ID:
	d.RouteTableID = *resp.RouteTable.RouteTableId
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.RouteTableID}).
		Info("New route table added")

	return nil
}

//-----------------------------------------------------------------------------
// func: associateRouteTable
//-----------------------------------------------------------------------------

func (d *Data) associateRouteTable() error {

	// Forge the association request:
	params := &ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(d.RouteTableID),
		SubnetId:     aws.String(d.ExtSubnetID),
	}

	// Send the association request:
	resp, err := d.ec2.AssociateRouteTable(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	log.WithFields(log.Fields{
		"cmd": "ec2:" + d.command, "id": *resp.AssociationId}).
		Info("Route table association")

	return nil
}

//-----------------------------------------------------------------------------
// func: createInternetGateway
//-----------------------------------------------------------------------------

func (d *Data) createInternetGateway() error {

	// Return if already defined:
	if d.InetGatewayID != "" {
		log.WithFields(log.Fields{
			"cmd": "ec2:" + d.command, "id": d.InetGatewayID}).
			Info("Using defined internet gateway")
		return nil
	}

	// Forge the internet gateway request:
	params := &ec2.CreateInternetGatewayInput{}

	// Send the internet gateway request:
	resp, err := d.ec2.CreateInternetGateway(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Store the internet gateway ID:
	d.InetGatewayID = *resp.InternetGateway.InternetGatewayId
	log.WithFields(log.Fields{
		"cmd": "ec2:" + d.command, "id": d.InetGatewayID}).
		Info("New internet gateway")

	return nil
}

//-----------------------------------------------------------------------------
// func: attachInternetGateway
//-----------------------------------------------------------------------------

func (d *Data) attachInternetGateway() error {

	// Forge the attachement request:
	params := &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(d.InetGatewayID),
		VpcId:             aws.String(d.VpcID),
	}

	// Send the attachement request:
	for i := 0; i < 5; i++ {
		if _, err := d.ec2.AttachInternetGateway(params); err != nil {
			if ec2err, ok := err.(awserr.Error); ok {
				switch ec2err.Code() {
				case "InvalidInternetGatewayID.NotFound":
					time.Sleep(1e9)
					continue
				case "Resource.AlreadyAssociated":
					log.WithField("cmd", "ec2:"+d.command).
						Info("Internet gateway already attached to VPC")
					return nil
				}
			}
			log.WithField("cmd", "ec2:"+d.command).Error(err)
			return err
		}
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
		RouteTableId:         aws.String(d.RouteTableID),
		GatewayId:            aws.String(d.InetGatewayID),
	}

	// Send the route request:
	if _, err := d.ec2.CreateRoute(params); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	log.WithField("cmd", "ec2:"+d.command).
		Info("Default route added via internet GW")

	return nil
}

//-----------------------------------------------------------------------------
// func: allocateElasticIP
//-----------------------------------------------------------------------------

func (d *Data) allocateElasticIP() error {

	// Return if already defined:
	if d.AllocationID != "" {
		log.WithFields(
			log.Fields{"cmd": "ec2:" + d.command, "id": d.AllocationID}).
			Info("Using defined elastic IP")
		return nil
	}

	// Forge the allocation request:
	params := &ec2.AllocateAddressInput{
		Domain: aws.String("vpc"),
	}

	// Send the allocation request:
	resp, err := d.ec2.AllocateAddress(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Store the EIP ID:
	d.AllocationID = *resp.AllocationId
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.AllocationID}).
		Info("New elastic IP allocated")

	return nil
}

//-----------------------------------------------------------------------------
// func: createNatGateway
//-----------------------------------------------------------------------------

func (d *Data) createNatGateway() error {

	// Forge the NAT gateway request:
	params := &ec2.CreateNatGatewayInput{
		AllocationId: aws.String(d.AllocationID),
		SubnetId:     aws.String(d.ExtSubnetID),
		ClientToken:  aws.String(d.Domain),
	}

	// Send the NAT gateway request:
	resp, err := d.ec2.CreateNatGateway(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Store the NAT gateway ID:
	d.NatGatewayID = *resp.NatGateway.NatGatewayId
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.NatGatewayID}).
		Info("New NAT gateway requested")

	// Wait until the NAT gateway is available:
	log.WithField("cmd", "ec2:"+d.command).
		Info("Waiting until NAT gateway is available")
	if err := d.ec2.WaitUntilNatGatewayAvailable(&ec2.DescribeNatGatewaysInput{
		NatGatewayIds: []*string{aws.String(d.NatGatewayID)},
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
		RouteTableId:         aws.String(d.MainRouteTableID),
		NatGatewayId:         aws.String(d.NatGatewayID),
	}

	// Send the route request:
	if _, err := d.ec2.CreateRoute(params); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	log.WithField("cmd", "ec2:"+d.command).
		Info("New default route added via NAT gateway")

	return nil
}

//-----------------------------------------------------------------------------
// func: setupIAMSecurity
//-----------------------------------------------------------------------------

func (d *Data) setupIAMSecurity(wg *sync.WaitGroup) {

	// Decrement:
	defer wg.Done()

	// Create REX-Ray policy:
	if err := d.createRexrayPolicy(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Create IAM role:
	if err := d.createIAMRole(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Create instance profile:
	d.createInstanceProfile()

	// Attach policies to IAM role:
	for _, policy := range [2]string{
		"arn:aws:iam::aws:policy/AmazonS3FullAccess",
		d.RexrayPolicy,
	} {
		if err := d.attachPolicyToRole(policy, "kato"); err != nil {
			log.WithField("cmd", "ec2:"+d.command).Fatal(err)
		}
	}

	// Add IAM role to instance profile:
	if err := d.addIAMRoleToInstanceProfile(); err != nil {
		os.Exit(1)
	}
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
	listRsp, err := d.iam.ListPolicies(listPrms)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Check whether the policy exists:
	for _, v := range listRsp.Policies {
		if *v.PolicyName == "REX-Ray" {
			d.RexrayPolicy = *listRsp.Policies[0].Arn
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
	policyRsp, err := d.iam.CreatePolicy(policyPrms)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Store the policy ARN:
	d.RexrayPolicy = *policyRsp.Policy.Arn
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": *policyRsp.Policy.
		PolicyId}).Info("Setup REX-Ray security policy")

	return nil
}

//-----------------------------------------------------------------------------
// func: attachPolicyToRole
//-----------------------------------------------------------------------------

func (d *Data) attachPolicyToRole(policy, role string) error {

	// Forge the attachment request:
	params := &iam.AttachRolePolicyInput{
		PolicyArn: aws.String(policy),
		RoleName:  aws.String(role),
	}

	// Send the attachement request:
	_, err := d.iam.AttachRolePolicy(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Log the policy attachment:
	split := strings.Split(policy, "/")
	log.WithField("cmd", "ec2:"+d.command).
		Info("Policy " + split[len(split)-1] + " attached to " + role)

	return nil
}

//-----------------------------------------------------------------------------
// func: createIAMRole
//-----------------------------------------------------------------------------

func (d *Data) createIAMRole() error {

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

	// Forge the role request:
	params := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(policy),
		RoleName:                 aws.String("kato"),
		Path:                     aws.String("/kato/"),
	}

	// Send the role request:
	resp, err := d.iam.CreateRole(params)
	if err != nil {
		if reqErr, ok := err.(awserr.RequestFailure); ok {
			if reqErr.StatusCode() == 409 {
				return nil
			}
		}
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Locally store the role ID:
	d.KatoRoleID = *resp.Role.RoleId
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.KatoRoleID}).
		Info("New kato IAM role")

	return nil
}

//-----------------------------------------------------------------------------
// func: createInstanceProfile
//-----------------------------------------------------------------------------

func (d *Data) createInstanceProfile() {

	// Forge the profile request:
	params := &iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String("kato"),
		Path:                aws.String("/kato/"),
	}

	// Send the profile request:
	resp, err := d.iam.CreateInstanceProfile(params)
	if err != nil {
		if reqErr, ok := err.(awserr.RequestFailure); ok {
			if reqErr.StatusCode() == 409 {
				return
			}
		}
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Wait until the instance profile exists:
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command,
		"id": *resp.InstanceProfile.InstanceProfileId}).
		Info("Waiting until kato profile exists")
	if err := d.iam.WaitUntilInstanceProfileExists(
		&iam.GetInstanceProfileInput{
			InstanceProfileName: aws.String("kato"),
		}); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}
}

//-----------------------------------------------------------------------------
// func: addIAMRoleToInstanceProfile
//-----------------------------------------------------------------------------

func (d *Data) addIAMRoleToInstanceProfile() error {

	// Forge the addition request:
	params := &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: aws.String("kato"),
		RoleName:            aws.String("kato"),
	}

	// Send the addition request:
	if _, err := d.iam.AddRoleToInstanceProfile(params); err != nil {
		if reqErr, ok := err.(awserr.RequestFailure); ok {
			if reqErr.StatusCode() == 409 {
				return nil
			}
		}
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Log the addition request:
	log.WithField("cmd", "ec2:"+d.command).
		Info("New kato IAM role added to profile")

	return nil
}

//-----------------------------------------------------------------------------
// func: setupEC2Firewall
//-----------------------------------------------------------------------------

func (d *Data) setupEC2Firewall(wg *sync.WaitGroup) {

	// Decrement:
	defer wg.Done()

	// Create quorum security group:
	if err := d.createSecurityGroup("quorum", &d.QuorumSecGrp); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Create master security group:
	if err := d.createSecurityGroup("master", &d.MasterSecGrp); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Create worker security group:
	if err := d.createSecurityGroup("worker", &d.WorkerSecGrp); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Create border security group:
	if err := d.createSecurityGroup("border", &d.BorderSecGrp); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Setup quorum nodes firewall:
	if err := d.firewallQuorum(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Setup master nodes firewall:
	if err := d.firewallMaster(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Setup worker nodes firewall:
	if err := d.firewallWorker(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Setup border nodes firewall:
	if err := d.firewallBorder(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}
}

//-----------------------------------------------------------------------------
// func: createSecurityGroup
//-----------------------------------------------------------------------------

func (d *Data) createSecurityGroup(name string, id *string) error {

	// Return if already defined:
	if *id != "" {
		log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": *id}).
			Info("Using defined " + name + " security group")
		return nil
	}

	// Forge the group request:
	params := &ec2.CreateSecurityGroupInput{
		Description: aws.String(d.Domain + " " + name),
		GroupName:   aws.String(name),
		VpcId:       aws.String(d.VpcID),
	}

	// Send the group request:
	resp, err := d.ec2.CreateSecurityGroup(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Locally store the group ID:
	*id = *resp.GroupId
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": *id}).
		Info("New EC2 " + name + " security group")

	// Tag the group:
	if err = d.tag(*id, "Name", d.Domain+" "+name); err != nil {
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: firewallQuorum
//-----------------------------------------------------------------------------

func (d *Data) firewallQuorum() error {

	// Forge the rule request:
	params := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(d.QuorumSecGrp),
		IpPermissions: []*ec2.IpPermission{
			{
				IpProtocol: aws.String("-1"),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						GroupId: aws.String(d.QuorumSecGrp),
					},
					{
						GroupId: aws.String(d.MasterSecGrp),
					},
					{
						GroupId: aws.String(d.WorkerSecGrp),
					},
					{
						GroupId: aws.String(d.BorderSecGrp),
					},
				},
			},
		},
	}

	// Send the rule request:
	if _, err := d.ec2.AuthorizeSecurityGroupIngress(params); err != nil {
		ec2err, ok := err.(awserr.Error)
		if ok && strings.Contains(ec2err.Code(), ".Duplicate") {
			log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": "quorum"}).
				Info("Using existing firewall rules")
			return nil
		}
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": "quorum"}).
		Info("New firewall rules defined")

	return nil
}

//-----------------------------------------------------------------------------
// func: firewallMaster
//-----------------------------------------------------------------------------

func (d *Data) firewallMaster() error {

	// Forge the rule request:
	params := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(d.MasterSecGrp),
		IpPermissions: []*ec2.IpPermission{
			{
				IpProtocol: aws.String("-1"),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						GroupId: aws.String(d.QuorumSecGrp),
					},
					{
						GroupId: aws.String(d.MasterSecGrp),
					},
					{
						GroupId: aws.String(d.WorkerSecGrp),
					},
					{
						GroupId: aws.String(d.BorderSecGrp),
					},
				},
			},
		},
	}

	// Send the rule request:
	if _, err := d.ec2.AuthorizeSecurityGroupIngress(params); err != nil {
		ec2err, ok := err.(awserr.Error)
		if ok && strings.Contains(ec2err.Code(), ".Duplicate") {
			log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": "master"}).
				Info("Using existing firewall rules")
			return nil
		}
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": "master"}).
		Info("New firewall rules defined")

	return nil
}

//-----------------------------------------------------------------------------
// func: firewallWorker
//-----------------------------------------------------------------------------

func (d *Data) firewallWorker() error {

	// Forge the rule request:
	params := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(d.WorkerSecGrp),
		IpPermissions: []*ec2.IpPermission{
			{
				IpProtocol: aws.String("-1"),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						GroupId: aws.String(d.QuorumSecGrp),
					},
					{
						GroupId: aws.String(d.MasterSecGrp),
					},
					{
						GroupId: aws.String(d.WorkerSecGrp),
					},
					{
						GroupId: aws.String(d.BorderSecGrp),
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
	if _, err := d.ec2.AuthorizeSecurityGroupIngress(params); err != nil {
		ec2err, ok := err.(awserr.Error)
		if ok && strings.Contains(ec2err.Code(), ".Duplicate") {
			log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": "worker"}).
				Info("Using existing firewall rules")
			return nil
		}
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": "worker"}).
		Info("New firewall rules defined")

	return nil
}

//-----------------------------------------------------------------------------
// func: firewallBorder
//-----------------------------------------------------------------------------

func (d *Data) firewallBorder() error {

	// Forge the rule request:
	params := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(d.BorderSecGrp),
		IpPermissions: []*ec2.IpPermission{
			{
				IpProtocol: aws.String("-1"),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						GroupId: aws.String(d.QuorumSecGrp),
					},
					{
						GroupId: aws.String(d.MasterSecGrp),
					},
					{
						GroupId: aws.String(d.WorkerSecGrp),
					},
					{
						GroupId: aws.String(d.BorderSecGrp),
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
	if _, err := d.ec2.AuthorizeSecurityGroupIngress(params); err != nil {
		ec2err, ok := err.(awserr.Error)
		if ok && strings.Contains(ec2err.Code(), ".Duplicate") {
			log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": "border"}).
				Info("Using existing firewall rules")
			return nil
		}
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": "border"}).
		Info("New firewall rules defined")

	return nil
}

//-----------------------------------------------------------------------------
// func: setupEC2Balancer
//-----------------------------------------------------------------------------

func (d *Data) setupEC2Balancer(wg *sync.WaitGroup) {

	// Decrement:
	defer wg.Done()

	// Create the ELB security group:
	if err := d.createSecurityGroup("elb", &d.ELBSecGrp); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Create the ELB:
	if err := d.createELB(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Setup the ELB firewall:
	if err := d.firewallELB(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}
}

//-----------------------------------------------------------------------------
// func: createELB
//-----------------------------------------------------------------------------

func (d *Data) createELB() error {

	// Forge the ELB creation request:
	params := &elb.CreateLoadBalancerInput{
		Listeners: []*elb.Listener{
			{
				InstancePort:     aws.Int64(80),
				LoadBalancerPort: aws.Int64(80),
				Protocol:         aws.String("TCP"),
				InstanceProtocol: aws.String("TCP"),
			},
			{
				InstancePort:     aws.Int64(443),
				LoadBalancerPort: aws.Int64(443),
				Protocol:         aws.String("TCP"),
				InstanceProtocol: aws.String("TCP"),
			},
		},
		LoadBalancerName: aws.String(d.ClusterID),
		SecurityGroups: []*string{
			aws.String(d.ELBSecGrp),
		},
		Subnets: []*string{
			aws.String(d.ExtSubnetID),
		},
		Tags: []*elb.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(d.ClusterID),
			},
		},
	}

	// Send the ELB creation request:
	resp, err := d.elb.CreateLoadBalancer(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Store the ELB DNS name:
	d.DNSName = *resp.DNSName
	log.WithFields(log.Fields{
		"cmd": "ec2:" + d.command, "id": d.DNSName}).
		Info("New ELB DNS name created")

	return nil
}

//-----------------------------------------------------------------------------
// func: firewallELB
//-----------------------------------------------------------------------------

func (d *Data) firewallELB() error {

	// Forge the rule request:
	params := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(d.ELBSecGrp),
		IpPermissions: []*ec2.IpPermission{
			{
				IpProtocol: aws.String("-1"),
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
	if _, err := d.ec2.AuthorizeSecurityGroupIngress(params); err != nil {
		ec2err, ok := err.(awserr.Error)
		if ok && strings.Contains(ec2err.Code(), ".Duplicate") {
			log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": "elb"}).
				Info("Using existing firewall rules")
			return nil
		}
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": "elb"}).
		Info("New firewall rules defined")

	return nil
}
