package game

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
		Dialogue:  false,
	}
	g.Clients[conn].Player = g.Players[name]
	g.Rooms["1"].Players = append(g.Rooms["1"].Players, g.Players[name])
	return "OK connected " + name
}

func (g *Game) Look(conn net.Conn) string {
	player_room := g.Clients[conn].Player.Room
	room := g.Rooms[player_room]
	players := []string{} // {} at the end
	for _, p := range room.Players {
		players = append(players, p.Name)
	}
	res := LookResponse{
		Room: RoomResponse{
			ID:          room.Id,
			Name:        room.Name,
			Description: room.Description,
			Exits:       room.Exits,
		},
		Players: players,
		Items:   room.Items,
		NPCs:    room.Npcs,
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

func (g *Game) Take(conn net.Conn, parts []string) string {
	player_room := g.Clients[conn].Player.Room
	room := g.Rooms[player_room]
	if len(parts) < 2 {
		return "ERR missing item name"
	}
	item := strings.Join(parts[1:], " ")
	_, ok := room.Items[item]
	if ok {
		g.Clients[conn].Player.Inventory[item] = room.Items[item]
		delete(room.Items, item)
		return "OK taken=" + item
	}
	return "ERR 404 ITEM_NOT_FOUND"
}

func (g *Game) Drop(conn net.Conn, parts []string) string {
	player_room := g.Clients[conn].Player.Room
	room := g.Rooms[player_room]
	if len(parts) < 2 {
		return "ERR missing item name"
	}
	item := strings.Join(parts[1:], " ")
	_, ok := g.Clients[conn].Player.Inventory[item]
	if ok {
		room.Items[item] = g.Clients[conn].Player.Inventory[item]
		delete(g.Clients[conn].Player.Inventory, item)
		return "OK dropped=" + item
	}
	return "ERR 404 ITEM_NOT_IN_INVENTORY"
}

// group functions

func (g *Game) Group(conn net.Conn, parts []string) string {
	if len(parts) < 2 {
		return "ERR missing argument"
	}
	switch parts[1] {
	case "CREATE":
		return g.GroupCreate(conn, parts)

	case "INVITE":
		return g.GroupInvite(conn, parts)

	case "JOIN":
		return g.GroupJoin(conn, parts)
	}
	return ""
}

func (g *Game) GroupCreate(conn net.Conn, parts []string) string {
	player_name := g.Clients[conn].Player.Name
	player := g.Players[player_name]
	if player.Group != nil {
		return "ERR already in a group"
	}
	group_name := player_name + "party"
	g.Groups[group_name] = &Group{
		ID:      group_name,
		Leader:  g.Players[player_name],
		Members: make(map[string]*Player),
	}
	g.Groups[group_name].Members[player_name] = player
	player.Group = g.Groups[group_name]
	return "OK group=" + group_name
}

func (g *Game) GroupInvite(conn net.Conn, parts []string) string {
	player_name := g.Clients[conn].Player.Name
	player := g.Players[player_name]

	if len(parts) < 3 {
		return "ERR missing player name"
	}

	if player.Group == nil {
		return "ERR not in a group"
	}

	if parts[2] == player_name {
		return "ERR can't invite yourself"
	}

	_, ok := g.Players[parts[2]]
	if !ok {
		return "ERR player not found"
	}
	var invite_conn net.Conn
	var invite_name string

	for i := range g.Clients {
		if g.Clients[i].Player.Name == parts[2] {
			invite_conn = g.Clients[i].Conn
			invite_name = g.Clients[i].Player.Name
		}
	}

	if g.Players[invite_name].Group != nil {
		return "ERR 402 ALREADY_IN_GROUP"
	}
	groupName := player.Group.ID
	fmt.Fprintln(invite_conn, "EVT GROUP INVITE "+groupName)
	g.Players[invite_name].Invites = append(g.Players[invite_name].Invites, groupName)
	return "OK"
}

func (g *Game) GroupJoin(conn net.Conn, parts []string) string {
	player_name := g.Clients[conn].Player.Name
	player := g.Players[player_name]

	if len(parts) < 3 {
		return "ERR missing group name"
	}

	if player.Invites == nil {
		return "ERR no invitation pending"
	}

	var is_valid bool

	for i := range player.Invites {
		if player.Invites[i] == parts[2] {
			is_valid = true
			player.Invites = append(player.Invites[:i], player.Invites[i+1:]...)
			break
		}
	}
	if is_valid == true {
		var group_names []string
		var group_conn []net.Conn
		group := g.Groups[parts[2]]
		for i := range group.Members {
			group_names = append(group_names, group.Members[i].Name)
		}
		for j := range group_names {
			for k := range g.Clients {
				if g.Clients[k].Player.Name == group_names[j] {
					group_conn = append(group_conn, g.Clients[k].Conn)
				}
			}
		}
		for l := range group_conn {
			fmt.Fprintln(group_conn[l], "EVT GROUP JOIN "+player_name)
		}
		group.Members[player.Name] = player
		player.Group = group
		return "OK group= " + parts[2]
	}
	return "ERR invalid invitaion"
}

// end group functions

func (g *Game) Inventory(conn net.Conn) string {
	player_name := g.Clients[conn].Player.Name
	player := g.Players[player_name]
	data, err := json.Marshal(player.Inventory)
	if err != nil {
		return "ERR 901 SEND_FAILED"
	}
	return "OK " + string(data)
}

func (g *Game) Badmanners(conn net.Conn) string {
	player_name := g.Clients[conn].Player.Name
	player := g.Players[player_name]
	for i := range player.Group.Members {
		fmt.Fprintln(conn, player.Group.Members[i].Name)
	}
	return ""
}
