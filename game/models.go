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
	Name      string
	Room      string
	Inventory map[string]string
	Group     *Group
	Invites   []string
	Dialogue  bool
	Hp        int
	Max_hp    int
	Status    string
}

type Room struct {
	Id          int               `json:"room"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       []string          `json:"exits"`
	Items       map[string]string `json:"items"`
	Neighbours  map[string]string `json:"neighbours"`
	Npcs        []string          `json:"npcs"`
	Dialogue    []string          `json:"dialogue"`
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
	Room    RoomResponse      `json:"room"`
	Players []string          `json:"players"`
	Items   map[string]string `json:"items"`
	NPCs    []string          `json:"npcs"`
}

type StatusResponse struct {
	Hp     int    `json:"hp"`
	Max_hp int    `json:"max_hp"`
	Status string `json:"status"`
}

type Group struct {
	ID      string
	Leader  *Player
	Members map[string]*Player
}
