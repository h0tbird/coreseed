package udata

//-----------------------------------------------------------------------------
// Typedefs:
//-----------------------------------------------------------------------------

type service struct {
	name   string
	roles  []string
	groups []string
	ports  []port
}

type port struct {
	num      int
	protocol string
	ingress  string
}

type svcMap map[string]service

//-----------------------------------------------------------------------------
// Globals:
//-----------------------------------------------------------------------------

var services = svcMap{

	//---------------------
	// Base service group:
	//---------------------

	"docker": {
		name:   "docker.service",
		roles:  []string{"quorum", "master", "worker", "border"},
		groups: []string{"base"},
		ports: []port{
			{num: 2375, protocol: "tcp", ingress: ""},
		},
	},

	"rexray": {
		name:   "rexray.service",
		roles:  []string{"quorum", "master", "worker", "border"},
		groups: []string{"base"},
		ports: []port{
			{num: 7979, protocol: "tcp", ingress: ""},
		},
	},

	"etchost": {
		name:   "etchost.timer",
		roles:  []string{"quorum", "master", "worker", "border"},
		groups: []string{"base"},
	},

	"etcd-proxy": {
		name:   "etcd2.service",
		roles:  []string{"master", "worker", "border"},
		groups: []string{"base"},
		ports: []port{
			{num: 2379, protocol: "tcp", ingress: ""},
		},
	},

	"calico": {
		name:   "calico.service",
		roles:  []string{"master", "worker", "border"},
		groups: []string{"base"},
	},

	"zookeeper": {
		name:   "zookeeper.service",
		roles:  []string{"quorum"},
		groups: []string{"base"},
		ports: []port{
			{num: 2181, protocol: "tcp", ingress: ""},
			{num: 2888, protocol: "tcp", ingress: ""},
			{num: 3888, protocol: "tcp", ingress: ""},
		},
	},

	"etcd-master": {
		name:   "etcd2.service",
		roles:  []string{"quorum"},
		groups: []string{"base"},
		ports: []port{
			{num: 2379, protocol: "tcp", ingress: ""},
			{num: 2380, protocol: "tcp", ingress: ""},
		},
	},

	"mesos-dns": {
		name:   "mesos-dns.service",
		roles:  []string{"master"},
		groups: []string{"base"},
		ports: []port{
			{num: 53, protocol: "tcp", ingress: ""},
			{num: 54, protocol: "tcp", ingress: ""},
		},
	},

	"mesos-master": {
		name:   "mesos-master.service",
		roles:  []string{"master"},
		groups: []string{"base"},
		ports: []port{
			{num: 5050, protocol: "tcp", ingress: ""},
		},
	},

	"marathon": {
		name:   "marathon.service",
		roles:  []string{"master"},
		groups: []string{"base"},
		ports: []port{
			{num: 8080, protocol: "tcp", ingress: ""},
			{num: 9292, protocol: "tcp", ingress: ""},
		},
	},

	"go-dnsmasq": {
		name:   "go-dnsmasq.service",
		roles:  []string{"worker"},
		groups: []string{"base"},
		ports: []port{
			{num: 53, protocol: "tcp", ingress: ""},
		},
	},

	"marathon-lb": {
		name:   "marathon-lb.service",
		roles:  []string{"worker"},
		groups: []string{"base"},
		ports: []port{
			{num: 80, protocol: "tcp", ingress: ""},
			{num: 443, protocol: "tcp", ingress: ""},
			{num: 9090, protocol: "tcp", ingress: ""},
			{num: 9091, protocol: "tcp", ingress: ""},
		},
	},

	"mesos-agent": {
		name:   "mesos-agent.service",
		roles:  []string{"worker"},
		groups: []string{"base"},
		ports: []port{
			{num: 5051, protocol: "tcp", ingress: ""},
		},
	},

	"mongodb": {
		name:   "mongodb.service",
		roles:  []string{"border"},
		groups: []string{"base"},
		ports: []port{
			{num: 27017, protocol: "tcp", ingress: ""},
		},
	},

	"pritunl": {
		name:   "pritunl.service",
		roles:  []string{"border"},
		groups: []string{"base"},
		ports: []port{
			{num: 80, protocol: "tcp", ingress: ""},
			{num: 443, protocol: "tcp", ingress: ""},
			{num: 9756, protocol: "tcp", ingress: ""},
			{num: 18443, protocol: "udp", ingress: ""},
		},
	},

	//------------------------
	// Insight service group:
	//------------------------

	"rkt-api": {
		name:   "rkt-api.service",
		roles:  []string{"quorum", "master", "worker", "border"},
		groups: []string{"insight"},
	},

	"cadvisor": {
		name:   "cadvisor.service",
		roles:  []string{"quorum", "master", "worker", "border"},
		groups: []string{"insight"},
	},

	"node-exporter": {
		name:   "node-exporter.service",
		roles:  []string{"quorum", "master", "worker", "border"},
		groups: []string{"insight"},
	},

	"zookeeper-exporter": {
		name:   "zookeeper-exporter.service",
		roles:  []string{"quorum"},
		groups: []string{"insight"},
	},

	"mesos-master-exporter": {
		name:   "mesos-master-exporter.service",
		roles:  []string{"master"},
		groups: []string{"insight"},
	},

	"mesos-agent-exporter": {
		name:   "mesos-agent-exporter.service",
		roles:  []string{"worker"},
		groups: []string{"insight"},
	},

	"haproxy-exporter": {
		name:   "haproxy-exporter.service",
		roles:  []string{"worker"},
		groups: []string{"insight"},
	},

	"confd": {
		name:   "confd.service",
		roles:  []string{"master"},
		groups: []string{"insight"},
	},

	"alertmanager": {
		name:   "alertmanager.service",
		roles:  []string{"master"},
		groups: []string{"insight"},
	},

	"prometheus": {
		name:   "prometheus.service",
		roles:  []string{"master"},
		groups: []string{"insight"},
	},
}
