package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	mrand "math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type Message struct {
	Author  string `json:"author"`
	Content string `json:"content"`
	PeerID  string `json:"-"`
}

const protocolID = "/chat/1.0.0"

var peers = map[network.Stream]bool{}
var msgChan = make(chan Message)
var author = "guest" + strconv.Itoa(mrand.Int())

func writeMessages() {

	for {
		msg := <-msgChan
		for p := range peers {
			if p.ID() == msg.PeerID {
				continue
			}
			if err := json.NewEncoder(p).Encode(msg); err != nil {

				delete(peers, p)
				p.Close()
			}

		}
	}
}

func readCounter(p network.Stream, d *bufio.ReadWriter) {
	for {
		var msg Message
		// but at the same time it works as a io.Reader
		// hm
		err := json.NewDecoder(d).Decode(&msg)
		if err != nil {
			delete(peers, p)
			p.Close()

			return
		}
		msg.PeerID = p.ID()
		fmt.Printf("\r%s > %s\n\r> ", msg.Author, msg.Content)
		if msg.Author != author {
			msgChan <- msg
		}
	}
}
func main() {

	peerAddr := flag.String("peer-address", "", "peer address")
	auth := flag.String("username", "", "")
	prvKey, _, _ := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, mrand.New(mrand.NewSource(time.Now().Unix())))

	node, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"), libp2p.Identity(prvKey))
	if err != nil {
		panic(err)
	}
	defer node.Close()
	fmt.Println("Host address:", node.Addrs())
	fmt.Println("Host id:", node.ID())
	fmt.Println("connect with:", node.Addrs()[0].String()+"/p2p/"+node.ID().String())
	// con esto lo que hago es simplemente dejar que en un puerto me pueda comunicar por medio de un protocolo
	node.SetStreamHandler(protocolID, func(s network.Stream) {
		peers[s] = true
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
		go readCounter(s, rw)
	})
	flag.Parse()
	author = *auth
	if *peerAddr != "" {
		// Parse the multiaddr string.
		peerMA, err := multiaddr.NewMultiaddr(*peerAddr)
		if err != nil {
			log.Println(err)
		}
		peerAddrInfo, err := peer.AddrInfoFromP2pAddr(peerMA)
		if err != nil {
			log.Println(err)
		}

		// Connect to the node at the given address.
		if err := node.Connect(context.Background(), *peerAddrInfo); err != nil {
			log.Println(err)
		}

		// con esto solo me conecto a ese puerto
		s, _ := node.NewStream(context.Background(), peerAddrInfo.ID, protocolID)
		peers[s] = true

		// Start the write and read threads.
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

		go readCounter(s, rw)
	}
	go writeMessages()
	go func() {
		for {
			fmt.Print("> ")
			var msg Message
			reader := bufio.NewReader(os.Stdin)
			content, _ := reader.ReadString('\n')
			msg.Content = content[:len(content)-1]
			msg.Author = author
			msgChan <- msg

		}
	}()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	fmt.Println("Received signal, shutting down...")

}
