// Copyright 2025 Ghislain Bourgeois
// SPDX-License-Identifier: GPL-3.0-or-later

package pfcp

import (
	"math"
	"net"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

const (
	InterfaceAccess uint8 = iota
	InterfaceCore
)

const gtpuUdpIpv4 uint16 = 0x0100

var uplinkPDRID uint16 = 0
var downlinkPDRID uint16 = 1
var uplinkFARID uint32 = 10
var downlinkFARID uint32 = 11
var uplinkQERID uint32 = 20
var downlinkQERID uint32 = 21
var teidUPF uint32 = 0x000000010
var teidRAN uint32 = 0x000000001

func buildPfcpSessionEstablishmentRequest(nodeID *net.UDPAddr) *message.SessionEstablishmentRequest {
	ies := make([]*ie.IE, 0)
	ies = append(ies, ie.NewNodeIDHeuristic(nodeID.IP.String()))
	ies = append(ies, ie.NewFSEID(0, nodeID.IP, nil))

	ies = append(ies, buildUplinkPDR())
	ies = append(ies, buildDownlinkPDR())

	ies = append(ies, buildUplinkFAR())
	ies = append(ies, buildDownlinkFAR())

	ies = append(ies, buildUplinkQER())
	ies = append(ies, buildDownlinkQER())

	ies = append(ies, ie.NewPDNType(ie.PDNTypeIPv4))

	return message.NewSessionEstablishmentRequest(1, 0, 0, 0, 0, ies...)
}

func buildUplinkPDR() *ie.IE {
	accessIP := net.ParseIP("10.202.0.10")
	ies := make([]*ie.IE, 0)
	ies = append(ies, ie.NewPDRID(uplinkPDRID))
	ies = append(ies, ie.NewPrecedence(math.MaxUint32))
	pdiies := make([]*ie.IE, 0)
	pdiies = append(pdiies, ie.NewSourceInterface(InterfaceAccess))
	pdiies = append(pdiies, ie.NewFTEID(1, teidUPF, accessIP, nil, 0))
	pdiies = append(pdiies, ie.NewNetworkInstance("internet"))
	pdiies = append(pdiies, ie.NewUEIPAddress(2, "172.250.0.42", "", 0, 0))
	ies = append(ies, ie.NewPDI(pdiies...))
	ies = append(ies, ie.NewOuterHeaderRemoval(0, 0))
	ies = append(ies, ie.NewFARID(uplinkFARID))
	ies = append(ies, ie.NewQERID(uplinkQERID))
	return ie.NewCreatePDR(ies...)
}

func buildDownlinkPDR() *ie.IE {
	ies := make([]*ie.IE, 0)
	ies = append(ies, ie.NewPDRID(downlinkPDRID))
	ies = append(ies, ie.NewPrecedence(math.MaxUint32))
	pdiies := make([]*ie.IE, 0)
	pdiies = append(pdiies, ie.NewSourceInterface(InterfaceCore))
	pdiies = append(pdiies, ie.NewNetworkInstance("internet"))
	pdiies = append(pdiies, ie.NewUEIPAddress(2, "172.250.0.42", "", 0, 0))
	ies = append(ies, ie.NewPDI(pdiies...))
	ies = append(ies, ie.NewFARID(downlinkFARID))
	ies = append(ies, ie.NewQERID(downlinkQERID))
	return ie.NewCreatePDR(ies...)
}

func buildUplinkFAR() *ie.IE {
	faries := make([]*ie.IE, 0)
	faries = append(faries, ie.NewFARID(uplinkFARID))
	faries = append(faries, ie.NewApplyAction(0x02)) // Forward
	fwdies := make([]*ie.IE, 0)
	fwdies = append(fwdies, ie.NewDestinationInterface(InterfaceCore))
	fwdies = append(fwdies, ie.NewOuterHeaderRemoval(0, 0)) // Remove gtpuudpipv4 header
	fwdies = append(fwdies, ie.NewNetworkInstance("internet"))
	faries = append(faries, ie.NewForwardingParameters(fwdies...))
	return ie.NewCreateFAR(faries...)
}

func buildDownlinkFAR() *ie.IE {
	faries := make([]*ie.IE, 0)
	faries = append(faries, ie.NewFARID(downlinkFARID))
	faries = append(faries, ie.NewApplyAction(0x02)) // Forward
	fwdies := make([]*ie.IE, 0)
	fwdies = append(fwdies, ie.NewDestinationInterface(InterfaceAccess))
	fwdies = append(fwdies, ie.NewNetworkInstance("internet"))
	fwdies = append(fwdies, ie.NewOuterHeaderCreation(gtpuUdpIpv4, teidRAN, "10.204.0.42", "", 0, 0, 0))
	faries = append(faries, ie.NewForwardingParameters(fwdies...))
	return ie.NewCreateFAR(faries...)
}

func buildUplinkQER() *ie.IE {
	qeries := make([]*ie.IE, 0)
	qeries = append(qeries, ie.NewQERID(uplinkQERID))
	qeries = append(qeries, ie.NewGateStatus(ie.GateStatusOpen, ie.GateStatusOpen))
	qeries = append(qeries, ie.NewQFI(1))
	qeries = append(qeries, ie.NewMBR(math.MaxUint64, math.MaxUint64))
	qeries = append(qeries, ie.NewGBR(math.MaxUint64, math.MaxUint64))
	return ie.NewCreateQER(qeries...)
}

func buildDownlinkQER() *ie.IE {
	qeries := make([]*ie.IE, 0)
	qeries = append(qeries, ie.NewQERID(downlinkQERID))
	qeries = append(qeries, ie.NewGateStatus(ie.GateStatusOpen, ie.GateStatusOpen))
	qeries = append(qeries, ie.NewQFI(1))
	qeries = append(qeries, ie.NewMBR(math.MaxUint64, math.MaxUint64))
	qeries = append(qeries, ie.NewGBR(math.MaxUint64, math.MaxUint64))
	return ie.NewCreateQER(qeries...)
}
