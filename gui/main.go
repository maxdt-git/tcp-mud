package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func readServer(conn net.Conn, incoming chan <-string) {
	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			incoming <- "Disconnected from server"
			return
		}
		incoming <-line
	}
}

func main() {
	conn, err := net.Dial("tcp", "localhost:4242")
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}
	defer conn.Close()

	incoming := make(chan string) //tuyau
	// incoming <-"msg"
	// ms := <- incoming
	go readServer(conn, incoming)

	stdin := bufio.NewReader(os.Stdin)

	for {
		text, err := stdin.ReadString('\n')
		if err != nil {
			return
		}
		fmt.Fprint(conn, text)
	}
}
