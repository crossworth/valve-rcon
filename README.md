## Valve RCON server implementation

Simple server implementation for the [RCON protocol](https://developer.valvesoftware.com/wiki/Source_RCON_Protocol).


**Example usage**
```go
package main

import (
	"log"

	rcon "github.com/crossworth/valve-rcon"
)

const (
	host     = "0.0.0.0"
	port     = rcon.DefaultPort
	password = "test"
)

func main() {
	server := rcon.NewRCON(host, port, password)
	server.SetBanList([]string{
		"192.168.0.10",
		// ...
	})

	// echo server
	server.OnCommand(func(command string, client rcon.Client) {
		log.Printf("command: %s", command)
		_ = client.Write("server: " + command)
	})

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}
}
```
