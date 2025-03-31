UProot
======

UProot is a tool that enables testing the throughput of 5G User Plane Functions
in isolation.

It acts as both the Session Management Function (SMF) and a combined gNodeB and
User Equipment (UE), implementing minimal 3GPP protocols for the task at hand.

Compared to more complete testing tools, UProot does not require a core network
and can enable tests of a UPF without other supporting software.

***This is pre-alpha software. It is expected to change a lot. Use at your own risk.***

Installation
------------

Clone this repository and build uproot:

```
go build uproot.go
```

Create a configuration file for your environment:

```yaml
upfIP: 10.5.5.200       # IP address that the UPF listens on for PFCP
pfcpPort: 8805          # Port that the UPF listens on for PFCP
upfN3IP: 10.202.0.10    # IP address of the access interface of the UPF (N3)
gnbIP: 10.204.0.42      # IP address of the machine running uproot to use for N3 communications
```

## License

GNU General Public License v3.0 or later

See [LICENSE](LICENSE) to see the full text.
