package main

import (
	"bufio"
	// "container/list"
	"fmt"
	"net"
	"strings"
)

func handleCommand(command string, conn net.Conn, game *Game) string {
	parts := strings.Fields(command)

	if len(parts) == 0 {
		return "ERR empty_command"
	}

	if parts[0] != "CONNECT" {
		if game.Clients[conn].Player == nil {
			return "ERR connect first"
		}
	}

	if game.Clients[conn].Player != nil {
		if game.Clients[conn].Player.Dialogue == true {
			test := strings.Join(parts, " ")
			switch test {
			case "yes":
				return game.Talk(conn, parts)
			case "no":
				return game.Talk(conn, parts)
			case "a river":
				return game.Talk(conn, parts)
			case "river":
				return game.Talk(conn, parts)
			default:
				return "OK " + game.Rooms[game.Clients[conn].Player.Room].Dialogue[3]
			}
		}
	}

	switch parts[0] {
	case "LOOK":
		return game.Look(conn)

	case "WHO":
		return game.Who(conn)

	case "CONNECT":
		return game.Connect(conn, parts)

	case "MOVE":
		return game.Move(conn, parts)

	case "CHAT":
		return game.Chat(conn, parts)

	case "TALK":
		return game.Talk(conn, parts)

	case "QUIT":
		return game.Quit(conn)

	case "TAKE":
		return game.Take(conn)

	case "BAG":
		return game.Bag(conn)

	case "WTF":
		return game.Badmanners(conn)

	default:
		return "ERR unknown_command"
	}
}

func handle(conn net.Conn, game *Game) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected")
			delete(game.Players, game.Clients[conn].Player.Name)
			return
		}
		command := strings.TrimSpace(message)
		response := handleCommand(command, conn, game)

		fmt.Fprintln(conn, response)

		if command == "QUIT" {
			return
		}
	}
}

func main() {

	game := &Game{
		Players: make(map[string]*Player),
		Clients: make(map[net.Conn]*Client),
		Rooms:   make(map[string]*Room),
	}
	// open the json and stock the data in game
	openJson(game)

	listener, err := net.Listen("tcp", ":4242") // creates listener
	if err != nil {
		fmt.Println("ERR:", err)
		return
	}

	defer listener.Close()
	fmt.Println("Server listening on port 4242")

	for { // infinite loop because we want the server to listen to new connections indefinitely
		conn, err := listener.Accept() // waits for connection to listener port
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		game.Clients[conn] = &Client{
			Conn: conn,
		}
		fmt.Println("Client connected:", conn.RemoteAddr())
		go handle(conn, game) // starts a go routine and returns to the loop
	}
}
