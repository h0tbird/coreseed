package cli

//-----------------------------------------------------------------------------
// Import:
//-----------------------------------------------------------------------------

import kingpin "gopkg.in/alecthomas/kingpin.v2"

//-----------------------------------------------------------------------------
// katoctl root level command:
//-----------------------------------------------------------------------------

// App contains flags, arguments and commands for an application:
var App = kingpin.New("katoctl", "Katoctl defines and deploys Kato's infrastructure.")

//----------------------------------------------------------------------------
// func init() is called after all the variable declarations in the package
// have evaluated their initializers, and those are evaluated only after all
// the imported packages have been initialized:
//----------------------------------------------------------------------------

func init() {

	// Customize kingpin:
	App.Version("0.1.0").Author("Marc Villacorta Morera")
	App.UsageTemplate(usageTemplate)
	App.HelpFlag.Short('h')
}
