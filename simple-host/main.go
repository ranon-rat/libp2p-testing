package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p"
)

func main() {
	// you can configure the port and your address and the protocol that you are accepting

	host, err := libp2p.New()
	if err != nil {
		panic(err)
	}
	defer host.Close()
	// the host addrs give you a list of addreses that you can use
	// for example , it gives you your private ip and your public ip
	fmt.Println("Host address:", host.Addrs())
	fmt.Println("Host id:", host.ID())

	// this is only for closing the host, its nothing really useful
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	fmt.Println("Received signal, shutting down...")

}
