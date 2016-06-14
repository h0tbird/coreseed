package ns1

import "fmt"

//-----------------------------------------------------------------------------
// Typedefs:
//-----------------------------------------------------------------------------

// Data struct for zone information.
type Data struct {
	APIKey string
	Link   string
	Zones  []string
}

//-----------------------------------------------------------------------------
// func: AddZones
//-----------------------------------------------------------------------------

// AddZones adds one or more zones to NS1.
func (d *Data) AddZones() error {

	for _, e := range d.Zones {
		fmt.Println(e)
	}

	return nil
}
