package udata

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"compress/gzip"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/template"

	// Community:
	log "github.com/Sirupsen/logrus"
)

//-----------------------------------------------------------------------------
// Typedefs:
//-----------------------------------------------------------------------------

// Data contains variables to be interpolated in templates.
type Data struct {
	MasterCount         int
	HostID              string
	Domain              string
	Role                string
	Ns1ApiKey           string
	CaCert              string
	EtcdToken           string
	ZkServers           string
	GzipUdata           bool
	FlannelNetwork      string
	FlannelSubnetLen    string
	FlannelSubnetMin    string
	FlannelSubnetMax    string
	FlannelBackend      string
	RexrayStorageDriver string
	RexrayConfigSnippet string
	RexrayEndpointIP    string
}

//-----------------------------------------------------------------------------
// func: caCert
//-----------------------------------------------------------------------------

func (d *Data) caCert() error {

	if d.CaCert != "" {

		data, err := ioutil.ReadFile(d.CaCert)
		if err != nil {
			log.WithField("cmd", "udata").Error(err)
			return err
		}

		d.CaCert = strings.TrimSpace(strings.
			Replace(string(data), "\n", "\n    ", -1))
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: forgeZookeeperURL
//-----------------------------------------------------------------------------

func (d *Data) forgeZookeeperURL() {

	for i := 1; i <= d.MasterCount; i++ {
		d.ZkServers = d.ZkServers + "master-" + strconv.Itoa(i) + ":2181"
		if i != d.MasterCount {
			d.ZkServers = d.ZkServers + ","
		}
	}
}

//-----------------------------------------------------------------------------
// func: rexraySnippet
//-----------------------------------------------------------------------------

func (d *Data) rexraySnippet() {

	switch d.RexrayStorageDriver {

	case "virtualbox":
		d.RexrayConfigSnippet = `virtualbox:
      endpoint: http://` + d.RexrayEndpointIP + `:18083
      volumePath: ` + os.Getenv("HOME") + `/VirtualBox Volumes`
	case "ec2":
		d.RexrayConfigSnippet = `ec2`
	}
}

//-----------------------------------------------------------------------------
// func: Render
//-----------------------------------------------------------------------------

// Render takes a Data structure and outputs valid CoreOS cloud-config
// in YAML format to stdout.
func (d *Data) Render() error {

	var err error

	// Read the CA certificate:
	if err = d.caCert(); err != nil {
		return err
	}

	// Forge the Zookeeper URL:
	d.forgeZookeeperURL()

	// REX-Ray configuration snippet:
	d.rexraySnippet()

	// Role-based parsing:
	t := template.New("udata")

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
	if d.GzipUdata {
		log.WithFields(log.Fields{"cmd": "udata", "id": d.Role + "-" + d.HostID}).
			Info("- Rendering gzipped cloud-config template")
		w := gzip.NewWriter(os.Stdout)
		defer w.Close()
		if err = t.Execute(w, d); err != nil {
			log.WithField("cmd", "udata").Error(err)
			return err
		}
	} else {
		log.WithField("cmd", "udata").Info("- Rendering plain text cloud-config template")
		if err = t.Execute(os.Stdout, d); err != nil {
			log.WithField("cmd", "udata").Error(err)
			return err
		}
	}

	// Return on success:
	return nil
}
