package main

import (
	"net"
	"time"
)

type udpConnection struct {
	localConnection  net.PacketConn
	localAddress     net.Addr     // Address of the client
	remoteConnection *net.UDPConn // UDP connection to remote server
	remoteAddress    net.Addr
	closeChan        chan string
}

// Reply reads packets from remote server and forwards it on to the client connection
func (p *udpConnection) Reply() {
	buf := make([]byte, 32*1024) // THIS SHOULD BE CONFIGURABLE
	for {
		// Read from server
		p.remoteConnection.SetReadDeadline(time.Now().Add(time.Second * 36)) // timeout = 36s
		n, err := p.remoteConnection.Read(buf)
		if err != nil {
			p.closeChan <- p.localAddress.String()
			return
		}
		// Relay data from remote back to client
		_, err = p.localConnection.WriteTo(buf[0:n], p.localAddress)
		if err != nil {
			p.closeChan <- p.localAddress.String()
			return
		}
	}
}
