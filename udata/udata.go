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
	ClusterID           string
	HostName            string
	HostID              string
	Domain              string
	Role                string
	Roles               []string
	Aliases             []string
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
	Ec2Region           string
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
      volumePath: ` + os.Getenv("HOME") + `/VirtualBox Volumes
      controllerName: SATA`
	case "ec2":
		d.RexrayConfigSnippet = `aws:
      rexrayTag: kato`
	}
}

//-----------------------------------------------------------------------------
// func: forgeAliases
//-----------------------------------------------------------------------------

func (d *Data) forgeAliases() {

	// Return if exists:
	for _, i := range d.Roles {
		if i == d.HostName {
			d.Aliases = d.Roles
			return
		}
	}

	// Prepend HostName if missing:
	d.Aliases = append(strings.Fields(d.HostName), d.Roles...)
}

//-----------------------------------------------------------------------------
// func: Render
//-----------------------------------------------------------------------------

// Render takes a Data structure and outputs valid CoreOS cloud-config
// in YAML format to stdout.
func (d *Data) Render() {

	// Read the CA certificate:
	if err := d.caCert(); err != nil {
		log.WithField("cmd", "udata").Fatal(err)
	}

	d.forgeZookeeperURL() // Forge the Zookeeper URL.
	d.rexraySnippet()     // REX-Ray configuration snippet.
	d.forgeAliases()      // Forge the aliases array.

	// Role-based parsing:
	t := template.New("udata")
	var err error

	switch d.Role {
	case "master":
		t, err = t.Parse(templMaster)
	case "worker":
		t, err = t.Parse(templWorker)
	case "edge":
		t, err = t.Parse(templEdge)
	}

	if err != nil {
		log.WithField("cmd", "udata").Fatal(err)
	}

	// Apply parsed template to data object:
	if d.GzipUdata {
		log.WithFields(log.Fields{"cmd": "udata", "id": d.Role + "-" + d.HostID}).
			Info("Rendering gzipped cloud-config template")
		w := gzip.NewWriter(os.Stdout)
		if err = t.Execute(w, d); err != nil {
			_ = w.Close()
			log.WithField("cmd", "udata").Fatal(err)
		}
		_ = w.Close()
	} else {
		log.WithField("cmd", "udata").Info("Rendering plain text cloud-config template")
		if err = t.Execute(os.Stdout, d); err != nil {
			log.WithField("cmd", "udata").Fatal(err)
		}
	}
}
