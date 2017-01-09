package udata

//-----------------------------------------------------------------------------
// Typedefs:
//-----------------------------------------------------------------------------

type instance struct {
	name  string
	roles []*role
}

type role struct {
	name     string
	services []*service
}

type service struct {
	name  string
	tags  []string
	ports []port
}

type port struct {
	num      int
	protocol string
	ingress  string
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

func (d *CmdData) loadServices() {

	//---------------------------------------------

	d.services = append(d.services, service{
		name: "docker",
		tags: []string{"base"},
		ports: []port{
			{num: 2375, protocol: "tcp", ingress: ""},
		},
	})

	//---------------------------------------------

	d.services = append(d.services, service{
		name: "etcd2",
		tags: []string{"base"},
		ports: []port{
			{num: 2379, protocol: "tcp", ingress: ""},
			{num: 2380, protocol: "tcp", ingress: ""},
		},
	})

	//---------------------------------------------

	d.services = append(d.services, service{
		name: "go-dnsmasq",
		tags: []string{"base"},
		ports: []port{
			{num: 53, protocol: "tcp", ingress: ""},
		},
	})

	//---------------------------------------------

	d.services = append(d.services, service{
		name: "marathon",
		tags: []string{"base"},
		ports: []port{
			{num: 8080, protocol: "tcp", ingress: ""},
			{num: 9292, protocol: "tcp", ingress: ""},
		},
	})

	//---------------------------------------------

	d.services = append(d.services, service{
		name: "marathon-lb",
		tags: []string{"base"},
		ports: []port{
			{num: 80, protocol: "tcp", ingress: ""},
			{num: 443, protocol: "tcp", ingress: ""},
			{num: 9090, protocol: "tcp", ingress: ""},
			{num: 9091, protocol: "tcp", ingress: ""},
		},
	})

	//---------------------------------------------

	d.services = append(d.services, service{
		name: "mesos-dns",
		tags: []string{"base"},
		ports: []port{
			{num: 53, protocol: "tcp", ingress: ""},
			{num: 54, protocol: "tcp", ingress: ""},
		},
	})

	//---------------------------------------------

	d.services = append(d.services, service{
		name: "mesos-master",
		tags: []string{"base"},
		ports: []port{
			{num: 5050, protocol: "tcp", ingress: ""},
		},
	})

	//---------------------------------------------

	d.services = append(d.services, service{
		name: "mesos-slave",
		tags: []string{"base"},
		ports: []port{
			{num: 5051, protocol: "tcp", ingress: ""},
		},
	})

	//---------------------------------------------

	d.services = append(d.services, service{
		name: "rexray",
		tags: []string{"base"},
		ports: []port{
			{num: 7979, protocol: "tcp", ingress: ""},
		},
	})

	//---------------------------------------------

	d.services = append(d.services, service{
		name: "zookeeper",
		tags: []string{"base"},
		ports: []port{
			{num: 2181, protocol: "tcp", ingress: ""},
			{num: 2888, protocol: "tcp", ingress: ""},
			{num: 3888, protocol: "tcp", ingress: ""},
		},
	})
}

//-----------------------------------------------------------------------------
// func: loadRoles
//-----------------------------------------------------------------------------

func (d *CmdData) loadRoles() {

	//---------------------------------------------

	d.roles = append(d.roles, role{
		name: "quorum",
		services: []*service{
			serviceLink(&d.services, "etcd2"),
			serviceLink(&d.services, "docker"),
			serviceLink(&d.services, "zookeeper"),
			serviceLink(&d.services, "rexray"),
			serviceLink(&d.services, "etchost.timer"),
			serviceLink(&d.services, "rkt-api"),
			serviceLink(&d.services, "cadvisor"),
			serviceLink(&d.services, "node-exporter"),
			serviceLink(&d.services, "zookeeper-exporter"),
		},
	})

	//---------------------------------------------

	d.roles = append(d.roles, role{
		name: "master",
		services: []*service{
			serviceLink(&d.services, "etcd2"),
			serviceLink(&d.services, d.NetworkBackend),
			serviceLink(&d.services, "docker"),
			serviceLink(&d.services, "rexray"),
			serviceLink(&d.services, "mesos-master"),
			serviceLink(&d.services, "mesos-dns"),
			serviceLink(&d.services, "marathon"),
			serviceLink(&d.services, "etchost.timer"),
			serviceLink(&d.services, "rkt-api"),
			serviceLink(&d.services, "cadvisor"),
			serviceLink(&d.services, "confd"),
			serviceLink(&d.services, "alertmanager"),
			serviceLink(&d.services, "prometheus"),
			serviceLink(&d.services, "node-exporter"),
			serviceLink(&d.services, "mesos-master-exporter"),
		},
	})

	//---------------------------------------------

	d.roles = append(d.roles, role{
		name: "worker",
		services: []*service{
			serviceLink(&d.services, "etcd2"),
			serviceLink(&d.services, d.NetworkBackend),
			serviceLink(&d.services, "docker"),
			serviceLink(&d.services, "rexray"),
			serviceLink(&d.services, "go-dnsmasq"),
			serviceLink(&d.services, "mesos-agent"),
			serviceLink(&d.services, "marathon-lb"),
			serviceLink(&d.services, "etchost.timer"),
			serviceLink(&d.services, "rkt-api"),
			serviceLink(&d.services, "cadvisor"),
			serviceLink(&d.services, "mesos-agent-exporter"),
			serviceLink(&d.services, "node-exporter"),
			serviceLink(&d.services, "haproxy-exporter"),
		},
	})

	//---------------------------------------------

	d.roles = append(d.roles, role{
		name: "border",
		services: []*service{
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
