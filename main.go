package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"

	libp2p "github.com/libp2p/go-libp2p"
	circuit "github.com/libp2p/go-libp2p-circuit"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pion/stun"
)

func handleStream(stream network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw)
	go writeData(rw)

	select {} // hang forever
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			panic(err)
		}
		if len(str) > 0 {
			fmt.Printf("%s", str)
		}
	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			panic(err)
		}
	}
}

func getPublicAddress(stunServerAddr string) (net.Addr, error) {
	conn, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	raddr, err := net.ResolveUDPAddr("udp4", stunServerAddr)
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
	ctx := context.Background()

	node, err := libp2p.New(ctx,
		libp2p.EnableNATService(),
		libp2p.EnableRelay(circuit.RelayOpt(circuit.Hop)),
		libp2p.NATPortMap(),
	)
	if err != nil {
		panic(err)
	}

	node.SetStreamHandler("/chat/1.0.0", handleStream)

	stunServerAddr := "stun.l.google.com:19302"

	// 퍼블릭 주소 얻기
	publicAddr, err := getPublicAddress(stunServerAddr)
	if err != nil {
		fmt.Println("퍼블릭 주소 얻기 실패:", err)
		return
	}
	fmt.Print("퍼블릭 주소: ")
	fmt.Println(publicAddr)

	fmt.Println("This node:", node.ID().Pretty(), "\n", node.Addrs())

	if len(os.Args) > 1 {
		addr, err := multiaddr.NewMultiaddr(os.Args[1])
		if err != nil {
			panic(err)
		}
		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			panic(err)
		}

		err = node.Connect(ctx, *peerInfo)
		if err != nil {
			panic(err)
		}

		fmt.Println("Connected to:", peerInfo.ID)

		stream, err := node.NewStream(ctx, peerInfo.ID, "/chat/1.0.0")
		if err != nil {
			panic(err)
		}

		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		go readData(rw)
		go writeData(rw)

		select {} // hang forever
	} else {
		fmt.Println("Waiting for incoming connections...")
		select {} // hang forever
	}
}
