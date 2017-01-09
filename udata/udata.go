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

// CmdData holds the data used by the udata sub-command
type CmdData struct {
	CmdFlags // Command-line flags
	Postproc // Post-processed data
	intData  // Internal logic data
}

// CmdFlags honored by the udata sub-command
type CmdFlags struct {
	AdminEmail          string   // --admin-email
	CaCertPath          string   // --ca-cert-path
	CalicoIPPool        string   // --calico-ip-pool
	ClusterID           string   // --cluster-id
	ClusterState        string   // --cluster-state
	DatadogAPIKey       string   // --datadog-api-key
	Domain              string   // --domain
	Ec2Region           string   // --ec2-region
	EtcdToken           string   // --etcd-token
	FlannelBackend      string   // --flannel-backend
	FlannelNetwork      string   // --flannel-network
	FlannelSubnetLen    string   // --flannel-subnet-len
	FlannelSubnetMax    string   // --flannel-subnet-max
	FlannelSubnetMin    string   // --flannel-subnet-min
	GzipUdata           bool     // --gzip-udata
	HostID              string   // --host-id
	HostName            string   // --host-name
	IaasProvider        string   // --iaas-provider
	MasterCount         int      // --master-count
	NetworkBackend      string   // --network-backend
	Ns1ApiKey           string   // --ns1-api-key
	Prometheus          bool     // --prometheus
	QuorumCount         int      // --quorum-count
	RexrayEndpointIP    string   // --rexray-endpoint-ip
	RexrayStorageDriver string   // --rexray-storage-driver
	Roles               []string // --roles
	SlackWebhook        string   // --slack-webhook
	SMTPURL             string   // --smtp-url
	StubZones           []string // --stub-zone
	SysdigAccessKey     string   // --sysdig-access-key
}

// Postproc data based on previous flags
type Postproc struct {
	AlertManagers string
	Aliases       []string
	CaCert        string
	EtcdEndpoints string
	EtcdServers   string
	MesosDNSPort  int
	SMTP
	SystemdUnits []string
	ZkServers    string
}

// SMTP structure
type SMTP struct {
	Host string
	Port string
	User string
	Pass string
}

// Internal logic data
type intData struct {
	frags    []fragment
	roles    []role
	services []service
	template string
	userData *bytes.Buffer
}

//-----------------------------------------------------------------------------
// func: caCert
//-----------------------------------------------------------------------------

func caCert(path string) (cert string) {
	if path != "" {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.WithField("cmd", "udata").Fatal(err)
		}
		cert = strings.TrimSpace(strings.
			Replace(string(data), "\n", "\n    ", -1))
	}
	return
}

//-----------------------------------------------------------------------------
// func: zkServers
//-----------------------------------------------------------------------------

func zkServers(quorumCount int) (zkServers string) {
	for i := 1; i <= quorumCount; i++ {
		zkServers = zkServers + "quorum-" + strconv.Itoa(i) + ":2181"
		if i != quorumCount {
			zkServers = zkServers + ","
		}
	}
	return
}

//-----------------------------------------------------------------------------
// func: etcdServers
//-----------------------------------------------------------------------------

func etcdServers(quorumCount int) (etcdServers string) {
	for i := 1; i <= quorumCount; i++ {
		etcdServers = etcdServers +
			"quorum-" + strconv.Itoa(i) + "=http://quorum-" + strconv.Itoa(i) + ":2380"
		if i != quorumCount {
			etcdServers = etcdServers + ","
		}
	}
	return
}

//-----------------------------------------------------------------------------
// func: etcdEndpoints
//-----------------------------------------------------------------------------

func etcdEndpoints(quorumCount int) (etcdEndpoints string) {
	for i := 1; i <= quorumCount; i++ {
		etcdEndpoints = etcdEndpoints +
			"http://quorum-" + strconv.Itoa(i) + ":2379"
		if i != quorumCount {
			etcdEndpoints = etcdEndpoints + ","
		}
	}
	return
}

