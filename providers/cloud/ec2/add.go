package ec2

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"os/exec"
	"strconv"
	"strings"

	// Community:
	log "github.com/Sirupsen/logrus"
	"github.com/katosys/kato/pkg/tools"
)

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
		"--calico-ip-pool", d.CalicoIPPool,
		"--rexray-storage-driver", "ebs",
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

	// Append the --ca-cert-path flag if present:
	if d.CaCertPath != "" {
		argsUdata = append(argsUdata, "--ca-cert-path", d.CaCertPath)
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
		"--source-dest-check", "false",
		"--public-ip", "true",
	}

	// Append the --private-ip if master:
	if strings.Contains(d.Roles, "master") {
		i, _ := strconv.Atoi(d.HostID)
		argsRun = append(argsRun, "--private-ip", tools.OffsetIP(d.ExtSubnetCidr, 10+i))
	}

	// Append the --elb-name if worker:
	if strings.Contains(d.Roles, "worker") {
		argsRun = append(argsRun, "--elb-name", d.ClusterID)
	}

	// Forge the commands:
	cmdUdata := exec.Command("katoctl", argsUdata...)
	cmdRun := exec.Command("katoctl", argsRun...)

	// Execute the pipeline:
	if err := tools.ExecutePipeline(cmdUdata, cmdRun); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}
}
