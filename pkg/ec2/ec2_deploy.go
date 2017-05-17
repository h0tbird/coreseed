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
	"github.com/katosys/kato/pkg/kato"
)

//-----------------------------------------------------------------------------
// func: Deploy
//-----------------------------------------------------------------------------

// Deploy Kato's infrastructure on Amazon EC2.
func (d *Data) Deploy() {

	// Initializations:
	d.command = "deploy"
	wch := kato.NewWaitChan(3)

	// Count quorum and master nodes:
	d.QuorumCount = kato.CountNodes(d.Quadruplets, "quorum")
	d.MasterCount = kato.CountNodes(d.Quadruplets, "master")

	// Setup the environment (I):
	go d.setupEC2(wch)
	go kato.CreateDNSZones(wch, d.DNSProvider, d.DNSApiKey, d.Domain)
	go kato.NewEtcdToken(wch, d.QuorumCount, &d.EtcdToken)

	// Wait and check for errors:
	if err := wch.WaitErr(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Dump state to file (II):
	if err := d.dumpState(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}

	// Deploy all the nodes (III):
	for _, q := range d.Quadruplets {
		wch.WaitGrp.Add(1)
		s := strings.Split(q, ":")
		i, _ := strconv.Atoi(s[0])
		go d.deployNodes(wch, i, s[1], s[2], s[3])
	}

	// Wait for the nodes:
	if err := wch.WaitErr(); err != nil {
		log.WithField("cmd", "ec2:"+d.command).Fatal(err)
	}
}

//-----------------------------------------------------------------------------
// func: setupEC2
//-----------------------------------------------------------------------------

func (d *Data) setupEC2(wch *kato.WaitChan) {

	// Decrement:
	defer wch.WaitGrp.Done()

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
// func: deployNodes
//-----------------------------------------------------------------------------

func (d *Data) deployNodes(wch *kato.WaitChan, count int, itype, hostname, roles string) {

	// Decrement:
	defer wch.WaitGrp.Done()
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
