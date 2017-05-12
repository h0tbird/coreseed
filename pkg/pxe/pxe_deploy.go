package pxe

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"sync"

	// Community:
	log "github.com/Sirupsen/logrus"
)

//-----------------------------------------------------------------------------
// func: Deploy
//-----------------------------------------------------------------------------

// Deploy Kato's infrastructure on PXE clients.
func (d *Data) Deploy() {

	// Initializations:
	d.command = "deploy"
	var wg sync.WaitGroup

	// Setup the environment (I):
	wg.Add(3)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		log.WithField("cmd", "pxe:"+d.command).Info("PXE deploy 1")
	}(&wg)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		log.WithField("cmd", "pxe:"+d.command).Info("PXE deploy 2")
	}(&wg)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		log.WithField("cmd", "pxe:"+d.command).Info("PXE deploy 3")
	}(&wg)
	wg.Wait()

	// Dump state to file (II):

	// Deploy all the nodes (III):
}
