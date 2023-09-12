package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage: proxy <listen address> <target address>")
		os.Exit(1)
	}
	listenAddr := os.Args[1]
	targetAddr := os.Args[2]

	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	log.Printf("server started on: %v", l.Addr())
	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go handleConn(conn, targetAddr)
	}
}

func handleConn(clientConn net.Conn, targetAddr string) {
	defer clientConn.Close()

	log.Printf("[%v] new connection established", clientConn.RemoteAddr())
	defer func() {
		log.Printf("[%v] connection closed", clientConn.RemoteAddr())
	}()

	targetConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		log.Printf("failed connect target server: %v", err)
		return
	}
	defer targetConn.Close()
	log.Printf("[%v -> %v] proxy tunnel established", clientConn.RemoteAddr(), targetAddr)

	var closeOnce sync.Once
	closeConns := func() {
		closeOnce.Do(func() {
			clientConn.Close()
			targetConn.Close()
		})
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer closeConns()

		io.Copy(targetConn, clientConn)
	}()
	go func() {
		defer wg.Done()
		defer closeConns()

		io.Copy(clientConn, targetConn)
	}()
	wg.Wait()
}
