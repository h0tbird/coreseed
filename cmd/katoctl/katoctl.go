package main

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"io/ioutil"
	"os"

	// Local:
	"github.com/h0tbird/kato/providers/ec2"
	"github.com/h0tbird/kato/providers/pkt"
	"github.com/h0tbird/kato/udata"

	// Community:
	log "github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

//--------------------------------------------------------------------------
// Typedefs:
//--------------------------------------------------------------------------

type cloudProvider interface {
	Deploy() error
	Setup() error
	Run(udata []byte) error
}

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
			MasterCount:         *flUdataMasterCount,
			HostID:              *flUdataHostID,
			Domain:              *flUdataDomain,
			Role:                *flUdataRole,
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
		}

		err := udata.Render()
		checkError(err)

	//--------------------
	// katoctl pkt deploy
	//--------------------

	case cmdPktDeploy.FullCommand():

		pkt := pkt.Data{}
		err := pkt.Deploy()
		checkError(err)

	//-------------------
	// katoctl pkt setup
	//-------------------

	case cmdPktSetup.FullCommand():

		pkt := pkt.Data{}
		err := pkt.Setup()
		checkError(err)

	//-----------------
	// katoctl pkt run
	//-----------------

	case cmdPktRun.FullCommand():

		pkt := pkt.Data{
			APIKey:    *flPktRunAPIKey,
			HostName:  *flPktRunHostname,
			ProjectID: *flPktRunProjectID,
			Plan:      *flPktRunPlan,
			OS:        *flPktRunOS,
			Facility:  *flPktRunFacility,
			Billing:   *flPktRunBilling,
		}

		udata, err := readUdata()
		checkError(err)
		err = pkt.Run(udata)
		checkError(err)

	//--------------------
	// katoctl ec2 deploy
	//--------------------

	case cmdEc2Deploy.FullCommand():

		ec2 := ec2.Data{
			ClusterID:        *flEc2DeployClusterID,
			MasterCount:      *flEc2DeployMasterCount,
			NodeCount:        *flEc2DeployNodeCount,
			EdgeCount:        *flEc2DeployEdgeCount,
			MasterType:       *flEc2DeployMasterType,
			NodeType:         *flEc2DeployNodeType,
			EdgeType:         *flEc2DeployEdgeType,
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
		}

		err := ec2.Deploy()
		checkError(err)

	//-------------------
	// katoctl ec2 setup
	//-------------------

	case cmdEc2Setup.FullCommand():

		ec2 := ec2.Data{
			ClusterID:     *flEc2SetupClusterID,
			Domain:        *flEc2SetupDomain,
			Region:        *flEc2SetupRegion,
			Zone:          *flEc2SetupZone,
			VpcCidrBlock:  *flEc2SetupVpcCidrBlock,
			IntSubnetCidr: *flEc2SetupIntSubnetCidr,
			ExtSubnetCidr: *flEc2SetupExtSubnetCidr,
		}

		err := ec2.Setup()
		checkError(err)

	//-----------------
	// katoctl ec2 add
	//-----------------

	case cmdEc2Add.FullCommand():

		ec2 := ec2.Data{
			ClusterID: *flEc2AddCluserID,
			Role:      *flEc2AddRole,
			ID:        *flEc2AddID,
		}

		err := ec2.Add()
		checkError(err)

	//-----------------
	// katoctl ec2 run
	//-----------------

	case cmdEc2Run.FullCommand():

		ec2 := ec2.Data{
			Region:       *flEc2RunRegion,
			Zone:         *flEc2RunZone,
			SubnetID:     *flEc2RunSubnetID,
			SecGrpID:     *flEc2RunSecGrpID,
			ImageID:      *flEc2RunImageID,
			KeyPair:      *flEc2RunKeyPair,
			InstanceType: *flEc2RunInsType,
			Hostname:     *flEc2RunHostname,
			PublicIP:     *flEc2RunPublicIP,
			IAMRole:      *flEc2RunIAMRole,
			SrcDstCheck:  *flEc2RunSrcDstCheck,
		}

		udata, err := readUdata()
		checkError(err)
		err = ec2.Run(udata)
		checkError(err)
	}
}

//--------------------------------------------------------------------------
// func: readUdata
//--------------------------------------------------------------------------

func readUdata() ([]byte, error) {

	// Read data from stdin:
	udata, err := ioutil.ReadAll(os.Stdin)
	return udata, err
}

//---------------------------------------------------------------------------
// func: checkError
//---------------------------------------------------------------------------

func checkError(err error) {
	if err != nil {
		os.Exit(1)
	}
}
