package main

import "fyne.io/fyne/v2/widget"

type UI struct {
	RoomTitle       *widget.Label
	RoomDescription *widget.Label

	Players         []string
	PlayersList     *widget.List

	Monsters        []string
	MonstersList    *widget.List

	Items           []string
	ItemsList       *widget.List

	Response        *widget.Label

	Exits           []string
	MoveButton      *widget.Button

	ConnectButton   *widget.Button
}
