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
	PostProc // Post-processed data
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
	GzipUdata           bool     // --gzip-udata
	HostID              string   // --host-id
	HostName            string   // --host-name
	IaasProvider        string   // --iaas-provider
	MasterCount         int      // --master-count
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

// PostProc data based on previous flags
type PostProc struct {
	AlertManagers string
	Aliases       []string
	CaCert        string
	EtcdEndpoints string
	EtcdServers   string
	HostTCPPorts  []int
	HostUDPPorts  []int
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
	fragments fragmentSlice
	services  serviceMap
	template  string
	userData  *bytes.Buffer
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
// func: groups
//-----------------------------------------------------------------------------

func groups(prometheus bool) (groups []string) {
	groups = []string{"base"}
	if prometheus {
		groups = append(groups, "insight")
	}
	return
}

//-----------------------------------------------------------------------------
// func: listOfTags
//-----------------------------------------------------------------------------

func (d *CmdData) listOfTags() (tags []string) {

	tags = append(d.Roles, d.IaasProvider)
	tags = append(tags, d.ClusterState)

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
	for _, frag := range d.fragments {
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
	d.services.load(d.Roles, groups(d.Prometheus)) // Predefined services map.
	d.CaCert = caCert(d.CaCertPath)                // Retrieve the CA certificate.
	d.ZkServers = zkServers(d.QuorumCount)         // Forge the Zookeeper URL.
	d.EtcdServers = etcdServers(d.QuorumCount)     // Initial etcd servers URL.
	d.EtcdEndpoints = etcdEndpoints(d.QuorumCount) // Initial etcd endpoints URL.
	d.AlertManagers = alertManagers(d.MasterCount) // Comma separated alertmanagers.
	d.SMTP = smtpURLSplit(d.SMTPURL)               // Split URL into its components.
	d.MesosDNSPort = mesosDNSPort(d.Roles)         // One of 53 or 54.
	d.Aliases = aliases(d.Roles, d.HostName)       // Hostname aliases array.
	d.SystemdUnits = d.services.listUnits()
	d.HostTCPPorts = d.services.listPorts("tcp")
	d.HostUDPPorts = d.services.listPorts("udp")

	// Template:
	d.fragments.load()  // Predefined template fragments.
	d.composeTemplate() // Compose the template.
	d.renderTemplate()  // Render the template into memory.

	// User data:
	d.validateUserData() // Validate the generated user data.
	d.outputUserData()   // Output user data to stdout.
}
