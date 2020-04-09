package rcon

import (
	"io"
)

type Client interface {
	Write(content string) error
	Close() error
}

type client struct {
	conn       io.WriteCloser
	lastPacket Packet
}

func (c *client) Write(content string) error {
	responsePacket := Packet{ID: c.lastPacket.ID, Type: ServerDataResponseValue, Body: content}
	data, _ := EncodePacket(responsePacket)
	_, err := c.conn.Write(data)
	return err
}

func (c *client) Close() error {
	return c.conn.Close()
}

func NewClient(conn io.WriteCloser, lastPacket Packet) Client {
	return &client{
		conn:       conn,
		lastPacket: lastPacket,
	}
}
