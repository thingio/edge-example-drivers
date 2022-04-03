package driver

import "github.com/thingio/edge-device-std/models"

var Protocol = &models.Protocol{
	ID:       "randnum",
	Name:     "randnum",
	Desc:     "Random Number Simulator",
	Category: "tick",
	Language: "ch",
	SupportFuncs: []string{
		"property",
		"event",
		"method",
	},
}
