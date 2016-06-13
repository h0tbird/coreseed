package main

//-----------------------------------------------------------------------------
// 'katoctl pkt' command flags definitions:
//-----------------------------------------------------------------------------

var (

	//------------------------
	// pkt: top level command
	//------------------------

	cmdPkt = app.Command("pkt", "Kato's Packet.net provider.")

	//----------------------------
	// pkt deploy: nested command
	//----------------------------

	cmdPktDeploy = cmdPkt.Command("deploy", "Deploy Kato's infrastructure on Packet.net")

	//---------------------------
	// pkt setup: nested command
	//---------------------------

	cmdPktSetup = cmdPkt.Command("setup", "Setup a Packet.net project to be used by katoctl.")

	//-------------------------
	// pkt run: nested command
	//-------------------------

	cmdPktRun = cmdPkt.Command("run", "Starts a CoreOS instance on Packet.net.")

	flPktRunAPIKey = cmdPktRun.Flag("api-key", "Packet API key.").
			Required().PlaceHolder("KATO_RUN_PKT_APIKEY").
			OverrideDefaultFromEnvar("KATO_RUN_PKT_APIKEY").
			Short('k').String()

	flPktRunHostname = cmdPktRun.Flag("hostname", "Used in the Packet.net dashboard.").
				Required().PlaceHolder("KATO_RUN_PKT_HOSTNAME").
				OverrideDefaultFromEnvar("KATO_RUN_PKT_HOSTNAME").
				Short('h').String()

	flPktRunProjectID = cmdPktRun.Flag("project-id", "Format: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee").
				Required().PlaceHolder("KATO_RUN_PKT_PROJECT_ID").
				OverrideDefaultFromEnvar("KATO_RUN_PKT_PROJECT_ID").
				Short('i').String()

	flPktRunPlan = cmdPktRun.Flag("plan", "One of [ baremetal_0 | baremetal_1 | baremetal_2 | baremetal_3 ]").
			Required().PlaceHolder("KATO_RUN_PKT_PLAN").
			OverrideDefaultFromEnvar("KATO_RUN_PKT_PLAN").
			Short('p').Enum("baremetal_0", "baremetal_1", "baremetal_2", "baremetal_3")

	flPktRunOS = cmdPktRun.Flag("os", "One of [ coreos_stable | coreos_beta | coreos_alpha ]").
			Default("coreos_stable").OverrideDefaultFromEnvar("KATO_RUN_PKT_OS").
			Short('o').Enum("coreos_stable", "coreos_beta", "coreos_alpha")

	flPktRunFacility = cmdPktRun.Flag("facility", "One of [ ewr1 | ams1 ]").
				Required().PlaceHolder("KATO_RUN_PKT_FACILITY").
				OverrideDefaultFromEnvar("KATO_RUN_PKT_FACILITY").
				Short('f').Enum("ewr1", "ams1")

	flPktRunBilling = cmdPktRun.Flag("billing", "One of [ hourly | monthly ]").
			Default("hourly").OverrideDefaultFromEnvar("KATO_RUN_PKT_BILLING").
			Short('b').Enum("hourly", "monthly")
)
