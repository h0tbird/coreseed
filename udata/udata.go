package udata

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	// Community:
	log "github.com/Sirupsen/logrus"
	"github.com/coreos/coreos-cloudinit/config/validate"
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

// ExtData contains variables to be interpolated in templates.
type ExtData struct {
	QuorumCount         int
	MasterCount         int
	GzipUdata           bool
	Prometheus          bool
	ClusterID           string
	ClusterState        string
	HostName            string
	HostID              string
	Domain              string
	Ns1ApiKey           string
	CaCert              string
	EtcdToken           string
	EtcdServers         string
	EtcdEndpoints       string
	ZkServers           string
	AlertManagers       string
	MesosDNSPort        string
	NetworkBackend      string
	CalicoIPPool        string
	FlannelNetwork      string
	FlannelSubnetLen    string
	FlannelSubnetMin    string
	FlannelSubnetMax    string
	FlannelBackend      string
	RexrayStorageDriver string
	RexrayEndpointIP    string
	Ec2Region           string
	IaasProvider        string
	SlackWebhook        string
	SysdigAccessKey     string
	DatadogAPIKey       string
	SMTPURL             string
	SMTPHost            string
	SMTPPort            string
	SMTPUser            string
	SMTPPass            string
	AdminEmail          string
	Roles               []string
	Aliases             []string
	SystemdUnits        []string
	StubZones           []string
}

type intData struct {
	buffer   *bytes.Buffer
	frags    []fragment
	template string
}

