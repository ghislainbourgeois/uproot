// Copyright 2025 Ghislain Bourgeois
// SPDX-License-Identifier: GPL-3.0-or-later

package pfcp

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type Connection struct {
	localAddr *net.UDPAddr
	conn      net.Conn
	upf       upf
}

type upf struct {
	ip   string
	port int
}

type Response struct {
	from *net.UDPAddr
	msg  message.Message
}

func NewPFCPConnection(upfIP string, port int) (*Connection, error) {
	addr, err := findLocalIPAddr(upfIP, port)
	if err != nil {
		return nil, fmt.Errorf("cannot open UDP port for listening: %#v", err)
	}
	return &Connection{localAddr: addr, upf: upf{ip: upfIP, port: port}}, nil
}

func (c *Connection) Start() error {
	conn, err := net.ListenUDP("udp", c.localAddr)
	if err != nil {
		return fmt.Errorf("cannot open UDP port for listening: %#v", err)
	}
	c.conn = conn

	msgch := make(chan Response, 1)
	abort := make(chan bool)

	go listen(conn, msgch, abort)

	serverStartTime := time.Now()
	upfAddr := &net.UDPAddr{
		IP:   net.ParseIP(c.upf.ip),
		Port: c.upf.port,
	}

	associationReq := message.NewAssociationSetupRequest(
		1,
		ie.NewNodeIDHeuristic(c.localAddr.IP.String()),
		ie.NewRecoveryTimeStamp(serverStartTime),
		ie.NewCPFunctionFeatures(0),
	)

	err = sendPfcpMessage(conn, associationReq, upfAddr)
	if err != nil {
		return fmt.Errorf("could not send association setup request: %v", err)
	}

	resp := <-msgch
	if resp.msg.MessageType() != message.MsgTypeAssociationSetupResponse {
		return fmt.Errorf("did not recieve expected response, got %#v", resp)

	}
	associationResp, ok := resp.msg.(*message.AssociationSetupResponse)
	if !ok {
		return fmt.Errorf("did not recieve expected response, got %#v", resp)
	}
	if cause, err := associationResp.Cause.Cause(); err != nil || cause != ie.CauseRequestAccepted {
		return fmt.Errorf("association with UPF not accepted")
	}

	err = sendPfcpMessage(conn, buildPfcpSessionEstablishmentRequest(conn.LocalAddr().(*net.UDPAddr)), upfAddr)
	if err != nil {
		return fmt.Errorf("could not send session establishment request: %v", err)
	}
	resp = <-msgch
	if resp.msg.MessageType() != message.MsgTypeSessionEstablishmentResponse {
		return fmt.Errorf("did not recieve expected response, got %#v", resp)

	}
	sessResp, ok := resp.msg.(*message.SessionEstablishmentResponse)
	if !ok {
		return fmt.Errorf("did not recieve expected response, got %#v", resp)
	}
	if cause, err := sessResp.Cause.Cause(); err != nil || cause != ie.CauseRequestAccepted {
		return fmt.Errorf("session establishmed not accepted")
	}

	go func() {
		var seq uint32
		for {
			msg := message.NewHeartbeatRequest(seq, ie.NewRecoveryTimeStamp(serverStartTime), nil)
			err = sendPfcpMessage(conn, msg, upfAddr)
			if err != nil {
				log.Printf("could not send heartbeat request: %v", err)
			}
			resp = <-msgch
			if resp.msg.MessageType() != message.MsgTypeHeartbeatResponse {
				log.Printf("did not received expected heartbeat response, got %#v", resp)
			}
			time.Sleep(2 * time.Second)
		}
	}()

	return nil
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func findLocalIPAddr(dest string, port int) (*net.UDPAddr, error) {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", dest, port))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	addr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return nil, err
	}

	return addr, nil
}

func listen(conn *net.UDPConn, msgch chan Response, abort chan bool) {
	for {
		select {
		case <-abort:
			close(msgch)
			break
		default:
			addr, msg, err := readPfcpMessage(conn)
			if err != nil {
				log.Println(err)
				continue
			}
			msgch <- Response{from: addr, msg: msg}
		}
	}
}

func readPfcpMessage(conn *net.UDPConn) (*net.UDPAddr, message.Message, error) {
	if conn == nil {
		return nil, nil, fmt.Errorf("server not opened")
	}

	buf := make([]byte, 2048)
	conn.SetReadDeadline(time.Now().Add(1 * time.Minute))
	n, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return addr, nil, err
	}

	msg, err := message.Parse(buf[:n])
	if err != nil {
		return addr, nil, err
	}
	return addr, msg, nil
}

func sendPfcpMessage(conn *net.UDPConn, msg message.Message, addr *net.UDPAddr) error {
	if conn == nil {
		return fmt.Errorf("server not opened")
	}

	buf := make([]byte, msg.MarshalLen())
	err := msg.MarshalTo(buf)
	if err != nil {
		return err
	}
	_, err = conn.WriteToUDP(buf, addr)
	return err
}
