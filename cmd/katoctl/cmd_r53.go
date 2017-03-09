package katoctl

//-----------------------------------------------------------------------------
// 'katoctl r53' command flags definitions:
//-----------------------------------------------------------------------------

var (

	//------------------------
	// r53: top level command
	//------------------------

	cmdR53 = app.Command("r53", "Manages Route 53 zones and records.")
)
