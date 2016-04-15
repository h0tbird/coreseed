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
	masterSecGrp       string //             | setup:ec2 |       |
	nodeSecGrp         string //             | setup:ec2 |       |
	edgeSecGrp         string //             | setup:ec2 |       |
	IntSubnetID        string //             | setup:ec2 |       |
	ExtSubnetID        string //             | setup:ec2 |       |
	allocationID       string //             | setup:ec2 |       | run:ec2
	instanceID         string //             |           |       | run:ec2
	SubnetID           string //             |           |       | run:ec2
	SecGrpID           string //             |           |       | run:ec2
	ImageID            string //             |           |       | run:ec2
	KeyPair            string //             |           |       | run:ec2
	InstanceType       string //             |           |       | run:ec2
	Hostname           string //             |           |       | run:ec2
	PublicIP           string //             |           |       | run:ec2
	interfaceID        string //             |           |       | run:ec2
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
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(d.Region)}))

	// Run the EC2 instance:
	if err := d.runInstance(udata, *svc); err != nil {
		return err
	}

	if d.PublicIP == "elastic" {

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
	if err := d.masterFirewall(*svc); err != nil {
		return err
	}

	// Setup worker nodes firewall:
	if err := d.nodeFirewall(*svc); err != nil {
		return err
	}

	// Setup edge nodes firewall:
	if err := d.edgeFirewall(*svc); err != nil {
		return err
	}

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
	res.Body.Close()
	if err != nil {
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

func (d *Data) deployMasterNodes(wg *sync.WaitGroup) error {

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

	// Wait and return:
	wgInt.Wait()
	return nil
}

//-----------------------------------------------------------------------------
// func: deployWorkerNodes
//-----------------------------------------------------------------------------

func (d *Data) deployWorkerNodes(wg *sync.WaitGroup) error {

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
				"--gzip-udata",
				"--flannel-network", "10.128.0.0/21",
				"--flannel-subnet-len", "27",
				"--flannel-subnet-min", "10.128.0.192",
				"--flannel-subnet-max", "10.128.7.224",
				"--flannel-backend", "vxlan")

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

	// Wait and return:
	wgInt.Wait()
	return nil
}

//-----------------------------------------------------------------------------
// func: deployEdgeNodes
//-----------------------------------------------------------------------------

func (d *Data) deployEdgeNodes(wg *sync.WaitGroup) error {

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
				"--gzip-udata")

			// Forge the run command:
			cmdRun := exec.Command("katoctl", "run", "ec2",
				"--hostname", "edge-"+strconv.Itoa(id)+"."+d.Domain,
				"--region", d.Region,
				"--image-id", d.ImageID,
				"--instance-type", d.NodeType,
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

	// Wait and return:
	wgInt.Wait()
	return nil
}

//-----------------------------------------------------------------------------
// func: forgeNetworkInterfaces
//-----------------------------------------------------------------------------

func (d *Data) forgeNetworkInterfaces(svc ec2.EC2) []*ec2.
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

func (d *Data) runInstance(udata []byte, svc ec2.EC2) error {

	// Send the instance request:
	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:           aws.String(d.ImageID),
		MinCount:          aws.Int64(1),
		MaxCount:          aws.Int64(1),
		KeyName:           aws.String(d.KeyPair),
		InstanceType:      aws.String(d.InstanceType),
		NetworkInterfaces: d.forgeNetworkInterfaces(svc),
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

	// Wait until instance is running:
	if err := svc.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
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
	resp, err := svc.AssociateAddress(params)
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
	d.masterSecGrp = grps["master"]["SecGrpID"]
	d.nodeSecGrp = grps["node"]["SecGrpID"]
	d.edgeSecGrp = grps["edge"]["SecGrpID"]

	return nil
}

//-----------------------------------------------------------------------------
// func: masterFirewall
//-----------------------------------------------------------------------------

func (d *Data) masterFirewall(svc ec2.EC2) error {

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
	_, err := svc.AuthorizeSecurityGroupIngress(params)
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

func (d *Data) nodeFirewall(svc ec2.EC2) error {

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
	_, err := svc.AuthorizeSecurityGroupIngress(params)
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

func (d *Data) edgeFirewall(svc ec2.EC2) error {

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
	_, err := svc.AuthorizeSecurityGroupIngress(params)
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
		MasterSecGrp       string
		NodeSecGrp         string
		EdgeSecGrp         string
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
		MasterSecGrp:       d.masterSecGrp,
		NodeSecGrp:         d.nodeSecGrp,
		EdgeSecGrp:         d.edgeSecGrp,
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