// Data struct
type Data struct {
	ExtData
	intData
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
// func: etcdURL
//-----------------------------------------------------------------------------

func (d *Data) etcdURL() {
	for i := 1; i <= d.QuorumCount; i++ {
		d.EtcdServers = d.EtcdServers +
			"quorum-" + strconv.Itoa(i) + "=http://quorum-" + strconv.Itoa(i) + ":2380"
		d.EtcdEndpoints = d.EtcdEndpoints +
			"http://quorum-" + strconv.Itoa(i) + ":2379"
		if i != d.QuorumCount {
			d.EtcdServers = d.EtcdServers + ","
			d.EtcdEndpoints = d.EtcdEndpoints + ","
		}
	}
}

//-----------------------------------------------------------------------------
// func: alertmanagerURL
//-----------------------------------------------------------------------------

func (d *Data) alertmanagerURL() {
	for i := 1; i <= d.MasterCount; i++ {
		d.AlertManagers = d.AlertManagers + "http://master-" + strconv.Itoa(i) + ":9093"
		if i != d.MasterCount {
			d.AlertManagers = d.AlertManagers + ","
		}
	}
}

//-----------------------------------------------------------------------------
// func: smtpURL
//-----------------------------------------------------------------------------

func (d *Data) smtpURL() {
	if d.SMTPURL != "" {
		r, _ := regexp.Compile("^smtp://(.+):(.+)@(.+):(\\d+)$")
		if sub := r.FindStringSubmatch(d.SMTPURL); sub != nil {
			d.SMTPUser = sub[1]
			d.SMTPPass = sub[2]
			d.SMTPHost = sub[3]
			d.SMTPPort = sub[4]
		} else {
			log.WithFields(log.Fields{"cmd": "udata", "id": d.SMTPURL}).
				Fatal("Invalid SMTP URL format.")
		}
	}
}

//-----------------------------------------------------------------------------
// func: mesosDNSPort
//-----------------------------------------------------------------------------

func (d *Data) mesosDNSPort() {
	d.MesosDNSPort = "53"
	for _, role := range d.Roles {
		if role == "master" {
			for _, role := range d.Roles {
				if role == "worker" {
					d.MesosDNSPort = "54"
					return
				}
			}
		}
	}
}

//-----------------------------------------------------------------------------
// func: hostnameAliases
//-----------------------------------------------------------------------------

func (d *Data) hostnameAliases() {
	for _, i := range d.Roles {
		if i != d.HostName {
			d.Aliases = append(d.Aliases, i)
		}
	}
}

//-----------------------------------------------------------------------------
// func: systemdUnits
//-----------------------------------------------------------------------------

func (d *Data) systemdUnits() {

	units := []string{}

	// Aggregate units of all roles:
	for _, i := range d.Roles {

		switch i {

		case "quorum":
			units = append(units,
				"etcd2", "docker", "zookeeper", "rexray", "etchost.timer")
			if d.Prometheus {
				units = append(units,
					"cadvisor", "node-exporter", "zookeeper-exporter")
			}

		case "master":
			units = append(units,
				"etcd2", d.NetworkBackend, "docker", "rexray", "mesos-master",
				"mesos-dns", "marathon", "etchost.timer")
			if d.Prometheus {
				units = append(units,
					"cadvisor", "confd", "alertmanager", "prometheus", "node-exporter",
					"mesos-master-exporter")
			}

		case "worker":
			units = append(units,
				"etcd2", d.NetworkBackend, "docker", "rexray", "go-dnsmasq",
				"mesos-agent", "marathon-lb", "etchost.timer")
			if d.Prometheus {
				units = append(units,
					"cadvisor", "mesos-agent-exporter", "node-exporter", "haproxy-exporter")
			}

		case "border":
			units = append(units,
				"etcd2", d.NetworkBackend, "docker", "rexray", "mongodb", "pritunl",
				"etchost.timer")

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
// func: listOfTags
//-----------------------------------------------------------------------------

func (d *Data) listOfTags() (tags []string) {

	tags = append(d.Roles, d.IaasProvider)
	tags = append(tags, d.ClusterState)
	tags = append(tags, d.NetworkBackend)

	if d.CaCert != "" {
		tags = append(tags, "cacert")
	}

	if d.Ns1ApiKey != "" {
		tags = append(tags, "ns1")
	}

	if d.SysdigAccessKey != "" {
		tags = append(tags, "sysdig")
	}

	if d.DatadogAPIKey != "" {
		tags = append(tags, "datadog")
	}

	if d.Prometheus {
		tags = append(tags, "prometheus")
	}

	return
}

//-----------------------------------------------------------------------------
// func: composeTemplate
//-----------------------------------------------------------------------------

func (d *Data) composeTemplate() {

	// Tags used to filter template fragments:
	tags := d.listOfTags()

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
// func: renderTemplate
//-----------------------------------------------------------------------------

func (d *Data) renderTemplate() {

	// Template parsing:
	t := template.New("udata")
	t, err := t.Parse(d.template)
	if err != nil {
		log.WithField("cmd", "udata").Fatal(err)
	}

	// Apply parsed template to data object:
	d.buffer = bytes.NewBuffer(make([]byte, 0, 65536))
	if err = t.Execute(d.buffer, d); err != nil {
		log.WithField("cmd", "udata").Fatal(err)
	}
}

//-----------------------------------------------------------------------------
// func: validateUserData
//-----------------------------------------------------------------------------

func (d *Data) validateUserData() {

	errors := []string{}

	report, err := validate.Validate(d.buffer.Bytes())
	if err != nil {
		errors = append(errors, fmt.Sprintf("%v", err))
	}
	for _, entry := range report.Entries() {
		errors = append(errors, fmt.Sprintf("%v", entry))
	}
	if len(errors) > 0 {
		log.WithField("cmd", "udata").Fatal(errors)
	}
}

//-----------------------------------------------------------------------------
// func: outputUserData
//-----------------------------------------------------------------------------

func (d *Data) outputUserData() {

	if d.GzipUdata {
		log.WithFields(log.Fields{"cmd": "udata", "id": d.HostName + "-" + d.HostID}).
			Info("Generating gzipped cloud-config user data")
		w := gzip.NewWriter(os.Stdout)
		if _, err := d.buffer.WriteTo(w); err != nil {
			_ = w.Close()
			log.WithField("cmd", "udata").Fatal(err)
		}
		_ = w.Close()
	} else {
		log.WithField("cmd", "udata").Info("Generating plain text cloud-config user data")
		if _, err := d.buffer.WriteTo(os.Stdout); err != nil {
			log.WithField("cmd", "udata").Fatal(err)
		}
	}
}

//-----------------------------------------------------------------------------
// func: Generate
//-----------------------------------------------------------------------------

// Generate takes a Data structure and outputs valid CoreOS cloud-config
// user data in YAML format to stdout.
func (d *Data) Generate() {

	// Variables:
	d.caCertificate()   // Retrieve the CA certificate.
	d.zookeeperURL()    // Forge the Zookeeper URL.
	d.etcdURL()         // Initial etcd cluster URL.
	d.alertmanagerURL() // Comma separated alertmanagers.
	d.smtpURL()         // Split URL into its components.
	d.mesosDNSPort()    // One of 53 or 54.
	d.hostnameAliases() // Hostname aliases array.
	d.systemdUnits()    // Systemd units array.

	// Template:
	d.loadFragments()   // Load the fragments array.
	d.composeTemplate() // Compose the template.
	d.renderTemplate()  // Render the template into memory.

	// User data:
	d.validateUserData() // Validate the generated user data.
	d.outputUserData()   // Output user data to stdout.
}
