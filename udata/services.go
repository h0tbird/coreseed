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

type serviceMap struct {
	roleServices  map[string][]string
	serviceConfig map[string]service
}

//-----------------------------------------------------------------------------
// func: load
//-----------------------------------------------------------------------------

func (s *serviceMap) load() {

	// Map roles to services:
	s.roleServices = map[string][]string{

		"quorum": []string{
			"docker", "rexray", "etchost", "zookeeper", "etcd-master", "rkt-api",
			"cadvisor", "node-exporter", "zookeeper-exporter"},

		"master": []string{
			"docker", "rexray", "etchost", "etcd-proxy", "calico", "mesos-dns",
			"mesos-master", "marathon", "rkt-api", "cadvisor", "node-exporter",
			"mesos-master-exporter", "confd", "alertmanager", "prometheus"},

		"worker": []string{
			"docker", "rexray", "etchost", "etcd-proxy", "calico", "go-dnsmasq",
			"marathon-lb", "mesos-agent", "rkt-api", "cadvisor", "node-exporter",
			"mesos-agent-exporter", "haproxy-exporter"},

		"border": []string{
			"docker", "rexray", "etchost", "etcd-proxy", "calico", "mongodb",
			"pritunl", "rkt-api", "cadvisor", "node-exporter"},
	}

	// Map services to config:
	s.serviceConfig = map[string]service{

		"docker": {
			name:   "docker.service",
			groups: []string{"base"},
			ports: []port{
				{num: 2375, protocol: "tcp", ingress: ""},
			},
		},

		"rexray": {
			name:   "rexray.service",
			groups: []string{"base"},
			ports: []port{
				{num: 7979, protocol: "tcp", ingress: ""},
			},
		},

		"etchost": {
			name:   "etchost.timer",
			groups: []string{"base"},
		},

		"etcd-proxy": {
			name:   "etcd2.service",
			groups: []string{"base"},
			ports: []port{
				{num: 2379, protocol: "tcp", ingress: ""},
			},
		},

		"calico": {
			name:   "calico.service",
			groups: []string{"base"},
		},

		"zookeeper": {
			name:   "zookeeper.service",
			groups: []string{"base"},
			ports: []port{
				{num: 2181, protocol: "tcp", ingress: ""},
				{num: 2888, protocol: "tcp", ingress: ""},
				{num: 3888, protocol: "tcp", ingress: ""},
			},
		},

		"etcd-master": {
			name:   "etcd2.service",
			groups: []string{"base"},
			ports: []port{
				{num: 2379, protocol: "tcp", ingress: ""},
				{num: 2380, protocol: "tcp", ingress: ""},
			},
		},

		"mesos-dns": {
			name:   "mesos-dns.service",
			groups: []string{"base"},
			ports: []port{
				{num: 53, protocol: "tcp", ingress: ""},
				{num: 54, protocol: "tcp", ingress: ""},
			},
		},

		"mesos-master": {
			name:   "mesos-master.service",
			groups: []string{"base"},
			ports: []port{
				{num: 5050, protocol: "tcp", ingress: ""},
			},
		},

		"marathon": {
			name:   "marathon.service",
			groups: []string{"base"},
			ports: []port{
				{num: 8080, protocol: "tcp", ingress: ""},
				{num: 9292, protocol: "tcp", ingress: ""},
			},
		},

		"go-dnsmasq": {
			name:   "go-dnsmasq.service",
			groups: []string{"base"},
			ports: []port{
				{num: 53, protocol: "tcp", ingress: ""},
			},
		},

		"marathon-lb": {
			name:   "marathon-lb.service",
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
			groups: []string{"base"},
			ports: []port{
				{num: 5051, protocol: "tcp", ingress: ""},
			},
		},

		"mongodb": {
			name:   "mongodb.service",
			groups: []string{"base"},
			ports: []port{
				{num: 27017, protocol: "tcp", ingress: ""},
			},
		},

		"pritunl": {
			name:   "pritunl.service",
			groups: []string{"base"},
			ports: []port{
				{num: 80, protocol: "tcp", ingress: ""},
				{num: 443, protocol: "tcp", ingress: ""},
				{num: 9756, protocol: "tcp", ingress: ""},
				{num: 18443, protocol: "udp", ingress: ""},
			},
		},

		"rkt-api": {
			name:   "rkt-api.service",
			groups: []string{"insight"},
		},

		"cadvisor": {
			name:   "cadvisor.service",
			groups: []string{"insight"},
		},

		"node-exporter": {
			name:   "node-exporter.service",
			groups: []string{"insight"},
		},

		"zookeeper-exporter": {
			name:   "zookeeper-exporter.service",
			groups: []string{"insight"},
		},

		"mesos-master-exporter": {
			name:   "mesos-master-exporter.service",
			groups: []string{"insight"},
		},

		"mesos-agent-exporter": {
			name:   "mesos-agent-exporter.service",
			groups: []string{"insight"},
		},

		"haproxy-exporter": {
			name:   "haproxy-exporter.service",
			groups: []string{"insight"},
		},

		"confd": {
			name:   "confd.service",
			groups: []string{"insight"},
		},

		"alertmanager": {
			name:   "alertmanager.service",
			groups: []string{"insight"},
		},

		"prometheus": {
			name:   "prometheus.service",
			groups: []string{"insight"},
		},
	}
}
