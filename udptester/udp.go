package main

import (
	"log"
	"fmt"
	"time"
	"context"
	"os"
	"os/signal"
	"syscall"

	"tailscale.com/tsnet"
)

func main() {
	s := &tsnet.Server{
		ControlURL: "https://login.xedge.cc",
		// A non reusable auth key.
		AuthKey: "5f19451a33005064",
	}
	defer s.Close()

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}

	// bring up the userspace client.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	status, err := s.Up(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("tsnet status %v\n", status)

	// Test UDP client.
	// You can do this in the server side to launch a simple UDP echo server.
	// ncat -e /bin/cat -k -u -l 1235
	// Use ncat -u 100.64.<server-ip> 1235 to test it locally to ensure it works.
	//
	// change the following ip to your ts server side ip.
	conn, err := s.Dial(ctx, "udp", "100.64.0.2:1235")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	want := "hellotsnet"
	if _, err := conn.Write([]byte(want)); err != nil {
		log.Fatal(err)
	}
	fmt.Println("send data ok")

	// Read echo reply the packet on the conn.
	got := make([]byte, 1024)
	n, err := conn.Read(got)
	if err != nil {
		log.Fatal(err)
	}
	got = got[:n]
	if string(got) != want {
		log.Fatal("got %q, want %q", got, want)
	}
	fmt.Printf("recv data %v ok\n", string(got))

	// Wait for non userspace client to ping us :)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)

	fmt.Println("Waiting for Ctrl+C to exit...")
	<-sig
	fmt.Println("Ctrl+C received. Exiting...")
}
