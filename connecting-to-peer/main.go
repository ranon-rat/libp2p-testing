package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

const protocolID = "/example/1.0.0"

func writeCounter(s network.Stream) {

	for {
		<-time.After(time.Second)
		// so , the network.Stream works as a io.Writer
		// thats make things easier I think
		err := binary.Write(s, binary.BigEndian, uint64(rand.Int()))
		if err != nil {

			s.Close()

			return
		}
	}
}

func readCounter(s network.Stream) {
	for {
		var counter uint64
		// but at the same time it works as a io.Reader
		// hm
		err := binary.Read(s, binary.BigEndian, &counter)
		if err != nil {
			fmt.Println("disconnected to", s.ID())
			s.Close()
			return
		}

		fmt.Printf("Received %d from %s\n", counter, s.ID())
	}
}
func main() {

	peerAddr := flag.String("peer-address", "", "peer address")
	author = *flag.String("username", "username")
	node, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
	if err != nil {
		panic(err)
	}
	defer node.Close()
	fmt.Println("Host address:", node.Addrs())
	fmt.Println("Host id:", node.ID())
	fmt.Println("connect with:", node.Addrs()[0].String()+"/p2p/"+node.ID().String())
	// con esto lo que hago es simplemente dejar que en un puerto me pueda comunicar por medio de un protocolo
	node.SetStreamHandler(protocolID, func(s network.Stream) {
		go writeCounter(s)
		go readCounter(s)
	})
	flag.Parse()

	if *peerAddr != "" {
		// Parse the multiaddr string.
		peerMA, err := multiaddr.NewMultiaddr(*peerAddr)
		if err != nil {
			panic(err)
		}
		peerAddrInfo, err := peer.AddrInfoFromP2pAddr(peerMA)
		if err != nil {
			panic(err)
		}

		// Connect to the node at the given address.
		if err := node.Connect(context.Background(), *peerAddrInfo); err != nil {
			panic(err)
		}

		fmt.Println("Connected to", peerAddrInfo.String())
		// con esto solo me conecto a ese puerto
		s, _ := node.NewStream(context.Background(), peerAddrInfo.ID, protocolID)

		// Start the write and read threads.
		go writeCounter(s)
		go readCounter(s)
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	fmt.Println("Received signal, shutting down...")

}