//-----------------------------------------------------------------------------
// func: alertManagers
//-----------------------------------------------------------------------------

func alertManagers(masterCount int) (alertManagers string) {
	for i := 1; i <= masterCount; i++ {
		alertManagers = alertManagers + "http://master-" + strconv.Itoa(i) + ":9093"
		if i != masterCount {
			alertManagers = alertManagers + ","
		}
	}
	return
}

//-----------------------------------------------------------------------------
// func: smtpURLSplit
//-----------------------------------------------------------------------------

func smtpURLSplit(smtpURL string) (smtp SMTP) {
	if smtpURL != "" {
		r, _ := regexp.Compile("^smtp://(.+):(.+)@(.+):(\\d+)$")
		if sub := r.FindStringSubmatch(smtpURL); sub != nil {
			smtp.User = sub[1]
			smtp.Pass = sub[2]
			smtp.Host = sub[3]
			smtp.Port = sub[4]
		} else {
			log.WithFields(log.Fields{"cmd": "udata", "id": smtpURL}).
				Fatal("Invalid SMTP URL format.")
		}
	}
	return
}

//-----------------------------------------------------------------------------
// func: mesosDNSPort
//-----------------------------------------------------------------------------

func mesosDNSPort(roles []string) (mesosDNSPort int) {
	mesosDNSPort = 53
	for _, role := range roles {
		if role == "master" {
			for _, role := range roles {
				if role == "worker" {
					mesosDNSPort = 54
					return
				}
			}
		}
	}
	return
}

//-----------------------------------------------------------------------------
// func: aliases
//-----------------------------------------------------------------------------

func aliases(roles []string, hostName string) (aliases []string) {
	for _, i := range roles {
		if i != hostName {
			aliases = append(aliases, i)
		}
	}
	return
}

//-----------------------------------------------------------------------------
// func: systemdUnits
//-----------------------------------------------------------------------------

