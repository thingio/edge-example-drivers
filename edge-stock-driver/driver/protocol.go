package driver

import "github.com/thingio/edge-device-std/models"

var Protocol = &models.Protocol{
	ID:       "stock",
	Name:     "stock",
	Desc:     "Stock Price From Xueqiu",
	Category: "tick",
	Language: "ch",
	SupportFuncs: []string{
		"property",
		"event",
		"method",
	},
}
