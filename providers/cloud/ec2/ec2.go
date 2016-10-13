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
	"strings"
	"sync"

	// Community:
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/imdario/mergo"
	"github.com/katosys/kato/katool"
)

//-----------------------------------------------------------------------------
// Typedefs:
//-----------------------------------------------------------------------------

// AWS API endpoints:
type svc struct {
	ec2 *ec2.EC2
	iam *iam.IAM
	elb *elb.ELB
}

// Instance data.
type Instance struct {
	AmiID        string `json:"AmiID"`        // deploy | add | run
	HostName     string `json:"HostName"`     //        | add |
	HostID       string `json:"HostID"`       //        | add |
	Roles        string `json:"Roles"`        //        | add |
	ClusterState string `json:"ClusterState"` //        | add |
	InstanceType string `json:"InstanceType"` //        | add | run
	SrcDstCheck  string `json:"SrcDstCheck"`  //        | add | run
	InstanceID   string `json:"InstanceID"`   //        |     | run
	SubnetID     string `json:"SubnetID"`     //        |     | run
	SecGrpIDs    string `json:"SecGrpIDs"`    //        |     | run
	PublicIP     string `json:"PublicIP"`     //        |     | run
	PrivateIP    string `json:"PrivateIP"`    //        |     | run
	IAMRole      string `json:"IAMRole"`      //        |     | run
	InterfaceID  string `json:"InterfaceID"`  //        |     | run
	ELBName      string `json:"ELBName"`      //        |     | run
	TagName      string `json:"TagName"`      //        |     | run
}

// State data.
type State struct {
	Quadruplets      []string `json:"-"`                // deploy |       | add |
	StubZones        []string `json:"StubZones"`        // deploy |       | add |
	QuorumCount      int      `json:"QuorumCount"`      // deploy |       | add |
	MasterCount      int      `json:"MasterCount"`      // deploy |       | add |
	Channel          string   `json:"Channel"`          // deploy |       | add |
	EtcdToken        string   `json:"EtcdToken"`        // deploy |       | add |
	Ns1ApiKey        string   `json:"Ns1ApiKey"`        // deploy |       | add |
	SysdigAccessKey  string   `json:"SysdigAccessKey:"` // deploy |       | add |
	DatadogAPIKey    string   `json:"DatadogAPIKey:"`   // deploy |       | add |
	SlackWebhook     string   `json:"SlackWebhook:"`    // deploy |       | add |
	SMTPURL          string   `json:"SMTPURL:"`         // deploy |       | add |
	AdminEmail       string   `json:"AdminEmail:"`      // deploy |       | add |
	CaCert           string   `json:"CaCert"`           // deploy |       | add |
	FlannelNetwork   string   `json:"FlannelNetwork"`   // deploy |       | add |
	FlannelSubnetLen string   `json:"FlannelSubnetLen"` // deploy |       | add |
	FlannelSubnetMin string   `json:"FlannelSubnetMin"` // deploy |       | add |
	FlannelSubnetMax string   `json:"FlannelSubnetMax"` // deploy |       | add |
	FlannelBackend   string   `json:"FlannelBackend"`   // deploy |       | add |
	Domain           string   `json:"Domain"`           // deploy | setup | add |
	ClusterID        string   `json:"ClusterID"`        // deploy | setup | add |
	Region           string   `json:"Region"`           // deploy | setup | add | run
	Zone             string   `json:"Zone"`             // deploy | setup | add | run
	VpcCidrBlock     string   `json:"VpcCidrBlock"`     // deploy | setup |     |
	IntSubnetCidr    string   `json:"IntSubnetCidr"`    // deploy | setup |     |
	ExtSubnetCidr    string   `json:"ExtSubnetCidr"`    // deploy | setup |     |
	AllocationID     string   `json:"AllocationID"`     //        | setup |     | run
	VpcID            string   `json:"VpcID"`            //        | setup |     |
	MainRouteTableID string   `json:"MainRouteTableID"` //        | setup |     |
	InetGatewayID    string   `json:"InetGatewayID"`    //        | setup |     |
	NatGatewayID     string   `json:"NatGatewayID"`     //        | setup |     |
	RouteTableID     string   `json:"RouteTableID"`     //        | setup |     |
	KatoRoleID       string   `json:"KatoRoleID"`       //        | setup |     |
	RexrayPolicy     string   `json:"RexrayPolicy"`     //        | setup |     |
	QuorumSecGrp     string   `json:"QuorumSecGrp"`     //        | setup |     |
	MasterSecGrp     string   `json:"MasterSecGrp"`     //        | setup |     |
	WorkerSecGrp     string   `json:"WorkerSecGrp"`     //        | setup |     |
	BorderSecGrp     string   `json:"BorderSecGrp"`     //        | setup |     |
	ELBSecGrp        string   `json:"ELBSecGrp"`        //        | setup |     |
	IntSubnetID      string   `json:"IntSubnetID"`      //        | setup |     |
	ExtSubnetID      string   `json:"ExtSubnetID"`      //        | setup |     |
	DNSName          string   `json:"DNSName"`          //        | setup |     |
	KeyPair          string   `json:"KeyPair"`          //        |       | add | run
}

