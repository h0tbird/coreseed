package ec2

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

	// Retrieve the CoreOS AMI ID:
	var err error
	if d.AmiID, err = d.retrieveCoreOSAmiID(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Execute the udata|run pipeline:
	if _, err := tools.ExecutePipeline(
		d.forgeUdataCommand(), d.forgeRunCommand()); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Publish DNS records:
	if err := publishDNSRecords("instanceID", d.Roles); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Warning(err)
	}
}

//-----------------------------------------------------------------------------
// func: retrieveCoreOSAmiID
//-----------------------------------------------------------------------------

func (d *Data) retrieveCoreOSAmiID() (string, error) {

	// Send the request:
	res, err := http.Get("https://coreos.com/dist/aws/aws-" +
		d.CoreOSChannel + ".json")
	if err != nil {
		return "", err
	}

	// Retrieve the data:
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	// Close the handler:
	if err = res.Body.Close(); err != nil {
		return "", err
	}

	// Decode JSON into Go values:
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return "", err
	}

	// Store the AMI ID:
	amis := jsonData[d.Region].(map[string]interface{})
	amiID := amis["hvm"].(string)

	// Log this action:
	log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": amiID}).
		Info("Latest CoreOS " + d.CoreOSChannel + " AMI located")

	// Return the AMI ID:
	return amiID, nil
}

//-----------------------------------------------------------------------------
// func: forgeUdataCommand
//-----------------------------------------------------------------------------

func (d *Data) forgeUdataCommand() *exec.Cmd {

	// Udata arguments bundle:
	args := []string{"udata",
		"--roles", d.Roles,
		"--cluster-id", d.ClusterID,
		"--cluster-state", d.ClusterState,
		"--quorum-count", strconv.Itoa(d.QuorumCount),
		"--master-count", strconv.Itoa(d.MasterCount),
		"--host-name", d.HostName,
		"--host-id", d.HostID,
		"--domain", d.Domain,
		"--ec2-region", d.Region,
		"--dns-provider", d.DNSProvider,
		"--dns-api-key", d.DNSApiKey,
		"--etcd-token", d.EtcdToken,
		"--calico-ip-pool", d.CalicoIPPool,
		"--rexray-storage-driver", "ebs",
		"--iaas-provider", "ec2",
		"--prometheus",
		"--gzip-udata",
	}

	// Append flags if present:
	if d.SysdigAccessKey != "" {
		args = append(args, "--sysdig-access-key", d.SysdigAccessKey)
	}
	if d.DatadogAPIKey != "" {
		args = append(args, "--datadog-api-key", d.DatadogAPIKey)
	}
	if d.SlackWebhook != "" {
		args = append(args, "--slack-webhook", d.SlackWebhook)
	}
	if d.CaCertPath != "" {
		args = append(args, "--ca-cert-path", d.CaCertPath)
	}
	for _, z := range d.StubZones {
		args = append(args, "--stub-zone", z)
	}
	if d.SMTPURL != "" {
		args = append(args, "--smtp-url", d.SMTPURL)
	}
	if d.AdminEmail != "" {
		args = append(args, "--admin-email", d.AdminEmail)
	}

	// Forge the command and return:
	return exec.Command("katoctl", args...)
}

//-----------------------------------------------------------------------------
// func: forgeRunCommand
//-----------------------------------------------------------------------------

func (d *Data) forgeRunCommand() *exec.Cmd {

	// Ec2 run arguments bundle:
	args := []string{"ec2", "run",
		"--tag-name", d.HostName + "-" + d.HostID + "." + d.Domain,
		"--region", d.Region,
		"--zone", d.Zone,
		"--ami-id", d.AmiID,
		"--instance-type", d.InstanceType,
		"--key-pair", d.KeyPair,
		"--subnet-id", d.ExtSubnetID,
		"--security-group-ids", strings.Join(d.securityGroupIDs(d.Roles), ","),
		"--iam-role", "kato",
		"--source-dest-check", "false",
		"--public-ip", "true",
	}

	// Append flags if present:
	if strings.Contains(d.Roles, "master") {
		i, _ := strconv.Atoi(d.HostID)
		args = append(args, "--private-ip", tools.OffsetIP(d.ExtSubnetCidr, 10+i))
	}
	if strings.Contains(d.Roles, "worker") {
		args = append(args, "--elb-name", d.ClusterID)
	}

	// Forge the command and return:
	return exec.Command("katoctl", args...)
}

//-----------------------------------------------------------------------------
// func: securityGroupIDs
//-----------------------------------------------------------------------------

func (d *Data) securityGroupIDs(roles string) (list []string) {
	for _, role := range strings.Split(roles, ",") {
		switch role {
		case "quorum":
			list = append(list, d.QuorumSecGrp)
		case "master":
			list = append(list, d.MasterSecGrp)
		case "worker":
			list = append(list, d.WorkerSecGrp)
		case "border":
			list = append(list, d.BorderSecGrp)
		}
	}
	return
}

//-----------------------------------------------------------------------------
// func: publishDNSRecords
//-----------------------------------------------------------------------------

func publishDNSRecords(instance, roles string) error {

	// Retrieve the instance IPs:

	// For every role in this instance:
	for _, role := range strings.Split(roles, ",") {
		fmt.Println(role)
	}

	// Success:
	return nil
}
