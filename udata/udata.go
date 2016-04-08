package udata

//----------------------------------------------------------------------------
// Package factored import statement:
//----------------------------------------------------------------------------

import (

	// Stdlib:
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	// Community:
	log "github.com/Sirupsen/logrus"
)

//----------------------------------------------------------------------------
// Typedefs:
//----------------------------------------------------------------------------

// Data contains variables to be interpolated in templates.
type Data struct {
	HostID           string
	Domain           string
	Role             string
	Ns1ApiKey        string
	CaCert           string
	EtcdToken        string
	FlannelNetwork   string
	FlannelSubnetLen string
	FlannelSubnetMin string
	FlannelSubnetMax string
	FlannelBackend   string
}

//--------------------------------------------------------------------------
// func: Render
//--------------------------------------------------------------------------

// Render takes a Data structure and outputs valid CoreOS cloud-config
// in YAML format to stdout.
func (d *Data) Render() error {

	// Read the CA certificate:
	if d.CaCert != "" {

		log.WithField("cmd", "udata").Info("- Reading CA certificate.")
		data, err := ioutil.ReadFile(d.CaCert)
		if err != nil {
			log.WithField("cmd", "udata").Error(err)
			return err
		}

		d.CaCert = strings.TrimSpace(strings.
			Replace(string(data), "\n", "\n    ", -1))
	}

	// Udata template:
	var err error
	t := template.New("udata")

	// Role based parse:
	switch d.Role {
	case "master":
		t, err = t.Parse(templMaster)
	case "node":
		t, err = t.Parse(templNode)
	case "edge":
		t, err = t.Parse(templEdge)
	}

	if err != nil {
		log.WithField("cmd", "udata").Error(err)
		return err
	}

	// Apply parsed template to data object:
	log.WithField("cmd", "udata").Info("- Rendering cloud-config template.")
	if err = t.Execute(os.Stdout, d); err != nil {
		log.WithField("cmd", "udata").Error(err)
		return err
	}

	// Return on success:
	return nil
}
