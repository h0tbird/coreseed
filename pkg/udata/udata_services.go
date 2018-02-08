package udata

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"sort"
	"strconv"
)

//-----------------------------------------------------------------------------
// Typedefs:
//-----------------------------------------------------------------------------

type service struct {
	name   string
	roles  []string
	groups []string
	ports  []portRange
}

type portRange struct {
	interval startEnd
	protocol string
	ingress  string
}

type startEnd struct {
	start, end int
}

type serviceMap map[string]service

//-----------------------------------------------------------------------------
// Custom sort:
//-----------------------------------------------------------------------------

type byStart []startEnd

func (a byStart) Len() int {
	return len(a)
}

func (a byStart) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a byStart) Less(i, j int) bool {
	return a[i].start > a[j].start
}

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
// func: minInt, maxInt
//-----------------------------------------------------------------------------

func minInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func maxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
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
// func: listIvals
//-----------------------------------------------------------------------------

func (s *serviceMap) listIvals(protocol string) (list []startEnd) {
	for _, service := range *s {
		for _, port := range service.ports {
			if port.protocol == protocol {
				list = append(list, port.interval)
			}
		}
	}
	return
}

//-----------------------------------------------------------------------------
// func: listPorts
//-----------------------------------------------------------------------------

func (s *serviceMap) listPorts(protocol string) (list []string) {

	// Sort the intervals array:
	ival := s.listIvals(protocol)
	sort.Sort(byStart(ival))
	i := 0

	// Merge the intervals:
	for _, v := range ival {

		// Overlap
		if i != 0 && ival[i-1].start <= v.end {
			for i != 0 && ival[i-1].start <= v.end {
				ival[i-1].end = maxInt(ival[i-1].end, v.end)
				ival[i-1].start = minInt(ival[i-1].start, v.start)
				i--
			}

			// Adjacent
		} else if i != 0 && ival[i-1].end == v.start+1 {
			ival[i-1].start = v.start
			ival[i-1].end = maxInt(ival[i-1].end, v.end)
			i--

			// Outlying
		} else {
			ival[i] = v
		}

		i++
	}

	// Append to the list:
	for j := 0; j < i; j++ {
		if ival[j].start == ival[j].end {
			list = append(list, strconv.Itoa(ival[j].start))
		} else {
			list = append(list,
				strconv.Itoa(ival[j].start)+":"+strconv.Itoa(ival[j].end))
		}
	}

	return
}

//-----------------------------------------------------------------------------
// func: load
//-----------------------------------------------------------------------------