// Data struct for EC2 endpoints, instance and state data.
type Data struct {
	command string
	svc
	Instance
	State
}

//-----------------------------------------------------------------------------
// func: Deploy
//-----------------------------------------------------------------------------

// Deploy Kato's infrastructure on Amazon EC2.
func (d *Data) Deploy() {

	// Initializations:
	d.command = "deploy"
	var wg sync.WaitGroup
	d.countNodes()

	// Setup the environment (I):
	wg.Add(4)
	go d.setupEC2(&wg)
	go d.createDNSZones(&wg)
	go d.retrieveEtcdToken(&wg)
	go d.retrieveCoreosAmiID(&wg)
	wg.Wait()

	// Dump state to file (II):
	if err := d.dumpState(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Deploy all the nodes (III):
	for _, q := range d.Quadruplets {
		wg.Add(1)
		s := strings.Split(q, ":")
		i, _ := strconv.Atoi(s[0])
		go d.deployNodes(i, s[1], s[2], s[3], &wg)
	}

	// Wait for the nodes:
	wg.Wait()
}

//-----------------------------------------------------------------------------
// func: Add
//-----------------------------------------------------------------------------

// Add a new instance to the cluster.
func (d *Data) Add() {

	// Set current command:
	d.command = "add"

	// Load state from state file:
	if err := d.loadState(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Discover CoreOS AMI (for standalone runs):
	if d.AmiID == "" {
		d.retrieveCoreosAmiID(nil)
	}

	// Whether or not to source-dest-check:
	if d.FlannelBackend == "host-gw" {
		d.SrcDstCheck = "false"
	} else {
		d.SrcDstCheck = "true"
	}

	// Security group IDs:
	var securityGroupIDs []string
	for _, role := range strings.Split(d.Roles, ",") {
		switch role {
		case "quorum":
			securityGroupIDs = append(securityGroupIDs, d.QuorumSecGrp)
		case "master":
			securityGroupIDs = append(securityGroupIDs, d.MasterSecGrp)
		case "worker":
			securityGroupIDs = append(securityGroupIDs, d.WorkerSecGrp)
		case "border":
			securityGroupIDs = append(securityGroupIDs, d.BorderSecGrp)
		}
	}

	// Udata arguments bundle:
	argsUdata := []string{"udata",
		"--roles", d.Roles,
		"--cluster-id", d.ClusterID,
		"--cluster-state", d.ClusterState,
		"--quorum-count", strconv.Itoa(d.QuorumCount),
		"--master-count", strconv.Itoa(d.MasterCount),
		"--host-name", d.HostName,
		"--host-id", d.HostID,
		"--domain", d.Domain,
		"--ec2-region", d.Region,
		"--ns1-api-key", d.Ns1ApiKey,
		"--etcd-token", d.EtcdToken,
		"--flannel-network", d.FlannelNetwork,
		"--flannel-subnet-len", d.FlannelSubnetLen,
		"--flannel-subnet-min", d.FlannelSubnetMin,
		"--flannel-subnet-max", d.FlannelSubnetMax,
		"--flannel-backend", d.FlannelBackend,
		"--rexray-storage-driver", "ec2",
		"--iaas-provider", "ec2",
		"--prometheus",
		"--gzip-udata",
	}

	// Append the --sysdig-access-key if present:
	if d.SysdigAccessKey != "" {
		argsUdata = append(argsUdata, "--sysdig-access-key", d.SysdigAccessKey)
	}

	// Append the --datadog-api-key if present:
	if d.DatadogAPIKey != "" {
		argsUdata = append(argsUdata, "--datadog-api-key", d.DatadogAPIKey)
	}

	// Append the --slack-webhook if present:
	if d.SlackWebhook != "" {
		argsUdata = append(argsUdata, "--slack-webhook", d.SlackWebhook)
	}

	// Append the --ca-cert flag if cert is present:
	if d.CaCert != "" {
		argsUdata = append(argsUdata, "--ca-cert", d.CaCert)
	}

	// Append --stub-zone flags if present:
	for _, z := range d.StubZones {
		argsUdata = append(argsUdata, "--stub-zone", z)
	}

	// Append --smtp-url flags if present:
	if d.SMTPURL != "" {
		argsUdata = append(argsUdata, "--smtp-url", d.SMTPURL)
	}

	// Append --admin-email flags if present:
	if d.AdminEmail != "" {
		argsUdata = append(argsUdata, "--admin-email", d.AdminEmail)
	}

	// Ec2 run arguments bundle:
	argsRun := []string{"ec2", "run",
		"--tag-name", d.HostName + "-" + d.HostID + "." + d.Domain,
		"--region", d.Region,
		"--zone", d.Zone,
		"--ami-id", d.AmiID,
		"--instance-type", d.InstanceType,
		"--key-pair", d.KeyPair,
		"--subnet-id", d.ExtSubnetID,
		"--security-group-ids", strings.Join(securityGroupIDs, ","),
		"--iam-role", "kato",
		"--source-dest-check", d.SrcDstCheck,
		"--public-ip", "true",
	}

	// Append the --private-ip if master:
	if strings.Contains(d.Roles, "master") {
		i, _ := strconv.Atoi(d.HostID)
		argsRun = append(argsRun, "--private-ip", katool.OffsetIP(d.ExtSubnetCidr, 10+i))
	}

	// Append the --elb-name if worker:
	if strings.Contains(d.Roles, "worker") {
		argsRun = append(argsRun, "--elb-name", d.ClusterID)
	}

	// Forge the commands:
	cmdUdata := exec.Command("katoctl", argsUdata...)
	cmdRun := exec.Command("katoctl", argsRun...)

	// Execute the pipeline:
	if err := katool.ExecutePipeline(cmdUdata, cmdRun); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}
}

//-----------------------------------------------------------------------------
// func: Run
//-----------------------------------------------------------------------------

// Run uses EC2 API to launch a new instance.
func (d *Data) Run() {

	// Set current command:
	d.command = "run"

	// Read udata from stdin:
	udata, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Connect and authenticate to the API endpoints:
	d.ec2 = ec2.New(session.New(&aws.Config{Region: aws.String(d.Region)}))
	d.elb = elb.New(session.New(&aws.Config{Region: aws.String(d.Region)}))

	// Run the EC2 instance:
	if err := d.runInstance(udata); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Modify instance attributes:
	if err := d.modifyInstanceAttribute(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Setup an elastic IP:
	if d.PublicIP == "elastic" {

		// Allocate an elastic IP address:
		if err := d.allocateElasticIP(); err != nil {
			log.WithField("cmd", "ec2:"+d.command).Fatal(err)
		}

		// Associate the elastic IP:
		if err := d.associateElasticIP(); err != nil {
			log.WithField("cmd", "ec2:"+d.command).Fatal(err)
		}
	}

	// Register with ELB:
	if d.ELBName != "" {
		if err := d.registerWithELB(); err != nil {
			log.WithField("cmd", "ec2:"+d.command).Fatal(err)
		}
	}
}

//-----------------------------------------------------------------------------
// func: Setup
//-----------------------------------------------------------------------------

// Setup VPC, IAM and EC2 components.
func (d *Data) Setup() {

	// Set current command:
	d.command = "setup"

	// Log this acction:
	log.WithField("cmd", "ec2:"+d.command).
		Info("Connecting to region " + d.Region)

	// Connect and authenticate to the API endpoints:
	d.ec2 = ec2.New(session.New(&aws.Config{Region: aws.String(d.Region)}))
	d.iam = iam.New(session.New(&aws.Config{Region: aws.String(d.Region)}))
	d.elb = elb.New(session.New(&aws.Config{Region: aws.String(d.Region)}))

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
// func: createDNSZones
//-----------------------------------------------------------------------------

func (d *Data) createDNSZones(wg *sync.WaitGroup) {

	// Decrement:
	defer wg.Done()

	// Return if no API key is provided:
	if d.Ns1ApiKey == "" {
		return
	}

	// Forge the zone command:
	cmdZoneSetup := exec.Command("katoctl", "ns1",
		"--api-key", d.Ns1ApiKey,
		"zone", "add",
		"int."+d.Domain,
		"ext."+d.Domain)

	// Execute the zone command:
	cmdZoneSetup.Stderr = os.Stderr
	if err := cmdZoneSetup.Run(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Forge the linked zone command:
	cmdLinkedZoneSetup := exec.Command("katoctl", "ns1",
		"--api-key", d.Ns1ApiKey,
		"zone", "add",
		"--link", "int."+d.Domain,
		d.Domain)

	// Execute the linked zone command:
	cmdLinkedZoneSetup.Stderr = os.Stderr
	if err := cmdLinkedZoneSetup.Run(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}
}

//-----------------------------------------------------------------------------
// func: setupEC2
//-----------------------------------------------------------------------------

func (d *Data) setupEC2(wg *sync.WaitGroup) {

	// Decrement:
	defer wg.Done()

	// Log this action:
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.Domain}).
		Info("Setup the EC2 environment")

	// Forge the setup command:
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
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Merge state from state file:
	if err := d.loadState(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}
}

//-----------------------------------------------------------------------------
// func: countNodes
//-----------------------------------------------------------------------------

func (d *Data) countNodes() {

	// Get the QuorumCount:
	for _, q := range d.Quadruplets {
		if strings.Contains(q, "quorum") {
			s := strings.Split(q, ":")
			d.QuorumCount, _ = strconv.Atoi(s[0])
			break
		}
	}

	// Get the MasterCount:
	for _, q := range d.Quadruplets {
		if strings.Contains(q, "master") {
			s := strings.Split(q, ":")
			d.MasterCount, _ = strconv.Atoi(s[0])
			break
		}
	}
}

//-----------------------------------------------------------------------------
// func: retrieveEtcdToken
//-----------------------------------------------------------------------------

func (d *Data) retrieveEtcdToken(wg *sync.WaitGroup) {

	// Decrement:
	defer wg.Done()
	var err error

	// Request the token:
	if d.EtcdToken == "auto" {
		if d.EtcdToken, err = katool.EtcdToken(d.QuorumCount); err != nil {
			log.WithField("cmd", "ec2:"+d.command).Fatal(err)
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
	if wg != nil {
		defer wg.Done()
	}

	// Send the request:
	res, err := http.
		Get("https://coreos.com/dist/aws/aws-" + d.Channel + ".json")
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Retrieve the data:
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Close the handler:
	if err = res.Body.Close(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Decode JSON into Go values:
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Store the AMI ID:
	amis := jsonData[d.Region].(map[string]interface{})
	d.AmiID = amis["hvm"].(string)

	// Log this action:
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.AmiID}).
		Info("Latest CoreOS " + d.Channel + " AMI located")
}

//-----------------------------------------------------------------------------
// func: deployNodes
//-----------------------------------------------------------------------------

func (d *Data) deployNodes(count int, itype, hostname, roles string, wg *sync.WaitGroup) {

	// Decrement:
	defer wg.Done()
	var wgInt sync.WaitGroup

	log.WithField("cmd", "ec2:"+d.command).
		Info("Deploying " + strconv.Itoa(count) + " " + hostname + " nodes")

	for i := 1; i <= count; i++ {
		wgInt.Add(1)

		go func(id int) {
			defer wgInt.Done()

			// Forge the add command:
			cmdAdd := exec.Command("katoctl", "ec2", "add",
				"--cluster-id", d.ClusterID,
				"--cluster-state", "new",
				"--roles", roles,
				"--host-name", hostname,
				"--host-id", strconv.Itoa(id),
				"--ami-id", d.AmiID,
				"--instance-type", itype)

			// Execute the add command:
			cmdAdd.Stderr = os.Stderr
			if err := cmdAdd.Run(); err != nil {
				log.WithField("cmd", "ec2:"+d.command).Fatal(err)
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

	// Append to security group array:
	for _, grp := range strings.Split(d.SecGrpIDs, ",") {
		securityGroupIds = append(securityGroupIds, aws.String(grp))
	}

	// Forge the interface data type:
	iface := ec2.InstanceNetworkInterfaceSpecification{
		DeleteOnTermination: aws.Bool(true),
		DeviceIndex:         aws.Int64(int64(0)),
		Groups:              securityGroupIds,
		SubnetId:            aws.String(d.SubnetID),
	}

	// Private IP address:
	if d.PrivateIP != "" {
		iface.PrivateIpAddress = aws.String(d.PrivateIP)
	}

	// Public IP address:
	if d.PublicIP == "true" {
		iface.AssociatePublicIpAddress = aws.Bool(true)
	}

	// Append to the interfaces array:
	networkInterfaces = append(networkInterfaces, &iface)

	return networkInterfaces
}

//-----------------------------------------------------------------------------
// func: runInstance
//-----------------------------------------------------------------------------

func (d *Data) runInstance(udata []byte) error {

	// Forge the instance request:
	params := &ec2.RunInstancesInput{
		ImageId:           aws.String(d.AmiID),
		MinCount:          aws.Int64(1),
		MaxCount:          aws.Int64(1),
		KeyName:           aws.String(d.KeyPair),
		InstanceType:      aws.String(d.InstanceType),
		NetworkInterfaces: d.forgeNetworkInterfaces(),
		Placement: &ec2.Placement{
			AvailabilityZone: aws.String(d.Region + d.Zone),
		},
		UserData: aws.String(base64.StdEncoding.EncodeToString([]byte(udata))),
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Name: aws.String(d.IAMRole),
		},
	}

	// Send the instance request:
	resp, err := d.ec2.RunInstances(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Store the instance ID:
	d.InstanceID = *resp.Instances[0].InstanceId
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.InstanceID}).
		Info("New " + d.InstanceType + " EC2 instance requested")

	// Store the interface ID:
	d.InterfaceID = *resp.Instances[0].
		NetworkInterfaces[0].NetworkInterfaceId

	// Tag the instance:
	if err := d.tag(d.InstanceID, "Name", d.TagName); err != nil {
		return err
	}

	// Pretty-print to stderr:
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.TagName}).
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
		InstanceId: aws.String(d.InstanceID),
		SourceDestCheck: &ec2.AttributeBooleanValue{
			Value: aws.Bool(SrcDstCheck),
		},
	}

	// Send the attribute modification request:
	_, err = d.ec2.ModifyInstanceAttribute(params)
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

	// Create a NAT gateway:
	if err := d.createNatGateway(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Create a default route via NAT GW (int):
	if err := d.createNatGatewayRoute(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
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
// func: createVPC
//-----------------------------------------------------------------------------

func (d *Data) createVPC() error {

	// Forge the VPC request:
	params := &ec2.CreateVpcInput{
		CidrBlock:       aws.String(d.VpcCidrBlock),
		DryRun:          aws.Bool(false),
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

	// Tag the VPC:
	if err = d.tag(d.VpcID, "Name", d.Domain); err != nil {
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
	d.MainRouteTableID = *resp.RouteTables[0].RouteTableId
	log.WithFields(log.Fields{
		"cmd": "ec2:" + d.command, "id": d.MainRouteTableID}).
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
			VpcId:            aws.String(d.VpcID),
			AvailabilityZone: aws.String(d.Region + d.Zone),
			DryRun:           aws.Bool(false),
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
		VpcId:  aws.String(d.VpcID),
		DryRun: aws.Bool(false),
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
		DryRun:       aws.Bool(false),
	}

	// Send the association request:
	resp, err := d.ec2.AssociateRouteTable(params)
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
		DryRun:            aws.Bool(false),
	}

	// Send the attachement request:
	if _, err := d.ec2.AttachInternetGateway(params); err != nil {
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
		RouteTableId:         aws.String(d.RouteTableID),
		DryRun:               aws.Bool(false),
		GatewayId:            aws.String(d.InetGatewayID),
	}

	// Send the route request:
	if _, err := d.ec2.CreateRoute(params); err != nil {
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
// func: associateElasticIP
//-----------------------------------------------------------------------------

func (d *Data) associateElasticIP() error {

	// Wait until instance is running:
	if err := d.ec2.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(d.InstanceID)},
	}); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Forge the association request:
	params := &ec2.AssociateAddressInput{
		AllocationId:       aws.String(d.AllocationID),
		AllowReassociation: aws.Bool(true),
		DryRun:             aws.Bool(false),
		NetworkInterfaceId: aws.String(d.InterfaceID),
	}

	// Send the association request:
	resp, err := d.ec2.AssociateAddress(params)
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
// func: registerWithELB
//-----------------------------------------------------------------------------

func (d *Data) registerWithELB() error {

	// Forge the register request:
	params := &elb.RegisterInstancesWithLoadBalancerInput{
		Instances: []*elb.Instance{
			{
				InstanceId: aws.String(d.InstanceID),
			},
		},
		LoadBalancerName: aws.String(d.ELBName),
	}

	// Send the register request:
	_, err := d.elb.RegisterInstancesWithLoadBalancer(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Log this action:
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.ELBName}).
		Info("Instance registered with ELB")

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
		DryRun:               aws.Bool(false),
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
// func: createSecurityGroup
//-----------------------------------------------------------------------------

func (d *Data) createSecurityGroup(name string, id *string) error {

	// Forge the group request:
	params := &ec2.CreateSecurityGroupInput{
		Description: aws.String(d.Domain + " " + name),
		GroupName:   aws.String(name),
		DryRun:      aws.Bool(false),
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
	_, err := d.ec2.AuthorizeSecurityGroupIngress(params)
	if err != nil {
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
	_, err := d.ec2.AuthorizeSecurityGroupIngress(params)
	if err != nil {
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
	_, err := d.ec2.AuthorizeSecurityGroupIngress(params)
	if err != nil {
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
	_, err := d.ec2.AuthorizeSecurityGroupIngress(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": "border"}).
		Info("New firewall rules defined")

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
	_, err := d.ec2.AuthorizeSecurityGroupIngress(params)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": "elb"}).
		Info("New firewall rules defined")

	return nil
}

//-----------------------------------------------------------------------------
// func: dumpState
//-----------------------------------------------------------------------------

func (d *Data) dumpState() error {

	// Marshal the data:
	data, err := json.MarshalIndent(d.State, "", "  ")
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
	err = ioutil.WriteFile(path+"/"+d.ClusterID+".json", data, 0600)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: loadState
//-----------------------------------------------------------------------------

func (d *Data) loadState() error {

	// Load raw data from state file:
	raw, err := katool.LoadState(d.ClusterID)
	if err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Decode the loaded JSON data:
	dat := State{}
	if err := json.Unmarshal(raw, &dat); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	// Merge the decoded data into the current state:
	if err := mergo.Map(&d.State, dat); err != nil {
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
	if _, err := d.ec2.CreateTags(params); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Error(err)
		return err
	}

	return nil
}
