package r53

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (
	"github.com/katosys/kato/pkg/cli"
)

//-----------------------------------------------------------------------------
// 'katoctl r53' command flags definitions:
//-----------------------------------------------------------------------------

var (

	// r53:
	cmdR53 = cli.App.Command("r53", "Manages Route 53 zones and records.")

	// r53 zone add
	cmdR53Zone    = cmdR53.Command("zone", "Manage Route 53 zones.")
	cmdR53ZoneAdd = cmdR53Zone.Command("add", "Adds Route 53 zones.")

	// r53 record add
	cmdR53Record    = cmdR53.Command("record", "Manage Route 53 records.")
	cmdR53RecordAdd = cmdR53Record.Command("add", "Adds records to Route 53 zones.")
)

//-----------------------------------------------------------------------------
// RunCmd:
//-----------------------------------------------------------------------------

// RunCmd runs the cmd if owned by this package.
func RunCmd(cmd string) bool {

	switch cmd {

	// katoctl r53 zone add:
	case cmdR53ZoneAdd.FullCommand():
		d := Data{}
		d.AddZones()

	// katoctl r53 record add:
	case cmdR53RecordAdd.FullCommand():
		d := Data{}
		d.AddRecords()

	// Nothing to do:
	default:
		return false
	}

	return true
}