func (s *serviceMap) load(roles, groups []string) {

	// Map roles to services:
	roleServices := map[string][]string{

		"quorum": {
			"docker", "rexray", "etchosts", "zookeeper", "etcd-master", "rkt-api",
			"cadvisor", "node-exporter", "zookeeper-exporter"},

		"master": {
			"docker", "rexray", "etchosts", "etcd-proxy", "calico", "mesos-dns",
			"mesos-master", "marathon", "rkt-api", "cadvisor", "node-exporter",
			"mesos-master-exporter", "confd", "alertmanager", "prometheus"},

		"worker": {
			"docker", "rexray", "etchosts", "etcd-proxy", "calico", "go-dnsmasq",
			"marathon-lb", "mesos-agent", "rkt-api", "cadvisor", "node-exporter",
			"mesos-agent-exporter", "haproxy-exporter"},

		"border": {
			"docker", "rexray", "etchosts", "etcd-proxy", "calico", "mongodb",
			"pritunl", "rkt-api", "cadvisor", "node-exporter"},
	}

	// Map services to config:
	serviceConfig := map[string]service{

		"docker": {
			name:   "docker.service",
			groups: []string{"base"},
			ports: []portRange{
				{interval: startEnd{2375, 2375}, protocol: "tcp", ingress: ""},
			},
		},

		"rexray": {
			name:   "rexray.service",
			groups: []string{"base"},
			ports: []portRange{
				{interval: startEnd{7979, 7979}, protocol: "tcp", ingress: ""},
			},
		},

		"etchosts": {
			name:   "etchosts.timer",
			groups: []string{"base"},
			ports: []portRange{
				{interval: startEnd{22, 22}, protocol: "tcp", ingress: ""},
			},
		},

		"etcd-proxy": {
			name:   "etcd-member.service",
			groups: []string{"base"},
			ports: []portRange{
				{interval: startEnd{2379, 2379}, protocol: "tcp", ingress: ""},
			},
		},

		"calico": {
			name:   "calico.service",
			groups: []string{"base"},
			ports: []portRange{
				{interval: startEnd{179, 179}, protocol: "tcp", ingress: ""},
			},
		},

		"zookeeper": {
			name:   "zookeeper.service",
			groups: []string{"base"},
			ports: []portRange{
				{interval: startEnd{2181, 2181}, protocol: "tcp", ingress: ""},
				{interval: startEnd{2888, 2888}, protocol: "tcp", ingress: ""},
				{interval: startEnd{3888, 3888}, protocol: "tcp", ingress: ""},
			},
		},

		"etcd-master": {
			name:   "etcd-member.service",
			groups: []string{"base"},
			ports: []portRange{
				{interval: startEnd{2379, 2380}, protocol: "tcp", ingress: ""},
			},
		},

		"mesos-dns": {
			name:   "mesos-dns.service",
			groups: []string{"base"},
			ports: []portRange{
				{interval: startEnd{53, 54}, protocol: "tcp", ingress: ""},
				{interval: startEnd{53, 54}, protocol: "udp", ingress: ""},
			},
		},

		"mesos-master": {
			name:   "mesos-master.service",
			groups: []string{"base"},
			ports: []portRange{
				{interval: startEnd{5050, 5050}, protocol: "tcp", ingress: ""},
			},
		},

		"marathon": {
			name:   "marathon.service",
			groups: []string{"base"},
			ports: []portRange{
				{interval: startEnd{8080, 8080}, protocol: "tcp", ingress: ""},
				{interval: startEnd{9292, 9292}, protocol: "tcp", ingress: ""},
			},
		},

		"go-dnsmasq": {
			name:   "go-dnsmasq.service",
			groups: []string{"base"},
			ports: []portRange{
				{interval: startEnd{53, 53}, protocol: "tcp", ingress: ""},
			},
		},

		"marathon-lb": {
			name:   "marathon-lb.service",
			groups: []string{"base"},
			ports: []portRange{
				{interval: startEnd{80, 80}, protocol: "tcp", ingress: ""},
				{interval: startEnd{443, 443}, protocol: "tcp", ingress: ""},
				{interval: startEnd{9090, 9091}, protocol: "tcp", ingress: ""},
				{interval: startEnd{10000, 10100}, protocol: "tcp", ingress: ""},
			},
		},

		"mesos-agent": {
			name:   "mesos-agent.service",
			groups: []string{"base"},
			ports: []portRange{
				{interval: startEnd{5051, 5051}, protocol: "tcp", ingress: ""},
			},
		},

		"mongodb": {
			name:   "mongodb.service",
			groups: []string{"base"},
			ports: []portRange{
				{interval: startEnd{27017, 27017}, protocol: "tcp", ingress: ""},
			},
		},

		"pritunl": {
			name:   "pritunl.service",
			groups: []string{"base"},
			ports: []portRange{
				{interval: startEnd{80, 80}, protocol: "tcp", ingress: ""},
				{interval: startEnd{443, 443}, protocol: "tcp", ingress: ""},
				{interval: startEnd{9756, 9756}, protocol: "tcp", ingress: ""},
				{interval: startEnd{18443, 18443}, protocol: "udp", ingress: ""},
			},
		},

		"rkt-api": {
			name:   "rkt-api.service",
			groups: []string{"insight"},
		},

		"cadvisor": {
			name:   "cadvisor.service",
			groups: []string{"insight"},
			ports: []portRange{
				{interval: startEnd{4194, 4194}, protocol: "tcp", ingress: ""},
			},
		},

		"node-exporter": {
			name:   "node-exporter.service",
			groups: []string{"insight"},
			ports: []portRange{
				{interval: startEnd{9101, 9101}, protocol: "tcp", ingress: ""},
			},
		},

		"zookeeper-exporter": {
			name:   "zookeeper-exporter.service",
			groups: []string{"insight"},
			ports: []portRange{
				{interval: startEnd{9103, 9103}, protocol: "tcp", ingress: ""},
			},
		},

		"mesos-master-exporter": {
			name:   "mesos-master-exporter.service",
			groups: []string{"insight"},
			ports: []portRange{
				{interval: startEnd{9104, 9104}, protocol: "tcp", ingress: ""},
			},
		},

		"mesos-agent-exporter": {
			name:   "mesos-agent-exporter.service",
			groups: []string{"insight"},
			ports: []portRange{
				{interval: startEnd{9105, 9105}, protocol: "tcp", ingress: ""},
			},
		},

		"haproxy-exporter": {
			name:   "haproxy-exporter.service",
			groups: []string{"insight"},
			ports: []portRange{
				{interval: startEnd{9102, 9102}, protocol: "tcp", ingress: ""},
			},
		},

		"confd": {
			name:   "confd.service",
			groups: []string{"insight"},
		},

		"alertmanager": {
			name:   "alertmanager.service",
			groups: []string{"insight"},
			ports: []portRange{
				{interval: startEnd{9093, 9093}, protocol: "tcp", ingress: ""},
			},
		},

		"prometheus": {
			name:   "prometheus.service",
			groups: []string{"insight"},
			ports: []portRange{
				{interval: startEnd{9191, 9191}, protocol: "tcp", ingress: ""},
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
