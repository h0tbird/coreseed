package main

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"os"
	"regexp"
	"strings"

	// Local:
	"github.com/h0tbird/kato/providers/cloud/ec2"
	"github.com/h0tbird/kato/providers/cloud/pkt"
	"github.com/h0tbird/kato/providers/dns/ns1"
	"github.com/h0tbird/kato/udata"

	// Community:
	log "github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

//-----------------------------------------------------------------------------
// katoctl root level command:
//-----------------------------------------------------------------------------

var app = kingpin.New("katoctl", "Katoctl defines and deploys Kato's infrastructure.")

//----------------------------------------------------------------------------
// func init() is called after all the variable declarations in the package
// have evaluated their initializers, and those are evaluated only after all
// the imported packages have been initialized:
//----------------------------------------------------------------------------

func init() {

	// Customize kingpin:
	app.Version("v0.1.0-beta").Author("Marc Villacorta Morera")
	app.UsageTemplate(usageTemplate)
	app.HelpFlag.Short('h')

	// Customize the default logger:
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetOutput(os.Stderr)
	log.SetLevel(log.InfoLevel)
}

//----------------------------------------------------------------------------
// Entry point:
//----------------------------------------------------------------------------

func main() {

	// Sub-command selector:
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	//---------------
	// katoctl udata
	//---------------

	case cmdUdata.FullCommand():

		udata := udata.Data{
			QuorumCount:         *flUdataQuorumCount,
			ClusterID:           *flUdataClusterID,
			HostName:            *flUdataHostName,
			HostID:              *flUdataHostID,
			Domain:              *flUdataDomain,
			Roles:               strings.Split(*flUdataRoles, ","),
			Ns1ApiKey:           *flUdataNs1Apikey,
			CaCert:              *flUdataCaCert,
			EtcdToken:           *flUdataEtcdToken,
			GzipUdata:           *flUdataGzipUdata,
			FlannelNetwork:      *flUdataFlannelNetwork,
			FlannelSubnetLen:    *flUdataFlannelSubnetLen,
			FlannelSubnetMin:    *flUdataFlannelSubnetMin,
			FlannelSubnetMax:    *flUdataFlannelSubnetMax,
			FlannelBackend:      *flUdataFlannelBackend,
			RexrayStorageDriver: *flUdataRexrayStorageDriver,
			RexrayEndpointIP:    *flUdataRexrayEndpointIP,
			Ec2Region:           *flUdataEc2Region,
			IaasProvider:        *flUdataIaasProvider,
		}

		udata.Render()

	//--------------------
	// katoctl pkt deploy
	//--------------------

	case cmdPktDeploy.FullCommand():

		pkt := pkt.Data{}
		pkt.Deploy()

	//-------------------
	// katoctl pkt setup
	//-------------------

	case cmdPktSetup.FullCommand():

		pkt := pkt.Data{}
		pkt.Setup()

	//-----------------
	// katoctl pkt run
	//-----------------

	case cmdPktRun.FullCommand():

		pkt := pkt.Data{
			APIKey:    *flPktRunAPIKey,
			HostName:  *flPktRunHostName,
			ProjectID: *flPktRunProjectID,
			Plan:      *flPktRunPlan,
			OS:        *flPktRunOS,
			Facility:  *flPktRunFacility,
			Billing:   *flPktRunBilling,
		}

		pkt.Run()

	//--------------------
	// katoctl ec2 deploy
	//--------------------

	case cmdEc2Deploy.FullCommand():

		ec2 := ec2.Data{
			State: ec2.State{
				ClusterID:        *flEc2DeployClusterID,
				MasterCount:      float64(*flEc2DeployMasterCount),
				WorkerCount:      float64(*flEc2DeployWorkerCount),
				BorderCount:      float64(*flEc2DeployBorderCount),
				MasterType:       *flEc2DeployMasterType,
				WorkerType:       *flEc2DeployWorkerType,
				BorderType:       *flEc2DeployBorderType,
				Channel:          *flEc2DeployChannel,
				EtcdToken:        *flEc2DeployEtcdToken,
				Ns1ApiKey:        *flEc2DeployNs1ApiKey,
				CaCert:           *flEc2DeployCaCert,
				Domain:           *flEc2DeployDomain,
				Region:           *flEc2DeployRegion,
				Zone:             *flEc2DeployZone,
				KeyPair:          *flEc2DeployKeyPair,
				VpcCidrBlock:     *flEc2DeployVpcCidrBlock,
				IntSubnetCidr:    *flEc2DeployIntSubnetCidr,
				ExtSubnetCidr:    *flEc2DeployExtSubnetCidr,
				FlannelNetwork:   *flEc2DeployFlannelNetwork,
				FlannelSubnetLen: *flEc2DeployFlannelSubnetLen,
				FlannelSubnetMin: *flEc2DeployFlannelSubnetMin,
				FlannelSubnetMax: *flEc2DeployFlannelSubnetMax,
				FlannelBackend:   *flEc2DeployFlannelBackend,
				Quadruplets:      *arEc2DeployQuadruplet,
			},
		}

		ec2.Deploy()

	//-------------------
	// katoctl ec2 setup
	//-------------------

	case cmdEc2Setup.FullCommand():

		ec2 := ec2.Data{
			State: ec2.State{
				ClusterID:     *flEc2SetupClusterID,
				Domain:        *flEc2SetupDomain,
				Region:        *flEc2SetupRegion,
				Zone:          *flEc2SetupZone,
				VpcCidrBlock:  *flEc2SetupVpcCidrBlock,
				IntSubnetCidr: *flEc2SetupIntSubnetCidr,
				ExtSubnetCidr: *flEc2SetupExtSubnetCidr,
			},
		}

		ec2.Setup()

	//-----------------
	// katoctl ec2 add
	//-----------------

	case cmdEc2Add.FullCommand():

		ec2 := ec2.Data{
			State: ec2.State{
				ClusterID: *flEc2AddCluserID,
			},
			Instance: ec2.Instance{
				Role:     *flEc2AddRole,
				Roles:    *flEc2AddRoles,
				HostName: *flEc2AddHostName,
				HostID:   *flEc2AddHostID,
				AmiID:    *flEc2AddAmiID,
			},
		}

		ec2.Add()

	//-----------------
	// katoctl ec2 run
	//-----------------

	case cmdEc2Run.FullCommand():

		ec2 := ec2.Data{
			State: ec2.State{
				Region:  *flEc2RunRegion,
				Zone:    *flEc2RunZone,
				KeyPair: *flEc2RunKeyPair,
			},
			Instance: ec2.Instance{
				SubnetID:     *flEc2RunSubnetID,
				SecGrpID:     *flEc2RunSecGrpID,
				InstanceType: *flEc2RunInsType,
				TagName:      *flEc2RunTagName,
				PublicIP:     *flEc2RunPublicIP,
				IAMRole:      *flEc2RunIAMRole,
				SrcDstCheck:  *flEc2RunSrcDstCheck,
				AmiID:        *flEc2RunAmiID,
				ELBName:      *flEc2RunELBName,
				PrivateIP:    *flEc2RunPrivateIP,
			},
		}

		ec2.Run()

	//----------------------
	// katoctl ns1 zone add
	//----------------------

	case cmdNs1ZoneAdd.FullCommand():

		ns1 := ns1.Data{
			APIKey: *flNs1APIKey,
			Link:   *flNs1ZoneAddLink,
			Zones:  *arNs1ZoneAddName,
		}

		ns1.AddZones()

	//------------------------
	// katoctl ns1 record add
	//------------------------

	case cmdNs1RecordAdd.FullCommand():

		ns1 := ns1.Data{
			APIKey:  *flNs1APIKey,
			Zone:    *flNs1RecordAddZone,
			Records: *arNs1RecordAddName,
		}

		ns1.AddRecords()
	}
}

//-----------------------------------------------------------------------------
// Regular expression custom parser:
//-----------------------------------------------------------------------------

type regexpMatchValue struct {
	value  string
	regexp string
}

func (id *regexpMatchValue) Set(value string) error {

	if match, _ := regexp.MatchString(id.regexp, value); !match {
		log.WithField("value", value).Fatal("Value must match: " + id.regexp)
	}

	id.value = value
	return nil
}

func (id *regexpMatchValue) String() string {
	return id.value
}

func regexpMatch(s kingpin.Settings, regexp string) *string {
	target := &regexpMatchValue{}
	target.regexp = regexp
	s.SetValue(target)
	return &target.value
}
