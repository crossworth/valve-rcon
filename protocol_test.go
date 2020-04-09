package rcon

import (
	"bytes"
	"testing"
)

var (
	authPacket = []byte{0x11, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x70, 0x61, 0x73, 0x73, 0x77, 0x72, 0x64, 0x00, 0x00}
)

func Test_parsePacket(t *testing.T) {
	t.Run("authPacket", func(t *testing.T) {
		p, err := ParsePacket(bytes.NewReader(authPacket))
		if err != nil {
			t.Fatal(err)
		}

		if p.Size != 17 {
			t.Fatalf("expected size to be 17 bytes, got %d", p.Size)
		}

		if p.ID != 0 {
			t.Fatalf("expected id to be 0, got %d", p.ID)
		}

		if p.Type != ServerDataAuth {
			t.Fatalf("expected type to be ServerDataAuth, got %s", p.Type.Stringer())
		}

		if p.Body != "passwrd" {
			t.Fatalf("expected body to be passwrd, got %q", p.Body)
		}
	})
}

func Test_encodePacket(t *testing.T) {
	t.Run("authPacket", func(t *testing.T) {
		p := Packet{
			ID:   0,
			Type: ServerDataAuth,
			Body: "passwrd",
		}

		pBytes, err := EncodePacket(p)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(pBytes, authPacket) {
			t.Fatalf("could not encode packet, wrong bytes\n%q\n%q\n", pBytes, authPacket)
		}
	})
}
