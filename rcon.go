package rcon

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"syscall"
)

// RCON implementation based on: https://developer.valvesoftware.com/wiki/Source_RCON_Protocol

const (
	DefaultPort = 27015
)

type RCON struct {
	host           string
	port           int
	password       string
	commandHandler func(command string, client Client)
	banList        []string
}

func NewRCON(host string, port int, password string) *RCON {
	return &RCON{
		host:     host,
		port:     port,
		password: password,
	}
}

func (r *RCON) SetBanList(banList []string) {
	r.banList = banList
}

func (r *RCON) OnCommand(handle func(command string, client Client)) {
	r.commandHandler = handle
}

func (r *RCON) addressInBanList(addr string) bool {
	for _, a := range r.banList {
		if a == addr {
			return true
		}
	}

	return false
}

func (r *RCON) ListenAndServe() error {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", r.host, r.port))
	if err != nil {
		return err
	}
	defer l.Close()

	log.Printf("starting RCON server on port %d\n", r.port)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("could not accept connection, error %v\n", err)
			continue
		}

		ip := addressWithoutPort(conn.RemoteAddr().String())

		if r.addressInBanList(ip) {
			log.Printf("address %s present in ban list, dropping\n", ip)
			_ = conn.Close()
			continue
		}

		log.Printf("new connection from %s\n", conn.RemoteAddr().String())
		go r.acceptConnection(conn)
	}
}

func (r *RCON) acceptConnection(conn net.Conn) {
	authenticated := false

	for {
		p, err := ParsePacket(conn)
		if errors.Is(err, io.EOF) ||
			errors.Is(err, syscall.WSAECONNRESET) {
			log.Printf("%s: connection closed\n", conn.RemoteAddr().String())
			_ = conn.Close()
			break
		}

		if err != nil {
			log.Printf("%s: could not read packet, %v\n", conn.RemoteAddr().String(), err)
			continue
		}

		// handle commands
		if authenticated && p.Type == ServerDataExecCommand {
			if r.commandHandler != nil {
				r.commandHandler(p.Body, NewClient(conn, p))
			}
			continue
		}

		// not authenticated and not a ServerDataAuth packet
		if p.Type != ServerDataAuth {
			log.Printf("%s: got wrong packet type, expected ServerDataAuth, got %s\n", conn.RemoteAddr().String(), p.Type.Stringer())
			_ = conn.Close()
			break
		}

		// empty password, we should refuse the connection
		if p.Type == ServerDataAuth && r.password == "" {
			log.Printf("%s: RCON password not set, refusing connection\n", conn.RemoteAddr().String())
			_ = conn.Close()
			break
		}

		// authentication
		if p.Type == ServerDataAuth {
			correct := p.Body == r.password
			id := int32(-1)

			if correct {
				id = p.ID
			}

			responsePacket := Packet{ID: id, Type: ServerDataAuthResponse}
			responseBytes, _ := EncodePacket(responsePacket)
			_, _ = conn.Write(responseBytes)

			if correct {
				authenticated = true
				log.Printf("%s: connection authenticated with password\n", conn.RemoteAddr().String())
			} else {
				log.Printf("%s: wrong password provided (%s)\n", p.Body, conn.RemoteAddr().String())
				_ = conn.Close()
				break
			}

			continue
		}

		panic("wrong implementation, should not reach this point")
	}
}

func addressWithoutPort(addr string) string {
	parts := strings.Split(addr, ":")
	return parts[0]
}
