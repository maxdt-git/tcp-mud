package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

func (g *Game) Connect(conn net.Conn, parts []string) string {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	if len(parts) < 2 {
		return "ERR missing_name"
	}

	name := parts[1]
	_, ok := g.Players[name]

	if ok {
		return "ERR 201 NAME_IN_USE"
	}

	if g.Clients[conn].Player != nil {
		return "ERR client already have a player"
	}

	g.Players[name] = &Player{ // creates the player
		Name:      name,
		Room:      "1",
		Inventory: make(map[string]string),
		Dialogue: false,
	}
	g.Clients[conn].Player = g.Players[name]
	g.Rooms["1"].Players = append(g.Rooms["1"].Players, g.Players[name])
	return "OK connected " + name
}

func (g *Game) Look(conn net.Conn) string {
	player_room := g.Clients[conn].Player.Room
	room := g.Rooms[player_room]
	players := []string{} // {} at the end
	// players = players[:0]
	for _, p := range room.Players {
		players = append(players, p.Name)
	}
	fmt.Println(players)
	res := LookResponse{
		Room: RoomResponse{
			ID:          room.Id,
			Name:        room.Name,
			Description: room.Description,
			Exits:       room.Exits,
		},
		Players: players,
		Items:   room.Items,
		Npcs:    room.Npcs,
	}
	data, err := json.Marshal(res)
	if err != nil {
		return "ERR 901 SEND_FAILED"
	}
	return "OK " + string(data)
}

func (g *Game) Who(conn net.Conn) string {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	names := []string{}
	for name := range g.Players {
		names = append(names, name)
	}
	return "OK players=" + strings.Join(names, ",")
}

func (g *Game) Move(conn net.Conn, parts []string) string {
	if len(parts) < 2 {
		return "ERR missing_direction"
	}
	direction := parts[1]
	return "OK room= " + direction
}

func (g *Game) Chat(conn net.Conn, parts []string) string {
	if len(parts) < 2 {
		return "ERR missing_message"
	}
	message := strings.Join(parts[1:], " ")
	return "OK chat=" + message
}

func (g *Game) Talk(conn net.Conn, parts []string) string {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	if parts[0] == "TALK" && len(parts) < 2 {
		return "ERR missing_message"
	}

	test := strings.Join(parts, " ")
	test, _ = strings.CutPrefix(test, "TALK ")

	// fmt.Println(test)
	dialogue := g.Rooms[g.Clients[conn].Player.Room].Dialogue

	if g.Rooms[g.Clients[conn].Player.Room].Npcs[0] == test {
		g.Clients[conn].Player.Dialogue = true
		return "OK " + dialogue[0]
	}
	if test == "yes" {
		return "OK " + dialogue[1]
	}
	if test == "no" {
		g.Clients[conn].Player.Dialogue = false
		return "OK " + dialogue[2]
	}
	if test == "a river" || test == "river" {
		g.Clients[conn].Player.Dialogue = false
		return "OK " + dialogue[4]
	}
	return "ERR 404 NPC_NOT_FOUND"
}

func (g *Game) Quit(conn net.Conn) string {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	client := g.Clients[conn]

	if client != nil && client.Player != nil {
		delete(g.Players, client.Player.Name) // delete removes the g.players in key client.Player.Name
	}

	delete(g.Clients, conn)
	return "OK leaving game"
}

func (g *Game) Take(conn net.Conn) string {
	player_room := g.Clients[conn].Player.Room
	room := g.Rooms[player_room]
	if room.Items != nil {
		for name := range room.Items {
			g.Clients[conn].Player.Inventory[name] = room.Items[name]
			delete(room.Items, name)
			return "OK taken=" + name
		}
	}
	return "ERR 404 ITEM_NOT_FOUND"
}

func (g *Game) Bag(conn net.Conn) string {
	player_name := g.Clients[conn].Player.Name
	player := g.Players[player_name]
	data, err := json.Marshal(player.Inventory)
	if err != nil {
		return "ERR 901 SEND_FAILED"
	}
	return "OK " + string(data)
}

func (g *Game) Badmanners(conn net.Conn) string {
	data := []string{}
	for name := range g.Rooms["1"].Items {
		data = append(data, name)
	}
	return data[0]
}
