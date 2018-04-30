package main

import (
	"fmt"
	"net"
)

// tcpConnection resembles a forward connection and pipe data between local and remote.
type tcpConnection struct {
	localAddress, remoteAddress       string
	localConnection, remoteConnection net.Conn
	port                              int
}

// forward establishes the connection to the remote server and
// starts data exchange. It will block until a close signal is received
// so it's advisable to call as a goroutine
func (p *tcpConnection) forward() {
	defer p.localConnection.Close()
	var err error

	p.remoteConnection, err = net.Dial("tcp", p.remoteAddress)
	if err != nil {
		// TODO: log
		//p.errorFunc("Cannot connect to remote connection: %s", err)
		return
	}
	defer p.remoteConnection.Close()

	fmt.Printf("[%d] new TCP connection from %s via %s\n", p.port, p.localConnection.RemoteAddr(), p.remoteConnection.LocalAddr())

	var closed = make(chan bool)
	go p.exchangeData(p.remoteConnection, p.localConnection, closed)
	go p.exchangeData(p.localConnection, p.remoteConnection, closed)

	<-closed
	<-closed

	fmt.Printf("[%d] close TCP connection from %s via %s\n", p.port, p.localConnection.RemoteAddr(), p.remoteConnection.LocalAddr())
}

// exchangeData reads from source connection and forwards
// data to destination connection
func (p *tcpConnection) exchangeData(dst, src net.Conn, close chan bool) {
	buf := make([]byte, 512*1024) // 512kb
	for {
		bytesRead, err := src.Read(buf)
		if err != nil {
			// TODO: log
			//p.errorFunc(fmt.Sprintf("Error reading from client connection. src=%s %s dst=%s %s", src.LocalAddr(), src.RemoteAddr(), dst.LocalAddr(), dst.RemoteAddr()), err)
			close <- true
			return
		}

		if bytesRead > 0 {
			b := buf[:bytesRead]
			_, err = dst.Write(b)
			if err != nil {
				// TODO: log
				//p.errorFunc("Cannot write to remote connection", err)
				close <- true
				return
			}
		}
	}
}
