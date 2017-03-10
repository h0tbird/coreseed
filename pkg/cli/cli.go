package cli

//-----------------------------------------------------------------------------
// Import:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"regexp"

	// Community:
	log "github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

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

//-----------------------------------------------------------------------------
// Regular expression custom parser:
//-----------------------------------------------------------------------------

type regexpMatchValue struct {
	value  string
	regexp string
}

func (id *regexpMatchValue) Set(value string) error {

	if match, _ := regexp.MatchString(id.regexp, value); !match {
		log.WithField("value", value).Fatal("Value must match: " + id.regexp)
	}

	id.value = value
	return nil
}

func (id *regexpMatchValue) String() string {
	return id.value
}

// RegexpMatch is a regular expression custom parser.
func RegexpMatch(s kingpin.Settings, regexp string) *string {
	target := &regexpMatchValue{}
	target.regexp = regexp
	s.SetValue(target)
	return &target.value
}
