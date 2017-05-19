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

	flPxeDeployDNSProvider = cmdPxeDeploy.Flag("dns-provider",
		"DNS provider [ none | ns1 | r53 ]").
		Default("r53").PlaceHolder("KATO_PXE_DEPLOY_DNS_PROVIDER").
		OverrideDefaultFromEnvar("KATO_PXE_DEPLOY_DNS_PROVIDER").
		Enum("none", "ns1", "r53")

	flPxeDeployDNSApiKey = cmdPxeDeploy.Flag("dns-api-key",
		"DNS private API key.").
		PlaceHolder("KATO_PXE_DEPLOY_DNS_API_KEY").
		OverrideDefaultFromEnvar("KATO_PXE_DEPLOY_DNS_API_KEY").
		String()

	flPxeDeployDomain = cmdPxeDeploy.Flag("domain",
		"Used to identify the VPC.").
		Required().PlaceHolder("KATO_PXE_DEPLOY_DOMAIN").
		OverrideDefaultFromEnvar("KATO_PXE_DEPLOY_DOMAIN").
		String()

	flPxeDeployEtcdToken = cmdPxeDeploy.Flag("etcd-token",
		"Etcd bootstrap token [ auto | <token> ]").
		Default("auto").OverrideDefaultFromEnvar("KATO_PXE_DEPLOY_ETCD_TOKEN").
		HintOptions("auto").String()

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
				DNSProvider: *flPxeDeployDNSProvider,
				DNSApiKey:   *flPxeDeployDNSApiKey,
				Domain:      *flPxeDeployDomain,
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
