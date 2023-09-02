// cmd/main.go

package main

import (
	"fmt"
	"net"
	"os"

	"github.com/sadityakumar9211/go-res/internal/dns"
)

func main() {
	// Bind UDP socket
	socket, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 2053})
	if err != nil {
		os.Exit(1)
	}
	defer socket.Close()

	// Handle DNS queries
	for {
		if err := dns.HandleQuery(socket); err != nil {
			fmt.Printf("An error occurred: %v\n", err)
		}
	}
}
