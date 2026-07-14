package game

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func (g *Game) Connect(conn net.Conn, parts []string) string {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	if len(parts) < 2 {
		return "ERR 202 MISSING_NAME"
	}

	name := parts[1]
	_, ok := g.Players[name]

	if ok {
		return "ERR 201 NAME_IN_USE"
	}

	if g.Clients[conn].Player != nil {
		return "ERR 203 ALREADY_CONNECTED"
	}

	g.Players[name] = &Player{ // creates the player
		Name:      name,
		Room:      "1",
		Inventory: make(map[string]Item),
		Hp:        100,
		Max_hp:    100,
		Status:    "healthy",
		Damage:    10,
		Quests:    make(map[string]string),
	}
	g.Clients[conn].Player = g.Players[name]
	g.Rooms["1"].Players = append(g.Rooms["1"].Players, g.Players[name])
	return "OK connected " + name
}

func (g *Game) Status(conn net.Conn) string {
	player_name := g.Clients[conn].Player.Name
	player := g.Players[player_name]

	res := StatusResponse{
		Hp:     player.Hp,
		Max_hp: player.Max_hp,
		Status: player.Status,
	}

	data, err := json.Marshal(res)

	if err != nil {
		return "ERR 901 SEND_FAILED"
	}

	return "OK " + string(data)
}

func (g *Game) Look(conn net.Conn) string {
	player_room := g.Clients[conn].Player.Room
	room := g.Rooms[player_room]
	players := []string{} // {} at the end

	if player_room == "6" && !room.Lit {
		return "ERR 433 ROOM_TOO_DARK"
	}

	for _, p := range room.Players {
		players = append(players, p.Name)
	}
	npcs := []string{}
	monsters := []string{}
	items := []string{}
	questgivers := []string{}

	for name := range room.Items {
		items = append(items, name)
	}
	for _, npc := range room.Npcs {
		npcs = append(npcs, npc.Name)
	}
	for _, monster := range room.Monsters {
		monsters = append(monsters, monster.Name)
	}
	for _, questgiver := range room.Questgivers {
		questgivers = append(questgivers, questgiver.Name)
	}
	res := LookResponse{
		Room: RoomResponse{
			ID:          room.Id,
			Name:        room.Name,
			Description: room.Description,
			Exits:       room.Exits,
		},
		Players:     players,
		Items:       items,
		NPCs:        npcs,
		Monsters:    monsters,
		Questgivers: questgivers,
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
		return "ERR 302 MISSING_DIRECTION"
	}
	direction := parts[1]
	player_name := g.Clients[conn].Player.Name
	player := g.Players[player_name]
	cur_room := g.Rooms[player.Room]
	old_room := player.Room

	_, ok := player.Inventory["crafted key"]

	if player.Room == "5" && direction == "north" && !ok {
		return "ERR 442 DOOR_LOCKED"
	}

	for i := range cur_room.Exits {
		if cur_room.Exits[i] == direction {
			player.Room = cur_room.Neighbours[direction]
			// EVT presence enter and leave
			for _, client := range g.Clients {
				if client == nil || client.Player == nil {
					continue
				}

				if client.Player.Name == player_name {
					continue
				}

				if client.Player.Room == old_room {
					fmt.Fprintln(
						client.Conn,
						"EVT ROOM PRESENCE LEAVE "+player_name,
					)
				}

				if client.Player.Room == player.Room {
					fmt.Fprintln(
						client.Conn,
						"EVT ROOM PRESENCE ENTER "+player_name,
					)
				}
			}
			// add player to room, remove from old
			g.Rooms[player.Room].Players = append(g.Rooms[player.Room].Players, g.Players[player_name])
			players := g.Rooms[old_room].Players
			for i, p := range players {
				if p == player {
					g.Rooms[old_room].Players = append(players[:i], players[i+1:]...)
					break
				}
			}
			return "OK room=" + g.Rooms[player.Room].Name
		}
	}
	return "ERR 301 NO_EXIT"
}

// Chat functions

