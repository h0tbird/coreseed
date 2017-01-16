package ec2

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"time"

	// Community:
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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
	SrcDstCheck  string `json:"SrcDstCheck"`  //        |     | run
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
	CoreOSChannel    string   `json:"CoreOSChannel"`    // deploy |       | add |
	EtcdToken        string   `json:"EtcdToken"`        // deploy |       | add |
	Ns1ApiKey        string   `json:"Ns1ApiKey"`        // deploy |       | add |
	SysdigAccessKey  string   `json:"SysdigAccessKey:"` // deploy |       | add |
	DatadogAPIKey    string   `json:"DatadogAPIKey:"`   // deploy |       | add |
	SlackWebhook     string   `json:"SlackWebhook:"`    // deploy |       | add |
	SMTPURL          string   `json:"SMTPURL:"`         // deploy |       | add |
	AdminEmail       string   `json:"AdminEmail:"`      // deploy |       | add |
	CaCertPath       string   `json:"CaCertPath"`       // deploy |       | add |
	CalicoIPPool     string   `json:"CalicoIPPool"`     // deploy |       |     |
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
	}

	// Send the tag request:
	for i := 0; i < 5; i++ {
		if _, err := d.ec2.CreateTags(params); err != nil {
			ec2err, ok := err.(awserr.Error)
			if ok && strings.Contains(ec2err.Code(), ".NotFound") {
				time.Sleep(1e9)
				continue
			}
			log.WithField("cmd", "ec2:"+d.command).Error(err)
			return err
		}
	}

	return nil
}
