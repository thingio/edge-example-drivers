module github.com/thingio/edge-stock-driver

replace (
	github.com/thingio/edge-device-driver v0.2.2 => ../../edge-device-driver
	github.com/thingio/edge-device-std v0.2.2 => ../../edge-device-std
)

require (
	github.com/thingio/edge-device-driver v0.2.2
	github.com/thingio/edge-device-std v0.2.2
)

go 1.16
