package rcon

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

const MaxPacketSize = 4096

// PacketType represents an RCON packet type
type PacketType int32

const (
	ServerDataAuth          = PacketType(3) // Authentication packet
	ServerDataAuthResponse  = PacketType(2) // Authentication response packet
	ServerDataExecCommand   = PacketType(2) // Command packet
	ServerDataResponseValue = PacketType(0) // Response packet
)

func (p PacketType) Stringer() string {
	switch p {
	case ServerDataAuth:
		return "ServerDataAuth (3)"
	case ServerDataAuthResponse: // is the same as ServerDataExecCommand
		return "ServerDataAuthResponse|ServerDataExecCommand (2)"
	case ServerDataResponseValue:
		return "ServerDataResponseValue (0)"
	}

	return fmt.Sprintf("unknown (%d)", int(p))
}

// Packet represents an RCON packet
type Packet struct {
	Size int32
	ID   int32
	Type PacketType
	Body string
}

// ParsePacket return the packet parsed from the input
// or an error if the packet could not be parsed
func ParsePacket(input io.Reader) (Packet, error) {
	var p Packet
	reader := bufio.NewReader(input)

	err := binary.Read(reader, binary.LittleEndian, &p.Size)
	if err != nil {
		return p, errors.Wrap(err, "could not read packet size")
	}

	err = binary.Read(reader, binary.LittleEndian, &p.ID)
	if err != nil {
		return p, errors.Wrap(err, "could not read packet id")
	}

	err = binary.Read(reader, binary.LittleEndian, &p.Type)
	if err != nil {
		return p, errors.Wrap(err, "could not read packet type")
	}

	// subtract the id and type that we already read
	bodyLen := int(p.Size - 8)
	bodyBuf := make([]byte, bodyLen)
	err = binary.Read(reader, binary.LittleEndian, &bodyBuf)
	if err != nil {
		return p, errors.Wrap(err, "could not read packet body")
	}

	// we remove the empty string and the last null terminated ascii
	p.Body = string(bodyBuf[:bodyLen-2])
	return p, nil
}

// EncodePacket convert the packet to a binary representation
// expected by the RCON protocol, it will return an error
// if the packet exceeds the maximum packet size.
//
// if you need to send a lot of data
// you should send multiple packets
// https://developer.valvesoftware.com/wiki/Source_RCON_Protocol#Multiple-packet_Responses
func EncodePacket(p Packet) ([]byte, error) {
	var output bytes.Buffer
	p.Size = int32(4 + 4 + len(p.Body) + 1 + 1)
	_ = binary.Write(&output, binary.LittleEndian, p.Size)
	_ = binary.Write(&output, binary.LittleEndian, p.ID)
	_ = binary.Write(&output, binary.LittleEndian, p.Type)
	_ = binary.Write(&output, binary.LittleEndian, []byte(p.Body))
	_ = binary.Write(&output, binary.LittleEndian, byte(0x0))
	_ = binary.Write(&output, binary.LittleEndian, byte(0x0))

	if len(output.Bytes()) > MaxPacketSize {
		return output.Bytes(), errors.New(fmt.Sprintf("the packet exceeds the maximum size of %d, packet size %d", MaxPacketSize, len(output.Bytes())))
	}

	return output.Bytes(), nil
}
