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

type filter struct {
	anyOf  []string
	noneOf []string
	allOf  []string
}

type fragment struct {
	filter
	data string
}

// Data contains variables to be interpolated in templates.
type Data struct {
	QuorumCount         int
	GzipUdata           bool
	ClusterID           string
	HostName            string
	HostID              string
	Domain              string
	Role                string
	Ns1ApiKey           string
	CaCert              string
	EtcdToken           string
	ZkServers           string
	FlannelNetwork      string
	FlannelSubnetLen    string
	FlannelSubnetMin    string
	FlannelSubnetMax    string
	FlannelBackend      string
	RexrayStorageDriver string
	RexrayConfigSnippet string
	RexrayEndpointIP    string
	Ec2Region           string
	IaasProvider        string
	template            string
	Roles               []string
	Aliases             []string
	SystemdUnits        []string
	frags               []fragment
}

//-----------------------------------------------------------------------------
// func: anyOf
//-----------------------------------------------------------------------------

func (f *fragment) anyOf(tags []string) bool {
	for _, tag := range tags {
		for _, filter := range f.filter.anyOf {
			if tag == filter {
				return true
			}
		}
	}
	return false
}

//-----------------------------------------------------------------------------
// func: noneOf
//-----------------------------------------------------------------------------

func (f *fragment) noneOf(tags []string) bool {
	for _, tag := range tags {
		for _, filter := range f.filter.noneOf {
			if tag == filter {
				return false
			}
		}
	}
	return true
}

//-----------------------------------------------------------------------------
// func: allOf
//-----------------------------------------------------------------------------

func (f *fragment) allOf(tags []string) bool {
	for _, filter := range f.filter.allOf {
		present := false
		for _, tag := range tags {
			if tag == filter {
				present = true
				break
			}
		}
		if !present {
			return false
		}
	}
	return true
}

//-----------------------------------------------------------------------------
// func: caCertificate
//-----------------------------------------------------------------------------

func (d *Data) caCertificate() {

	if d.CaCert != "" {

		data, err := ioutil.ReadFile(d.CaCert)
		if err != nil {
			log.WithField("cmd", "udata").Fatal(err)
		}

		d.CaCert = strings.TrimSpace(strings.
			Replace(string(data), "\n", "\n    ", -1))
	}
}

//-----------------------------------------------------------------------------
// func: zookeeperURL
//-----------------------------------------------------------------------------

func (d *Data) zookeeperURL() {
	for i := 1; i <= d.QuorumCount; i++ {
		d.ZkServers = d.ZkServers + "quorum-" + strconv.Itoa(i) + ":2181"
		if i != d.QuorumCount {
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
// func: hostnameAliases
//-----------------------------------------------------------------------------

func (d *Data) hostnameAliases() {

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
// func: systemdUnits
//-----------------------------------------------------------------------------

func (d *Data) systemdUnits() {

	units := []string{}

	// Agregate units of all roles:
	for _, i := range d.Roles {
		switch i {
		case "quorum":
			units = append(units,
				"etcd2", "docker", "zookeeper", "rexray", "cadvisor", "node-exporter",
				"zookeeper-exporter")
		case "master":
			units = append(units,
				"etcd2", "flanneld", "docker", "rexray", "mesos-master", "mesos-dns",
				"marathon", "confd", "prometheus", "cadvisor", "mesos-master-exporter",
				"node-exporter")
		case "worker":
			units = append(units,
				"etcd2", "flanneld", "docker", "rexray", "go-dnsmasq", "mesos-agent",
				"marathon-lb", "cadvisor", "mesos-agent-exporter", "node-exporter",
				"haproxy-exporter")
		case "border":
			units = append(units,
				"etcd2", "flanneld", "docker", "rexray", "mongodb", "pritunl")
		default:
			log.WithField("cmd", "udata").Fatal("Invalid role: " + i)
		}
	}

	// Delete duplicated units:
	for _, unit := range units {
		if !func(str string, list []string) bool {
			for _, v := range list {
				if v == str {
					return true
				}
			}
			return false
		}(unit, d.SystemdUnits) {
			d.SystemdUnits = append(d.SystemdUnits, unit)
		}
	}
}

//-----------------------------------------------------------------------------
// func: composeTemplate
//-----------------------------------------------------------------------------

func (d *Data) composeTemplate() {

	// Tags used to filter template fragments:
	tags := append(d.Roles, d.IaasProvider)

	// Apply the filter:
	for _, frag := range d.frags {
		if frag.anyOf(tags) {
			if frag.noneOf(tags) {
				if frag.allOf(tags) {
					d.template += frag.data
				}
			}
		}
	}
}

//-----------------------------------------------------------------------------
// func: Render
//-----------------------------------------------------------------------------

// Render takes a Data structure and outputs valid CoreOS cloud-config
// in YAML format to stdout.
func (d *Data) Render() {

	d.caCertificate()   // Retrieve the CA certificate.
	d.zookeeperURL()    // Forge the Zookeeper URL.
	d.rexraySnippet()   // REX-Ray configuration snippet.
	d.hostnameAliases() // Hostname aliases array.
	d.systemdUnits()    // Systemd units array.
	d.loadFragments()   // Load the fragments array.
	d.composeTemplate() // Compose the template.

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
