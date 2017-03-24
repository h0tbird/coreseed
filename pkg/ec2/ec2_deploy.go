package ec2

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	// Community:
	log "github.com/Sirupsen/logrus"
	"github.com/katosys/kato/pkg/tools"
)

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
// func: retrieveEtcdToken
//-----------------------------------------------------------------------------

func (d *Data) retrieveEtcdToken(wg *sync.WaitGroup) {

	// Decrement:
	defer wg.Done()
	var err error

	// Request the token:
	if d.EtcdToken == "auto" {
		if d.EtcdToken, err = tools.EtcdToken(d.QuorumCount); err != nil {
			log.WithField("cmd", "ec2:"+d.command).Fatal(err)
		}
		log.WithFields(log.Fields{"cmd": "ec2:" + d.command, "id": d.EtcdToken}).
			Info("New etcd bootstrap token requested")
	}
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
