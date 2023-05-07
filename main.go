package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
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

func main() {
	ctx := context.Background()

	node, err := libp2p.New(ctx)
	if err != nil {
		panic(err)
	}

	node.SetStreamHandler("/chat/1.0.0", handleStream)

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