func (g *Game) Chat(conn net.Conn, parts []string) string {
	if len(parts) < 2 {
		return "ERR 407 MISSING_ARGUMENT"
	}
	switch parts[1] {
	case "GLOBAL":
		return g.ChatGlobal(conn, parts)

	case "ROOM":
		return g.ChatRoom(conn, parts)

	case "GROUP":
		return g.ChatGroup(conn, parts)
	}
	return "ERR 407 WRONG ARGUMENT"
}

func (g *Game) ChatGlobal(conn net.Conn, parts []string) string {
	if len(parts) < 3 {
		return "ERROR 407 MISSING_MESSAGE"
	}

	player_name := g.Clients[conn].Player.Name

	message := strings.Join(parts[2:], " ")

	for i := range g.Clients {
		if g.Clients[i].Player.Name != player_name {
			fmt.Fprintln(g.Clients[i].Conn, "EVT GLOBAL CHAT "+player_name+": "+message)
		}
	}

	return "OK"
}

func (g *Game) ChatRoom(conn net.Conn, parts []string) string {
	if len(parts) < 3 {
		return "ERROR 407 MISSING_MESSAGE"
	}

	player_name := g.Clients[conn].Player.Name
	player_room := g.Players[player_name].Room

	message := strings.Join(parts[2:], " ")

	for i := range g.Clients {
		if g.Clients[i].Player.Name != player_name {
			if g.Clients[i].Player.Room == player_room {
				fmt.Fprintln(g.Clients[i].Conn, "EVT ROOM CHAT "+player_name+": "+message)
			}
		}
	}

	return "OK"
}

func (g *Game) ChatGroup(conn net.Conn, parts []string) string {
	if len(parts) < 3 {
		return "ERROR 407 MISSING_MESSAGE"
	}

	player_name := g.Clients[conn].Player.Name
	player := g.Players[player_name]

	if player.Group == nil {
		return "ERR 401 NOT_IN_GROUP"
	}

	message := strings.Join(parts[2:], " ")

	for i := range g.Clients {
		if g.Clients[i].Player.Name != player_name {
			if g.Clients[i].Player.Group != nil && g.Clients[i].Player.Group == player.Group {
				fmt.Fprintln(g.Clients[i].Conn, "EVT GROUP CHAT "+player_name+": "+message)
			}
		}
	}

	return "OK"
}

// end Chat functions

func (g *Game) Talk(conn net.Conn, parts []string) string {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	if len(parts) < 2 {
		return "ERR 407 MISSING_MESSAGE"
	}

	player := g.Clients[conn].Player
	room := g.Rooms[player.Room]
	input := strings.Join(parts[1:], " ")

	// not currently in a conversation must name an NPC present in the room
	if player.TalkingNpc == "" {
		for _, npc := range room.Npcs {
			if npc.Name == input {
				node, ok := npc.Nodes[npc.Start]
				if !ok {
					return "ERR 404 NPC_NOT_FOUND"
				}
				player.TalkingNpc = npc.Name
				player.DialogueNode = npc.Start
				return "OK " + node.Text
			}
		}
		return "ERR 404 NPC_NOT_FOUND"
	}

	// already talking to an npc: try to advance in the dialogue
	var npc *Npc

	for i := range room.Npcs {
		if room.Npcs[i].Name == player.TalkingNpc {
			npc = &room.Npcs[i]
			break
		}
	}

	node := npc.Nodes[player.DialogueNode]
	next, ok := node.Responses[input]
	if !ok {
		next, ok = node.Responses["wrong"]
	}
	if !ok {
		return "ERR 404 INVALID_RESPONSE"
	}

	nextNode, ok := npc.Nodes[next]
	if !ok {
		return "ERR 404 NPC_NOT_FOUND"
	}

	if nextNode.End {
		player.TalkingNpc = ""
		player.DialogueNode = ""
	} else {
		player.DialogueNode = next
	}
	return "OK " + nextNode.Text
}

