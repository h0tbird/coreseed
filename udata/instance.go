package udata

//-----------------------------------------------------------------------------
// Typedefs:
//-----------------------------------------------------------------------------

type instance struct {
	name  string
	roles []*role
}

type role struct {
	name         string
	baseServices []*service
	promServices []*service
}

type service struct {
	name     string
	tcpPorts []int
	udpPorts []int
}

//-----------------------------------------------------------------------------
// func: serviceLink
//-----------------------------------------------------------------------------

func serviceLink(s *[]service, n string) *service {
	return nil
}

//-----------------------------------------------------------------------------
// func: loadServices
//-----------------------------------------------------------------------------

func (d *Data) loadServices() {

	//----------------------------------

	d.services = append(d.services, service{
		name:     "etcd2",
		tcpPorts: []int{2379, 2380},
	})

	//----------------------------------

	d.services = append(d.services, service{
		name: "docker",
	})

	//----------------------------------

	d.services = append(d.services, service{
		name:     "zookeeper",
		tcpPorts: []int{2888, 3888, 2181},
	})
}

//-----------------------------------------------------------------------------
// func: loadRoles
//-----------------------------------------------------------------------------

func (d *Data) loadRoles() {

	//----------------------------------

	d.roles = append(d.roles, role{
		name: "quorum",
		baseServices: []*service{
			serviceLink(&d.services, "etcd2"),
			serviceLink(&d.services, "docker"),
			serviceLink(&d.services, "zookeeper"),
			serviceLink(&d.services, "rexray"),
			serviceLink(&d.services, "etchost.timer"),
		},
		promServices: []*service{
			serviceLink(&d.services, "rkt-api"),
			serviceLink(&d.services, "cadvisor"),
			serviceLink(&d.services, "node-exporter"),
			serviceLink(&d.services, "zookeeper-exporter"),
		},
	})

	//----------------------------------

	d.roles = append(d.roles, role{
		name: "master",
		baseServices: []*service{
			serviceLink(&d.services, "etcd2"),
			serviceLink(&d.services, d.NetworkBackend),
			serviceLink(&d.services, "docker"),
			serviceLink(&d.services, "rexray"),
			serviceLink(&d.services, "mesos-master"),
			serviceLink(&d.services, "mesos-dns"),
			serviceLink(&d.services, "marathon"),
			serviceLink(&d.services, "etchost.timer"),
		},
		promServices: []*service{
			serviceLink(&d.services, "rkt-api"),
			serviceLink(&d.services, "cadvisor"),
			serviceLink(&d.services, "confd"),
			serviceLink(&d.services, "alertmanager"),
			serviceLink(&d.services, "prometheus"),
			serviceLink(&d.services, "node-exporter"),
			serviceLink(&d.services, "mesos-master-exporter"),
		},
	})

	//----------------------------------

	d.roles = append(d.roles, role{
		name: "worker",
		baseServices: []*service{
			serviceLink(&d.services, "etcd2"),
			serviceLink(&d.services, d.NetworkBackend),
			serviceLink(&d.services, "docker"),
			serviceLink(&d.services, "rexray"),
			serviceLink(&d.services, "go-dnsmasq"),
			serviceLink(&d.services, "mesos-agent"),
			serviceLink(&d.services, "marathon-lb"),
			serviceLink(&d.services, "etchost.timer"),
		},
		promServices: []*service{
			serviceLink(&d.services, "rkt-api"),
			serviceLink(&d.services, "cadvisor"),
			serviceLink(&d.services, "mesos-agent-exporter"),
			serviceLink(&d.services, "node-exporter"),
			serviceLink(&d.services, "haproxy-exporter"),
		},
	})

	//----------------------------------

	d.roles = append(d.roles, role{
		name: "border",
		baseServices: []*service{
			serviceLink(&d.services, "etcd2"),
			serviceLink(&d.services, d.NetworkBackend),
			serviceLink(&d.services, "docker"),
			serviceLink(&d.services, "rexray"),
			serviceLink(&d.services, "mongodb"),
			serviceLink(&d.services, "pritunl"),
			serviceLink(&d.services, "etchost.timer"),
		},
	})
}
