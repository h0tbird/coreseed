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

	arPxeDeployQuadruplet = cli.Quadruplets(cmdPxeDeploy.Arg("quadruplet",
		"<number_of_instances>::<host_name>:<comma_separated_list_of_roles>").
		Required(), []string{""}, cli.KatoRoles)
)

//-----------------------------------------------------------------------------
// RunCmd:
//-----------------------------------------------------------------------------

// RunCmd runs the cmd if owned by this package.
func RunCmd(cmd string) bool {

	switch cmd {

	// katoctl pxe deploy
	case cmdPxeDeploy.FullCommand():
		d := Data{
			State: State{
				Quadruplets: *arPxeDeployQuadruplet,
			},
		}
		d.Deploy()

	// Nothing to do:
	default:
		return false
	}

	return true
}
