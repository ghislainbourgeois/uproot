module github.com/ghislainbourgeois/uproot

go 1.24.1

require (
	github.com/songgao/water v0.0.0-20200317203138-2b4b6d7c09d8
	github.com/vishvananda/netlink v1.3.0
	github.com/wmnsk/go-gtp v0.0.0-00010101000000-000000000000
	github.com/wmnsk/go-pfcp v0.0.24
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/vishvananda/netns v0.0.4 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
)

replace github.com/wmnsk/go-gtp => github.com/ghislainbourgeois/go-gtp v0.0.0-20250331154722-5a82f0657cfe
