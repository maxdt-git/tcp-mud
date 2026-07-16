package main

import (
	"bufio"
	"fmt"
	"net"
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

	go readServer(conn, incoming)

	gui(conn, incoming)
}
