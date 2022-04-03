module github.com/thingio/edge-randnum-driver

// replace (
// 	github.com/thingio/edge-device-driver v0.2.2 => ../edge-device-driver
// 	github.com/thingio/edge-device-std v0.2.2 => ../edge-device-std
// )

require (
	github.com/spf13/afero v1.8.0 // indirect
	github.com/spf13/viper v1.10.1 // indirect
	github.com/thingio/edge-device-driver v0.2.2
	github.com/thingio/edge-device-std v0.2.2
	golang.org/x/net v0.0.0-20220111093109-d55c255bac03 // indirect
	golang.org/x/sys v0.0.0-20220111092808-5a964db01320 // indirect
)

go 1.16
