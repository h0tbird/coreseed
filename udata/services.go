package udata

//-----------------------------------------------------------------------------
// Import statements:
//-----------------------------------------------------------------------------

import "sort"

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

type serviceMap map[string]service

//-----------------------------------------------------------------------------
// func: findOne
//-----------------------------------------------------------------------------

func findOne(src, dst []string) bool {
	for _, i := range src {
		for _, j := range dst {
			if i == j {
				return true
			}
		}
	}
	return false
}

//-----------------------------------------------------------------------------
// func: listUnits
//-----------------------------------------------------------------------------

func (s *serviceMap) listUnits() (list []string) {

	// Map as set:
	m := map[string]struct{}{}
	for _, service := range *s {
		m[service.name] = struct{}{}
	}

	// Set to slice:
	for k := range m {
		list = append(list, k)
	}

	// Sort and return:
	sort.Strings(list)
	return
}

//-----------------------------------------------------------------------------
// func: listPorts
//-----------------------------------------------------------------------------

func (s *serviceMap) listPorts(protocol string) (list []int) {

	// Default ports:
	m := map[int]struct{}{}
	if protocol == "tcp" {
		m[22] = struct{}{}
	}

	// Map as set:
	for _, service := range *s {
		for _, port := range service.ports {
			if port.protocol == protocol {
				m[port.num] = struct{}{}
			}
		}
	}

	// Set to slice:
	for k := range m {
		list = append(list, k)
	}

	// Sort and return:
	sort.Ints(list)
	return
}

//-----------------------------------------------------------------------------
// func: load
//-----------------------------------------------------------------------------

func (s *serviceMap) load(roles, groups []string) {

	// Map roles to services:
	roleServices := map[string][]string{

		"quorum": {
			"docker", "rexray", "etchost", "zookeeper", "etcd-master", "rkt-api",
			"cadvisor", "node-exporter", "zookeeper-exporter"},

		"master": {
			"docker", "rexray", "etchost", "etcd-proxy", "calico", "mesos-dns",
			"mesos-master", "marathon", "rkt-api", "cadvisor", "node-exporter",
			"mesos-master-exporter", "confd", "alertmanager", "prometheus"},

		"worker": {
			"docker", "rexray", "etchost", "etcd-proxy", "calico", "go-dnsmasq",
			"marathon-lb", "mesos-agent", "rkt-api", "cadvisor", "node-exporter",
			"mesos-agent-exporter", "haproxy-exporter"},

		"border": {
			"docker", "rexray", "etchost", "etcd-proxy", "calico", "mongodb",
			"pritunl", "rkt-api", "cadvisor", "node-exporter"},
	}

	// Map services to config:
	serviceConfig := map[string]service{

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
			ports: []port{
				{num: 4194, protocol: "tcp", ingress: ""},
			},
		},

		"node-exporter": {
			name:   "node-exporter.service",
			groups: []string{"insight"},
			ports: []port{
				{num: 9101, protocol: "tcp", ingress: ""},
			},
		},

		"zookeeper-exporter": {
			name:   "zookeeper-exporter.service",
			groups: []string{"insight"},
			ports: []port{
				{num: 9103, protocol: "tcp", ingress: ""},
			},
		},

		"mesos-master-exporter": {
			name:   "mesos-master-exporter.service",
			groups: []string{"insight"},
			ports: []port{
				{num: 9104, protocol: "tcp", ingress: ""},
			},
		},

		"mesos-agent-exporter": {
			name:   "mesos-agent-exporter.service",
			groups: []string{"insight"},
			ports: []port{
				{num: 9105, protocol: "tcp", ingress: ""},
			},
		},

		"haproxy-exporter": {
			name:   "haproxy-exporter.service",
			groups: []string{"insight"},
			ports: []port{
				{num: 9102, protocol: "tcp", ingress: ""},
			},
		},

		"confd": {
			name:   "confd.service",
			groups: []string{"insight"},
		},

		"alertmanager": {
			name:   "alertmanager.service",
			groups: []string{"insight"},
			ports: []port{
				{num: 9093, protocol: "tcp", ingress: ""},
			},
		},

		"prometheus": {
			name:   "prometheus.service",
			groups: []string{"insight"},
			ports: []port{
				{num: 9191, protocol: "tcp", ingress: ""},
			},
		},
	}

	// Filter my services:
	*s = serviceMap{}
	for _, role := range roles {
		for _, service := range roleServices[role] {
			if findOne(serviceConfig[service].groups, groups) {
				(*s)[service] = serviceConfig[service]
			}
		}
	}
}
