package katoctl

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"os"
	"path"
	"runtime"
	"strings"

	// Local:
	"github.com/katosys/kato/pkg/cli"
	"github.com/katosys/kato/pkg/ec2"
	"github.com/katosys/kato/pkg/ns1"
	"github.com/katosys/kato/pkg/pkt"
	"github.com/katosys/kato/pkg/r53"
	"github.com/katosys/kato/pkg/udata"

	// Community:
	log "github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

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
	log.AddHook(contextHook{})
}

//----------------------------------------------------------------------------
// Entry point:
//----------------------------------------------------------------------------

func main() {

	// Sub-command selector:
	command := kingpin.MustParse(cli.App.Parse(os.Args[1:]))

	// New way:
	switch {
	case pkt.RunCmd(command):
	case ns1.RunCmd(command):
	case r53.RunCmd(command):
	}

	// Old way:
	switch command {

	//---------------
	// katoctl udata
	//---------------

	case cmdUdata.FullCommand():

		udata := udata.CmdData{
			CmdFlags: udata.CmdFlags{
				AdminEmail:          *flUdataAdminEmail,
				CaCertPath:          *flUdataCaCertPath,
				CalicoIPPool:        *flUdataCalicoIPPool,
				ClusterID:           *flUdataClusterID,
				ClusterState:        *flUdataClusterState,
				DatadogAPIKey:       *flUdataDatadogAPIKey,
				Domain:              *flUdataDomain,
				Ec2Region:           *flUdataEc2Region,
				EtcdToken:           *flUdataEtcdToken,
				GzipUdata:           *flUdataGzipUdata,
				HostID:              *flUdataHostID,
				HostName:            *flUdataHostName,
				IaasProvider:        *flUdataIaasProvider,
				MasterCount:         *flUdataMasterCount,
				Ns1ApiKey:           *flUdataNs1Apikey,
				Prometheus:          *flUdataPrometheus,
				QuorumCount:         *flUdataQuorumCount,
				RexrayEndpointIP:    *flUdataRexrayEndpointIP,
				RexrayStorageDriver: *flUdataRexrayStorageDriver,
				Roles:               strings.Split(*flUdataRoles, ","),
				SlackWebhook:        *flUdataSlackWebhook,
				SMTPURL:             *flUdataSMTPURL,
				StubZones:           *flUdataStubZones,
				SysdigAccessKey:     *flUdataSysdigAccessKey,
			},
		}

		udata.CmdRun()

	//--------------------
	// katoctl ec2 deploy
	//--------------------

	case cmdEc2Deploy.FullCommand():

		ec2 := ec2.Data{
			State: ec2.State{
				ClusterID:       *flEc2DeployClusterID,
				CoreOSChannel:   *flEc2DeployCoreOSChannel,
				KeyPair:         *flEc2DeployKeyPair,
				EtcdToken:       *flEc2DeployEtcdToken,
				Ns1ApiKey:       *flEc2DeployNs1ApiKey,
				SysdigAccessKey: *flEc2DeploySysdigAccessKey,
				DatadogAPIKey:   *flEc2DeployDatadogAPIKey,
				CaCertPath:      *flEc2DeployCaCertPath,
				Domain:          *flEc2DeployDomain,
				Region:          *flEc2DeployRegion,
				Zone:            *flEc2DeployZone,
				VpcCidrBlock:    *flEc2DeployVpcCidrBlock,
				CalicoIPPool:    *flEc2DeployCalicoIPPool,
				IntSubnetCidr:   *flEc2DeployIntSubnetCidr,
				ExtSubnetCidr:   *flEc2DeployExtSubnetCidr,
				StubZones:       *flEc2DeployStubZones,
				SlackWebhook:    *flEc2DeploySlackWebhook,
				SMTPURL:         *flEc2DeploySMTPURL,
				AdminEmail:      *flEc2DeployAdminEmail,
				Quadruplets:     *arEc2DeployQuadruplet,
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
				Roles:        *flEc2AddRoles,
				HostName:     *flEc2AddHostName,
				HostID:       *flEc2AddHostID,
				AmiID:        *flEc2AddAmiID,
				InstanceType: *flEc2AddInsanceType,
				ClusterState: *flEc2AddClusterState,
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
				SecGrpIDs:    *flEc2RunSecGrpIDs,
				InstanceType: *flEc2RunInstanceType,
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
	}
}

//-----------------------------------------------------------------------------
// Log filename and line number:
//-----------------------------------------------------------------------------

type contextHook struct{}

func (hook contextHook) Levels() []log.Level {
	levels := []log.Level{log.ErrorLevel, log.FatalLevel}
	return levels
}

func (hook contextHook) Fire(entry *log.Entry) error {
	pc := make([]uintptr, 3, 3)
	cnt := runtime.Callers(6, pc)

	for i := 0; i < cnt; i++ {
		fu := runtime.FuncForPC(pc[i] - 1)
		name := fu.Name()
		if !strings.Contains(name, "github.com/Sirupsen/logrus") {
			file, line := fu.FileLine(pc[i] - 1)
			entry.Data["file"] = path.Base(file)
			entry.Data["func"] = path.Base(name)
			entry.Data["line"] = line
			break
		}
	}
	return nil
}
