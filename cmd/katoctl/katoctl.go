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

	// Command parse and switch:
	command := kingpin.MustParse(cli.App.Parse(os.Args[1:]))

	switch {
	case ec2.RunCmd(command):
	case pkt.RunCmd(command):
	case ns1.RunCmd(command):
	case r53.RunCmd(command):
	}

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