func (g *Game) Quit(conn net.Conn) string {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	client := g.Clients[conn]
	player_room := client.Player.Room
	room := g.Rooms[player_room]

	if client != nil && client.Player != nil {
		for i, p := range room.Players {
			if p.Name == client.Player.Name {
				room.Players = append(room.Players[:i], room.Players[i+1:]...) // remove from room
				break
			}
		}
		g.GroupLeave(conn)
		delete(g.Players, client.Player.Name) // delete removes the g.players in key client.Player.Name
	}

	delete(g.Clients, conn)
	return "OK leaving game"
}

func (g *Game) Take(conn net.Conn, parts []string) string {
	player_room := g.Clients[conn].Player.Room
	room := g.Rooms[player_room]
	if len(parts) < 2 {
		return "ERR 407 MISSING_ITEM_NAME"
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
		return "ERR 407 MISSING_ITEM_NAME"
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
		return "ERR 407 MISSING_ARGUMENT"
	}
	switch parts[1] {
	case "CREATE":
		return g.GroupCreate(conn, parts)

	case "INVITE":
		return g.GroupInvite(conn, parts)

	case "JOIN":
		return g.GroupJoin(conn, parts)

	case "LEAVE":
		return g.GroupLeave(conn)
	}

	return ""
}

func (g *Game) GroupCreate(conn net.Conn, parts []string) string {
	player_name := g.Clients[conn].Player.Name
	player := g.Players[player_name]
	if player.Group != nil {
		return "ERR 402 ALREADY_IN_GROUP"
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
		return "ERR 407 MISSING_PLAYER_NAME"
	}

	if player.Group == nil {
		return "ERR 401 NOT_IN_GROUP"
	}

	if parts[2] == player_name {
		return "ERR 408 AUTO_INVITE"
	}

	_, ok := g.Players[parts[2]]
	if !ok {
		return "ERR 409 PLAYER_NOT_FOUND"
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
		return "ERR 407 MISSING_GROUP_NAME"
	}

	if player.Invites == nil {
		return "ERR 410 NO_INVITES"
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
		return "OK group=" + parts[2]
	}
	return "ERR 411 INVALID_INVITE"
}

