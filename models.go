package main

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
	Dialogue  bool
}

type Room struct {
	Id          int               `json:"room"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       []string          `json:"exits"`
	Items       map[string]string `json:"items"`
	Npcs        []string          `json:"npcs"`
	Dialogue    []string          `json:"dialogue"`
	Players     []*Player         `json:"-"`
}

type Game struct {
	Mutex   sync.Mutex
	Players map[string]*Player
	Clients map[net.Conn]*Client
	Rooms   map[string]*Room
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
	Npcs    []string          `json:"npcs"`
}

// type Dialogue struct {
// 	ID       int
// 	Dialogue bool
// }
