package main

import (
	"github.com/thingio/edge-device-driver/pkg/startup"
	"github.com/thingio/edge-stock-driver/driver"
)

func main() {
	startup.Startup(driver.Protocol, driver.NewStockTwin)
}
