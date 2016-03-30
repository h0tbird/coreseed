package udata

//----------------------------------------------------------------------------
// Package factored import statement:
//----------------------------------------------------------------------------

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
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

		data, err := ioutil.ReadFile(d.CaCert)

		if err != nil {
			return errors.New("Unable to read the certificate file")
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
		return errors.New("Unable to parse the template")
	}

	// Apply parsed template to data object:
	err = t.Execute(os.Stdout, d)
	if err != nil {
		return errors.New("Unable to execute the template")
	}

	// Return on success:
	return nil
}
