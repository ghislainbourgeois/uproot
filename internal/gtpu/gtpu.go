// Copyright 2025 Ghislain Bourgeois
// SPDX-License-Identifier: GPL-3.0-or-later

package gtpu

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
	v1 "github.com/wmnsk/go-gtp/gtpv1"
)

const gtpuPort = 2152

var teidUPF uint32 = 0x000000001
var teidRAN uint32 = 0x000000010

type Tunnel struct {
	Name string
}

func NewTunnel(gnbIP string, upfIP string) (*Tunnel, error) {
	ctx := context.Background()
	laddr := &net.UDPAddr{
		IP:   net.ParseIP(gnbIP),
		Port: gtpuPort,
	}
	raddr := &net.UDPAddr{
		IP:   net.ParseIP(upfIP),
		Port: gtpuPort,
	}

	uConn, err := v1.DialUPlane(ctx, laddr, raddr)
	if err != nil {
		return nil, fmt.Errorf("could not connect to UPF: %v", err)
	}
	uConn.DisableErrorIndication()

	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = "uproot0"
	ifce, err := water.New(config)
	if err != nil {
		return nil, fmt.Errorf("could not open TUN interface: %v", err)
	}

	eth, err := netlink.LinkByName(ifce.Name())
	if err != nil {
		return nil, fmt.Errorf("cannot read TUN interface: %v", err)
	}

	ueAddr, err := netlink.ParseAddr("172.250.0.42/24")
	if err != nil {
		return nil, fmt.Errorf("could not parse UE address: %v", err)
	}

	err = netlink.AddrAdd(eth, ueAddr)
	if err != nil {
		return nil, fmt.Errorf("could not assign UE address to TUN interface: %v", err)
	}

	err = netlink.LinkSetUp(eth)
	if err != nil {
		return nil, fmt.Errorf("could not set TUN interface UP: %v", err)
	}

	go tunToGtp(uConn, ifce, raddr)
	go gtpToTun(uConn, ifce)

	return &Tunnel{Name: ifce.Name()}, nil
}

func (t *Tunnel) Close() error {
	return nil
}

func tunToGtp(uConn *v1.UPlaneConn, ifce *water.Interface, raddr *net.UDPAddr) {
	packet := make([]byte, 2000)
	for {
		n, err := ifce.Read(packet)
		if err != nil {
			log.Printf("error reading from tun interface: %v", err)
			continue
		}
		if n == 0 {
			log.Println("read 0 bytes")
			continue
		}
		_, err = uConn.WriteToGTP(teidRAN, packet[:n], raddr)
		if err != nil {
			log.Printf("error writing to GTP: %v", err)
			continue
		}
	}
}

func gtpToTun(uConn *v1.UPlaneConn, ifce *water.Interface) {
	packet := make([]byte, 2000)
	for {
		n, _, _, err := uConn.ReadFromGTP(packet)
		if err != nil {
			log.Printf("error reading from GTP: %v", err)
		}
		// if rteid != teidRAN {
		// 	log.Println("received packet for other tunnel: %v", rteid)
		// 	continue
		// }
		_, err = ifce.Write(packet[:n])
		if err != nil {
			log.Printf("error writing to tun interface: %v", err)
			continue
		}
	}
}
