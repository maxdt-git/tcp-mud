package game

import (
	"net"
	"sync"
)

type Client struct {
	Conn   net.Conn
	Player *Player
}

type Player struct {
	Name         string
	Room         string
	Inventory    map[string]string
	Group        *Group
	Invites      []string
	TalkingNpc   string
	DialogueNode string
	Hp           int
	Max_hp       int
	Status       string
	Damage       int
}

type DialogueNode struct {
	Text      string            `json:"text"`
	Responses map[string]string `json:"responses"`
	End       bool              `json:"end"`
}

type Npc struct {
	Name  string                  `json:"name"`
	Start string                  `json:"start"`
	Nodes map[string]DialogueNode `json:"nodes"`
}

type Room struct {
	Id          int               `json:"room"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       []string          `json:"exits"`
	Items       map[string]string `json:"items"`
	Neighbours  map[string]string `json:"neighbours"`
	Monsters    []Monster         `json:"monsters"`
	Dialogue    []string          `json:"dialogue"`
	Npcs        []Npc             `json:"npcs"`
	Players     []*Player         `json:"-"`
}

type Game struct {
	Mutex   sync.Mutex
	Players map[string]*Player
	Clients map[net.Conn]*Client
	Rooms   map[string]*Room
	Groups  map[string]*Group
}

type RoomResponse struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Exits       []string `json:"exits"`
}

type LookResponse struct {
	Room     RoomResponse      `json:"room"`
	Players  []string          `json:"players"`
	Items    map[string]string `json:"items"`
	NPCs     []string          `json:"npcs"`
	Monsters []string          `json:"monsters"`
}

type StatusResponse struct {
	Hp     int    `json:"hp"`
	Max_hp int    `json:"max_hp"`
	Status string `json:"status"`
}

type CombatResponse struct {
	Player_hp int    `json:"hp"`
	Target_hp int    `json:"target_hp"`
	Damage    int    `json:"damage"`
	Counter   int    `json:"counter_atk"`
	Status    string `json:"status"`
}

type Group struct {
	ID      string
	Leader  *Player
	Members map[string]*Player
}

type Monster struct {
	Name   string `json:"name"`
	HP     int    `json:"hp"`
	Attack int    `json:"attack"`
	Alive  bool   `json:"alive"`
}