func (g *Game) GroupLeave(conn net.Conn) string {
	player_name := g.Clients[conn].Player.Name
	player := g.Players[player_name]

	if player.Group == nil {
		return "401	NOT_IN_GROUP"
	}
	group := player.Group
	group_name := group.ID
	delete(player.Group.Members, player_name)
	if player.Group.Leader.Name == player_name && len(player.Group.Members) > 0 {
		for _, member := range player.Group.Members {
			player.Group.Leader = member
			break
		}
	}
	if len(group.Members) > 0 {
		var members_name []string
		var members_conn []net.Conn

		for i := range group.Members {
			members_name = append(members_name, group.Members[i].Name)
		}
		for j := range members_name {
			for k := range g.Clients {
				if g.Clients[k].Player.Name == members_name[j] {
					members_conn = append(members_conn, g.Clients[k].Conn)
				}
			}
		}
		for l := range members_conn {
			fmt.Fprintln(members_conn[l], "EVT GROUP LEAVE "+player_name)
		}
	}
	if len(player.Group.Members) == 0 {
		delete(g.Groups, group_name)
	}
	player.Group = nil
	return "OK"
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

func (g *Game) Loot(conn net.Conn, parts []string) string {
	if len(parts) < 2 {
		return "ERR 407 MISSING_TARGET_NAME"
	}
	target := strings.Join(parts[1:], " ")
	player_name := g.Clients[conn].Player.Name
	player := g.Players[player_name]
	player_room := g.Rooms[player.Room]

	// get the id of monster
	id := -1

	for i := range player_room.Monsters {
		if target == player_room.Monsters[i].Name {
			id = i
		}
	}
	// check if possible to loot
	if id == -1 {
		return "ERR 404 MONSTER_NOT_FOUND"
	}

	if len(player_room.Monsters[id].Items) == 0 {
		return "ERR 404 NOTHING_TO_LOOT"
	}

	if player_room.Monsters[id].Alive == true {
		return "ERR 421 CANT_LOOT_ALIVE_MONSTER"
	}

	// loot
	for name, desc := range player_room.Monsters[id].Items {
		player.Inventory[name] = desc
		delete(player_room.Monsters[id].Items, name)
		return "OK taken=" + name
	}
	return ""
}

func (g *Game) Quest(conn net.Conn, parts []string) string {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	if len(parts) < 2 {
		return "ERR 407 MISSING_MESSAGE"
	}

	player := g.Clients[conn].Player
	room := g.Rooms[player.Room]
	input := strings.Join(parts[1:], " ")

	// not currently in a conversation must name a questgiver present in the room
	if player.TalkingQuestgiver == "" {
		for _, questgiver := range room.Questgivers {
			if questgiver.Name == input {
				node, ok := questgiver.Nodes[questgiver.Start]
				if !ok {
					return "ERR 404 QUESTGIVER_NOT_FOUND"
				}
				player.TalkingQuestgiver = questgiver.Name
				player.DialogueNode = questgiver.Start
				return "OK " + node.Text
			}
		}
		return "ERR 404 QUESTGIVER_NOT_FOUND"
	}

	// already talking to an questgiver: try to advance in the dialogue
	var questgiver *Questgiver

	for i := range room.Questgivers {
		if room.Questgivers[i].Name == player.TalkingQuestgiver {
			questgiver = &room.Questgivers[i]
			break
		}
	}

	node := questgiver.Nodes[player.DialogueNode]
	next, ok := node.Responses[input]
	if !ok {
		return "ERR 404 INVALID_RESPONSE"
	}

	nextNode, ok := questgiver.Nodes[next]
	if !ok {
		return "ERR 404 QUESTGIVER_NOT_FOUND"
	}

	if nextNode.End {
		player.TalkingQuestgiver = ""
		player.DialogueNode = ""
	} else {
		player.DialogueNode = next
	}
	if strings.HasPrefix(nextNode.Text, "QUEST ACCEPTED") {
		quest, ok := strings.CutPrefix(nextNode.Text, "QUEST ACCEPTED: ")
		if !ok {
			return "OK Quest not accepted"
		}
		player.Quests[quest] = questgiver.Description
	}
	return "OK " + nextNode.Text
}
func (g *Game) Quests(conn net.Conn, parts []string) string {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	player := g.Clients[conn].Player
	if len(player.Quests) == 0 {
		return "OK no active quest"
	}

	lines := []string{}
	for quest, description := range player.Quests {
		lines = append(lines, quest+": "+description)
	}
	return "OK " + strings.Join(lines, "\n")
}

func (g *Game) CompleteQuest(conn net.Conn, parts []string) string {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	if len(parts) < 2 {
		return "ERR 407 MISSING_MESSAGE"
	}
	player := g.Clients[conn].Player
	room := g.Rooms[player.Room]
	input := strings.Join(parts[1:], " ")
	find := false

	for quest := range player.Quests {
		if input == quest {
			switch quest {
			case "fouras's quest":
				for item := range player.Inventory {
					if item == "vial of venom" {
						find = true
					}
				}
				if room.Id == 2 && find {
					delete(player.Inventory, "vial of venom")
					delete(player.Quests, quest)
					key_item := Item{
						Description: "What can i do with this...",
						Type:        "special",
					}
					player.Inventory["key fragment two"] = key_item
					return "Quest finished, nice job!\nYou have received key fragment two"
				} else {
					return "ERR 456 CANT_FINISH_QUEST"
				}
			}
		}
	}
	return "ERR 456 CANT_FINISH_QUEST"
}

func (g *Game) AbandonQuest(conn net.Conn, parts []string) string {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	if len(parts) < 2 {
		return "ERR 407 MISSING_MESSAGE"
	}

	player := g.Clients[conn].Player
	input := strings.Join(parts[1:], " ")

	for quest := range player.Quests {
		if input == quest {
			switch quest {
			case "fouras's quest":
				delete(player.Quests, quest)
				return "Quest abandoned: "+quest
			}
		}
	}
	return "ERR 457 CANT_ABANDON_QUEST"
}

func (g *Game) Badmanners(conn net.Conn) string {
	player_name := g.Clients[conn].Player.Name
	player := g.Players[player_name]

	return "WTF " + strconv.Itoa(player.Damage)
}
