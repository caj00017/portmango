package main

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"time"

	"github.com/mdlayher/arp"
)

func main() {
	// target the network interface
	iface, err := net.InterfaceByName("enp42s0") // handy net method for grabbing an interface
	if err != nil {
		fmt.Printf("Cannot get interface: %v", err) // %v is the format code used to print an error's message text
	}

	// create a new client - I think I can use the Dial function here for it
	client, err := arp.Dial(iface) // Dial returns a pointer to the new client
	if err != nil {
		fmt.Printf("Cannot get client: %v", err)
	}
	defer func() {
		closeErr := client.Close()
		if closeErr != nil {
			fmt.Printf("Cannot close client: %v", closeErr)
		}
	}()

	// tell the client to send a Request for a specific IP address. I'll use a hardcoded one to start.
	target_ip, err := netip.ParseAddr("10.0.0.228") // parseAddr returns an Addr, which I can pass to Request
	if err != nil {
		fmt.Printf("Cannot get target IP: %v", err)
	}

	// create Request
	err = client.Request(target_ip)
	if err != nil {
		fmt.Printf("Cannot make Request: %v", err)
	}

	for {

		err = client.SetReadDeadline(time.Now().Add(5 * time.Second)) // set Read deadline
		if err != nil {
			fmt.Printf("Cannot set Read deadline: %v", err)
		}

		// use the Read method to listen for the specific Reply that contains the matching MAC address
		packet, _, err := client.Read() // removing the frame because i'm not sure I need it
		if err != nil {                 // error can be generic or it could be the timeout
			if errors.Is(err, os.ErrDeadlineExceeded) {
				fmt.Println("Timeout: No reply received from target.")
			} else {
				// generic error
				fmt.Printf("Cannot Read from Client: %v", err)
			}
		}

		// if the operation of the packet is a reply, examine it
		if packet.Operation == 2 && packet.SenderIP == target_ip { // 2 for reply, 1 for request

			fmt.Println("MAC Address Found!")

			fmt.Printf("Desired IP: %s\n", packet.SenderIP)
			fmt.Printf("Found MAC: %s\n", packet.SenderHardwareAddr)
			break
		}
	}
}
