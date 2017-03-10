package cli

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"regexp"
	"strconv"
	"strings"

	// Community:
	log "github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

//-----------------------------------------------------------------------------
// Package factored var statement:
//-----------------------------------------------------------------------------

var (

	// App contains flags, arguments and commands for an application:
	App = kingpin.New("katoctl", "Katoctl defines and deploys Kato's infrastructure.")

	// KatoRoles is a slice of valid Káto roles:
	KatoRoles = []string{"quorum", "master", "worker", "border"}
)

//----------------------------------------------------------------------------
// func init() is called after all the variable declarations in the package
// have evaluated their initializers, and those are evaluated only after all
// the imported packages have been initialized:
//----------------------------------------------------------------------------

func init() {

	// Customize kingpin:
	App.Version("v0.1.0").Author("Marc Villacorta Morera")
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

//-----------------------------------------------------------------------------
// Quadruplets custom parser:
//-----------------------------------------------------------------------------

type quadrupletsValue struct {
	quadList []string
	types    []string
	roles    []string
}

func (q *quadrupletsValue) Set(value string) error {

	// 1. Four elements:
	if quad := strings.Split(value, ":"); len(quad) != 4 {
		log.WithField("value", value).
			Fatal("Expected 4 elements, but got " + strconv.Itoa(len(quad)))

		// 2. Positive integer:
	} else if i, err := strconv.Atoi(quad[0]); err != nil || i < 0 {
		log.WithField("value", value).
			Fatal("First quadruplet element must be a positive integer, but got: " + quad[0])

		// 3. Valid instance type:
	} else if !func() bool {
		for _, t := range q.types {
			if t == quad[1] {
				return true
			}
		}
		return false
	}() {
		log.WithField("value", value).
			Fatal("Second quadruplet element must be a valid instance type, but got: " + quad[1])

		// 4. Valid DNS name:
	} else if match, err := regexp.MatchString("^[a-z\\d-]+$", quad[2]); err != nil || !match {
		log.WithField("value", value).
			Fatal("Third quadruplet element must matmatch ^[a-z\\d-]+$, but got: " + quad[2])

		// 5. Valid Káto roles:
	} else if !func() bool {
		for _, role := range strings.Split(quad[3], ",") {
			found := false
			for _, r := range q.roles {
				if r == role {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}() {
		log.WithField("value", value).
			Fatal("Fourth quadruplet element must be a valid list of Káto roles, but got: " + quad[3])
	}

	// All tests ok:
	q.quadList = append(q.quadList, value)
	return nil
}

func (q *quadrupletsValue) String() string {
	return ""
}

func (q *quadrupletsValue) IsCumulative() bool {
	return true
}

// Quadruplets is a custom parser:
func Quadruplets(s kingpin.Settings, types, roles []string) *[]string {
	target := &quadrupletsValue{}
	target.types = types
	target.roles = roles
	s.SetValue(target)
	return &target.quadList
}
