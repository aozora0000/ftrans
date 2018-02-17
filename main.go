package main

import (
	"flag"
	"fmt"
	"os"

	"git.mox.si/tsuzu/go-easyp2p"
)

var (
	mode      = flag.String("mode", "client", "Server(signaling server) or client(sender or reciever)")
	stun      = flag.String("stun", "stun.l.google.com:19302", "STUN server addresses(split with ',')")
	signaling = flag.String("sig", "wss://ftrans.cs3238.com/ws", "Signaling server address")
	version   = flag.Bool("v", false, "Show version")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage of ftrans:")
		fmt.Fprintln(os.Stderr, "  ftrans [options] password [file paths...]")
		fmt.Fprintln(os.Stderr, "  If no path is passed, this runs as a reciever.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "  To launch a signaling server: ftrans --mode server")
		fmt.Fprintln(os.Stderr)

		flag.PrintDefaults()
	}

	flag.Parse()

	if *version {
		fmt.Fprintln(os.Stderr, "ftrans version:", binaryVersion, "("+binaryRevision+")")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "[Details]")
		fmt.Fprintln(os.Stderr, "ftrans protocol version:", ProtocolVersionLatest)
		fmt.Fprintln(os.Stderr, "go-easyp2p version:", easyp2p.P2PVersionLatest)

		return
	}

	if *mode == "server" {
		runServer()
	} else {
		runClient()
	}
}
