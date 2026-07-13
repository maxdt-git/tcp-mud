package game

import (
	"encoding/json"
	"net"
)

func (g *Game) Attack(conn net.Conn, parts []string) string {
	if len(parts) < 2 {
		return "ERR 407 MISSING_TARGET"
	}

	player_name := g.Clients[conn].Player.Name
	player := g.Players[player_name]
	player_Room := g.Rooms[player.Room]

	// get index of monster
	monster_id := -1
	for i := range player_Room.Monsters {
		if player_Room.Monsters[i].Name == parts[1] {
			monster_id = i
			if player_Room.Monsters[i].Alive == false {
				return "ERR 420 MONSTER_DEAD"
			}
		}
	}

	if monster_id == -1 {
		return "ERR 404 MONSTER_NOT_FOUND"
	}

	// manage combat damage and status
	player_Room.Monsters[monster_id].HP = player_Room.Monsters[monster_id].HP - player.Damage
	player.Hp = player.Hp - player_Room.Monsters[monster_id].Attack
	player.Status = "combat"

	// monster death
	if player_Room.Monsters[monster_id].HP <= 0 {
		old_name := player_Room.Monsters[monster_id].Name
		player_Room.Monsters[monster_id].Alive = false
		player_Room.Monsters[monster_id].Name = "dead " + player_Room.Monsters[monster_id].Name
		player.Status = "healthy"
		return "OK " + old_name + " defeated"
	}

	// player death
	if player.Hp <= 0 {
		return "OK you died"
	}

	// create response
	res := CombatResponse{
		Player_hp: player.Hp,
		Target_hp: player_Room.Monsters[monster_id].HP,
		Damage:    player.Damage,
		Counter:   player_Room.Monsters[monster_id].Attack,
		Status:    player.Status,
	}

	data, err := json.Marshal(res)

	if err != nil {
		return "ERR 901 SEND_FAILED"
	}

	return "OK " + string(data)
}
