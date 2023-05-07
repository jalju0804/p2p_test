package main

import (
	"fmt"
	"net"
	"time"

	"github.com/pion/stun"
)

// getPublicAddress retrieves the public IP address and port using a STUN server.
func getPublicAddress(stunServer string) (net.Addr, error) {
	conn, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	raddr, err := net.ResolveUDPAddr("udp4", stunServer)
	if err != nil {
		return nil, err
	}

	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
	_, err = conn.WriteTo(message.Raw, raddr)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 1024)
	n, _, err := conn.ReadFrom(buf)
	if err != nil {
		return nil, err
	}

	var res stun.Message
	if err := res.UnmarshalBinary(buf[:n]); err != nil {
		return nil, err
	}

	var xorMappedAddress stun.XORMappedAddress
	if err := xorMappedAddress.GetFrom(&res); err != nil {
		return nil, err
	}

	return &net.UDPAddr{
		IP:   xorMappedAddress.IP,
		Port: xorMappedAddress.Port,
	}, nil
}

func main() {
	stunServer := "stun.l.google.com:19302"

	// Get public address from STUN server
	publicAddr, err := getPublicAddress(stunServer)
	if err != nil {
		fmt.Println("Failed to get public address:", err)
		return
	}
	fmt.Println("Public address:", publicAddr)

	// Start a TCP listener
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Println("Failed to start TCP listener:", err)
		return
	}
	defer listener.Close()

	// Share the public address and listener port with the other peer
	// (e.g., through a signaling server, manual input, etc.)
	localAddr := listener.Addr().(*net.TCPAddr)
	fmt.Printf("Share this address with the other peer: %s:%d\n", publicAddr.(*net.UDPAddr).IP, localAddr.Port)

	// Wait for a connection or connect to the other peer
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			return
		}
		defer conn.Close()
		fmt.Println("Connected to:", conn.RemoteAddr())
	}()

	// Connect to the other peer using its public address and port
	// (this should be replaced with the actual address and port shared by the other peer)
	fmt.Println("Enter the other peer's public address and port:")
	var otherPeerAddr string
	fmt.Scanln(&otherPeerAddr)

	conn, err := net.DialTimeout("tcp", otherPeerAddr, 5*time.Second)
	if err != nil {
		fmt.Println("Failed to connect to the other peer:", err)
		return
	}
	defer conn.Close()
	fmt.Println("Connected to:", conn.RemoteAddr())

	// Read and write data to the connection
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				fmt.Println("Failed to read data:", err)
				return
			}
			fmt.Printf("Received: %s\n", string(buf[:n]))
		}
	}()

	go func() {
		for {
			var message string
			fmt.Print("Enter a message: ")
			fmt.Scanln(&message)

			_, err := conn.Write([]byte(message))
			if err != nil {
				fmt.Println("Failed to send data:", err)
				return
			}
		}
	}()

	// Wait indefinitely
	select {}
}
