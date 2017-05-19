package pxe

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Community:
	log "github.com/Sirupsen/logrus"
	"github.com/katosys/kato/pkg/kato"
)

//-----------------------------------------------------------------------------
// func: Deploy
//-----------------------------------------------------------------------------

// Deploy Kato's infrastructure on PXE clients.
func (d *Data) Deploy() {

	// Initializations:
	d.command = "deploy"
	wch := kato.NewWaitChan(2)

	// Count quorum and master nodes:
	d.QuorumCount = kato.CountNodes(d.Quadruplets, "quorum")
	d.MasterCount = kato.CountNodes(d.Quadruplets, "master")

	// Setup the environment (I):
	go kato.CreateDNSZones(wch, d.DNSProvider, d.DNSApiKey, d.Domain)
	go kato.NewEtcdToken(wch, d.QuorumCount, &d.EtcdToken)

	// Wait and check for errors:
	if err := wch.WaitErr(); err != nil {
		log.WithField("cmd", "pxe:"+d.command).Fatal(err)
	}

	// Dump state to file (II):
	if err := kato.DumpState(d.State, d.ClusterID); err != nil {
		log.WithField("cmd", "pxe:"+d.command).Fatal(err)
	}

	// Deploy all the nodes (III):
}
