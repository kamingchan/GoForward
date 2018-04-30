package main

import (
	"net"
	"fmt"
)

type Server struct {
	localAddress    string
	destAddress     string
	port            int
	tcpListener     net.Listener
	udpListener     net.PacketConn
	udpClients      map[string]*udpConnection
	udpClientClosed chan string
}

func NewServer(l string, d string, p int) *Server {
	return &Server{
		localAddress: l,
		destAddress:  d,
		port:         p,
		udpClients:   make(map[string]*udpConnection),
	}
}

func (s *Server) Start() {
	// TCP forward
	go func() error {
		var l, err = net.Listen("tcp", s.localAddress)
		if err != nil {
			// TODO: Log
			return err
		}
		s.tcpListener = l
		for {
			connection, err := s.tcpListener.Accept()
			if err != nil {
				return err
			}

			p := &tcpConnection{
				localConnection: connection,
				localAddress:    s.localAddress,
				remoteAddress:   s.destAddress,
				port:            s.port,
			}

			go p.forward()
		}
	}()
	// UDP relay
	go func() error {
		var l, err = net.ListenPacket("udp", s.localAddress)
		if err != nil {
			// TODO: Log
			return err
		}
		raddr, err := net.ResolveUDPAddr("udp", s.destAddress)
		if err != nil {
			return err
		}
		s.udpListener = l
		s.udpClientClosed = make(chan string)

		// handle closed udp connection
		go func() {
			for {
				clientAddr := <-s.udpClientClosed

				conn, found := s.udpClients[clientAddr]
				if found {
					fmt.Printf("[%d] timeout UDP connection from %s via %s\n", s.port, clientAddr, conn.remoteAddress.String())
					conn.remoteConnection.Close()
					delete(s.udpClients, clientAddr)
				}
			}
		}()

		for {
			buf := make([]byte, 4096)
			size, address, err := s.udpListener.ReadFrom(buf)
			if err != nil {
				s.udpListener.Close()
			}

			connection, found := s.udpClients[address.String()]
			if !found {
				remoteUDPConn, err := net.DialUDP("udp", nil, raddr)
				if err != nil {
					return err
				}

				connection = &udpConnection{
					localConnection:  s.udpListener,
					localAddress:     address,
					remoteConnection: remoteUDPConn,
					remoteAddress:    remoteUDPConn.LocalAddr(),
					closeChan:        s.udpClientClosed,
				}

				s.udpClients[address.String()] = connection

				fmt.Printf("[%d] new UDP connection from %s via %s\n", s.port, address.String(), remoteUDPConn.LocalAddr().String())

				// wait for data from remote server
				go connection.Reply()
			}

			go func() {
				// forward data received to remote server
				_, err = connection.remoteConnection.Write(buf[0:size])
				if err != nil {
					return
				}
			}()
		}
	}()
}
