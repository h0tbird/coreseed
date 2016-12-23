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
// func: loadServices
//-----------------------------------------------------------------------------

func (d *Data) loadServices() {

	//----------------------------------

	d.services = append(d.services, service{
		name:     "foo",
		tcpPorts: []int{1, 2},
	})

	//----------------------------------

	d.services = append(d.services, service{
		name:     "bar",
		tcpPorts: []int{3, 4},
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
			&d.services[0],
			&d.services[1],
		},
	})

	//----------------------------------

	d.roles = append(d.roles, role{
		name: "master",
		baseServices: []*service{
			&d.services[0],
			&d.services[1],
		},
	})
}
