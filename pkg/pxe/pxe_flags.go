package pxe

//-----------------------------------------------------------------------------
// Package imports:
//-----------------------------------------------------------------------------

import "github.com/katosys/kato/pkg/cli"

//-----------------------------------------------------------------------------
// 'katoctl pxe' command flags definitions:
//-----------------------------------------------------------------------------

var (

	//------------------------
	// pxe: top level command
	//------------------------

	cmdPxe = cli.App.Command("pxe", "This is the Káto PXE provider.")

	//----------------------------
	// pxe deploy: nested command
	//----------------------------

	cmdPxeDeploy = cmdPxe.Command("deploy",
		"Deploy Káto's infrastructure on PXE clients.")
)

//-----------------------------------------------------------------------------
// RunCmd:
//-----------------------------------------------------------------------------

// RunCmd runs the cmd if owned by this package.
func RunCmd(cmd string) bool {

	switch cmd {

	// katoctl pxe deploy
	case cmdPxeDeploy.FullCommand():
		d := Data{}
		d.Deploy()

	// Nothing to do:
	default:
		return false
	}

	return true
}
