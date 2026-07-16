package game

import (
	"net"
	"strings"
)

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
			if questgiver.Quest_done {
				return "ERR 406 NO_QUEST_AVAILABLE "
			}
			if questgiver.Name == input {
				node, ok := questgiver.Nodes[questgiver.Start]
				if !ok {
					return "ERR 404 NPC_NOT_FOUND"
				}
				player.TalkingQuestgiver = questgiver.Name
				player.DialogueNode = questgiver.Start
				return "OK " + node.Text
			}
		}
		return "ERR 404 NPC_NOT_FOUND"
	}

	// already talking to a questgiver: try to advance in the dialogue
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
		return "ERR 404 NPC_NOT_FOUND"
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
					for i := range room.Questgivers {
						if room.Questgivers[i].Name == "fouras" {
							room.Questgivers[i].Quest_done = true
							break
						}
					}
					return "Quest finished, nice job!\nYou have received key fragment two"
				} else {
					return "ERR 456 CANT_FINISH_QUEST"
				}

			case "gatekeeper's quest":
				for _, room := range g.Rooms {
					if room.Id < 1 || room.Id > 4 {
						continue
					}
					for _, monster := range room.Monsters {
						if monster.Alive {
							return "ERR 456 CANT_FINISH_QUEST"
						}
					}
				}
				delete(player.Quests, quest)
				for i := range room.Questgivers {
					if room.Questgivers[i].Name == "gatekeeper" {
						room.Questgivers[i].Quest_done = true
						break
					}
				}
				return "Quest finished, you can now enter the dungeon"

			case "Alfred the innkeeper's quest":
				for item := range player.Inventory {
					if item == "onion" {
						find = true
					}
				}
				if room.Id == 3 && find {
					delete(player.Inventory, "onion")
					delete(player.Quests, quest)
					key_item := Item{
						Description: "I won the game i guess",
						Type:        "special",
					}
					player.Inventory["trophy"] = key_item
					for i := range room.Questgivers {
						if room.Questgivers[i].Name == "Alfred the innkeeper" {
							room.Questgivers[i].Quest_done = true
							break
						}
					}
					return "Quest finished, nice job!\nYou have received trophy"
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
				return "Quest abandoned: " + quest
			}
		}
	}
	return "ERR 457 CANT_ABANDON_QUEST"
}