func systemdUnits(roles []string, prometheus bool, networkBackend string) []string {

	var units, systemdUnits []string

	// Aggregate units of all roles:
	for _, i := range roles {

		switch i {

		case "quorum":
			units = append(units,
				"etcd2", "docker", "zookeeper", "rexray", "etchost.timer")
			if prometheus {
				units = append(units,
					"rkt-api", "cadvisor", "node-exporter", "zookeeper-exporter")
			}

		case "master":
			units = append(units,
				"etcd2", networkBackend, "docker", "rexray", "mesos-master",
				"mesos-dns", "marathon", "etchost.timer")
			if prometheus {
				units = append(units,
					"rkt-api", "cadvisor", "confd", "alertmanager", "prometheus",
					"node-exporter", "mesos-master-exporter")
			}

		case "worker":
			units = append(units,
				"etcd2", networkBackend, "docker", "rexray", "go-dnsmasq",
				"mesos-agent", "marathon-lb", "etchost.timer")
			if prometheus {
				units = append(units,
					"rkt-api", "cadvisor", "mesos-agent-exporter", "node-exporter",
					"haproxy-exporter")
			}

		case "border":
			units = append(units,
				"etcd2", networkBackend, "docker", "rexray", "mongodb", "pritunl",
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
		}(unit, systemdUnits) {
			systemdUnits = append(systemdUnits, unit)
		}
	}

	return systemdUnits
}

//-----------------------------------------------------------------------------
// func: servicePorts
//-----------------------------------------------------------------------------

func (d *CmdData) servicePorts() {

	// Ports shared by all roles:
	tcpPorts := []string{"22", "2379", "7979"}
	udpPorts := []string{}

	if d.Prometheus {
		tcpPorts = append(tcpPorts, "4194", "9101")
	}

	// Aggregate porst of all roles:
	for _, i := range d.Roles {

		switch i {

		case "quorum":
			tcpPorts = append(tcpPorts, "2181", "2380", "2888", "3888")
			if d.Prometheus {
				tcpPorts = append(tcpPorts, "9103")
			}

		case "master":
			tcpPorts = append(tcpPorts, "53", "5050", "8080", "9292")
			if d.Prometheus {
				tcpPorts = append(tcpPorts, "9093", "9104", "9191")
			}

		case "worker":
			tcpPorts = append(tcpPorts, "53", "80", "443", "5051", "9090", "9091")
			if d.Prometheus {
				tcpPorts = append(tcpPorts, "9102", "9105")
			}

		case "border":
			tcpPorts = append(tcpPorts, "80", "443", "9756", "27017")
			udpPorts = append(udpPorts, "18443")

		default:
			log.WithField("cmd", "udata").Fatal("Invalid role: " + i)
		}
	}
}

//-----------------------------------------------------------------------------
// func: listOfTags
//-----------------------------------------------------------------------------

func (d *CmdData) listOfTags() (tags []string) {

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

func (d *CmdData) composeTemplate() {

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

func (d *CmdData) renderTemplate() {

	// Template parsing:
	t := template.New("udata")
	t, err := t.Parse(d.template)
	if err != nil {
		log.WithField("cmd", "udata").Fatal(err)
	}

	// Apply parsed template to data object:
	d.userData = bytes.NewBuffer(make([]byte, 0, 65536))
	if err = t.Execute(d.userData, d); err != nil {
		log.WithField("cmd", "udata").Fatal(err)
	}
}

//-----------------------------------------------------------------------------
// func: validateUserData
//-----------------------------------------------------------------------------

func (d *CmdData) validateUserData() {

	errors := []string{}

	report, err := validate.Validate(d.userData.Bytes())
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

func (d *CmdData) outputUserData() {

	if d.GzipUdata {
		log.WithFields(log.Fields{"cmd": "udata", "id": d.HostName + "-" + d.HostID}).
			Info("Generating gzipped cloud-config user data")
		w := gzip.NewWriter(os.Stdout)
		if _, err := d.userData.WriteTo(w); err != nil {
			_ = w.Close()
			log.WithField("cmd", "udata").Fatal(err)
		}
		_ = w.Close()
	} else {
		log.WithField("cmd", "udata").Info("Generating plain text cloud-config user data")
		if _, err := d.userData.WriteTo(os.Stdout); err != nil {
			log.WithField("cmd", "udata").Fatal(err)
		}
	}
}

//-----------------------------------------------------------------------------
// func: CmdRun
//-----------------------------------------------------------------------------

// CmdRun takes data from CmdData and outputs valid CoreOS cloud-config user
// data in YAML format to stdout.
func (d *CmdData) CmdRun() {

	// Variables:
	d.CaCert = caCert(d.CaCertPath)                // Retrieve the CA certificate.
	d.ZkServers = zkServers(d.QuorumCount)         // Forge the Zookeeper URL.
	d.EtcdServers = etcdServers(d.QuorumCount)     // Initial etcd servers URL.
	d.EtcdEndpoints = etcdEndpoints(d.QuorumCount) // Initial etcd endpoints URL.
	d.AlertManagers = alertManagers(d.MasterCount) // Comma separated alertmanagers.
	d.SMTP = smtpURLSplit(d.SMTPURL)               // Split URL into its components.
	d.MesosDNSPort = mesosDNSPort(d.Roles)         // One of 53 or 54.
	d.Aliases = aliases(d.Roles, d.HostName)       // Hostname aliases array.
	d.SystemdUnits = systemdUnits(d.Roles,         // Systemd units array.
		d.Prometheus, d.NetworkBackend)

	// Template:
	d.loadFragments()   // Load the fragments array.
	d.composeTemplate() // Compose the template.
	d.renderTemplate()  // Render the template into memory.

	// User data:
	d.validateUserData() // Validate the generated user data.
	d.outputUserData()   // Output user data to stdout.
}
