package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image/color"
	"net"
	"strings"
	"tap/game"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func connectToServer(address string) (net.Conn, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func readServer(
	conn net.Conn,
	ui *UI,
) {
	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fyne.Do(func() {
				ui.Response.SetText("Disconnected from server")
			})
			return
		}

		message = strings.TrimSpace(message)

		if strings.HasPrefix(message, `OK {"room":`) {
			handleLookResponse(
				message,
				ui,
			)
			continue
		}

		fyne.Do(func() {
			ui.Response.SetText(message)
		})
	}
}

func handleLookResponse(
	message string,
	ui *UI,
) {
	jsonData := strings.TrimPrefix(message, "OK ")

	var response game.LookResponse

	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		fmt.Println("LOOK JSON error:", err)
		return
	}

	fyne.Do(func() {
		ui.RoomTitle.SetText(response.Room.Name)
		ui.RoomDescription.SetText(response.Room.Description)

		ui.Players = response.Players
		ui.PlayersList.Refresh()

		ui.Items = response.Items
		ui.ItemsList.Refresh()

		ui.Monsters = response.Monsters
		ui.MonstersList.Refresh()

		ui.Exits = response.Room.Exits
		ui.MoveButton.Enable()
		ui.ConnectButton.Enable()
	})
}

func main() {
	// connect
	conn, err := connectToServer("localhost:4242")
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}
	defer conn.Close()

	myApp := app.New()
	myWindow := myApp.NewWindow("TAP")

	roomTitle := widget.NewLabel("")
	roomDescription := widget.NewLabel("")
	responseLabel := widget.NewLabel("Connected to server")
	players := []string{}
	monsters := []string{}
	items := []string{}

	ui := &UI{
		RoomTitle:       roomTitle,
		RoomDescription: roomDescription,

		Players: players,

		Monsters: monsters,

		Items: items,

		Response: responseLabel,
	}

	playersList := widget.NewList(
		func() int {
			return len(ui.Players)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, object fyne.CanvasObject) {
			object.(*widget.Label).SetText(ui.Players[id])
		},
	)
	ui.PlayersList = playersList

	monstersList := widget.NewList(
		func() int {
			return len(ui.Monsters)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, object fyne.CanvasObject) {
			object.(*widget.Label).SetText(ui.Monsters[id])
		},
	)
	ui.MonstersList = monstersList

	itemsList := widget.NewList(
		func() int {
			return len(ui.Items)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, object fyne.CanvasObject) {
			object.(*widget.Label).SetText(ui.Items[id])
		},
	)
	ui.ItemsList = itemsList

	selectedMonster := ""
	monstersList.OnSelected = func(id widget.ListItemID) {
		selectedMonster = ui.Monsters[id]
	}

	selectedItem := ""
	itemsList.OnSelected = func(id widget.ListItemID) {
		selectedItem = ui.Items[id]
	}
	fourButtons := container.NewVBox()

	go readServer(
		conn,
		ui,
	)
	//
	// LOOK
	//
	button1 := widget.NewButton("LOOK", func() {
		_, err := fmt.Fprintln(conn, "LOOK")
		if err != nil {
			responseLabel.SetText("Send error: " + err.Error())
		}
	})
	//
	// CONNECT
	//
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Enter player name")

	var button2 *widget.Button

	button2 = widget.NewButton("CONNECT", func() {
		name := strings.TrimSpace(entry.Text)
		if name == "" {
			responseLabel.SetText("Please enter a player name")
			return
		}

		_, err := fmt.Fprintln(conn, "CONNECT "+name)
		if err != nil {
			responseLabel.SetText("Send error: " + err.Error())
			return
		}
		button2.Disable()
		button2.Hide()
		entry.Disable()
		entry.Hide()
	})
	ui.ConnectButton = button2
	//
	// MOVE
	//
	var button3 *widget.Button
	button3 = widget.NewButton("MOVE", func() {
		fourButtons.Objects = nil
		for _, exit := range ui.Exits {
			dir := exit
			btn := widget.NewButton(exit, func() {
				_, err := fmt.Fprintln(conn, "MOVE "+dir)
				if err != nil {
					responseLabel.SetText("Send error: " + err.Error())
				}
				fmt.Fprintln(conn, "LOOK")
				fourButtons.Objects = nil
				fourButtons.Refresh()
				button3.Disable()
			})
			fourButtons.Objects = append(fourButtons.Objects, btn)
		}
		fourButtons.Refresh()
	})
	ui.MoveButton = button3
	//
	// USE
	//
	button4 := widget.NewButton("USE", func() {
		_, err := fmt.Fprintln(conn, "USE "+selectedItem)
		if err != nil {
			responseLabel.SetText("Send error: " + err.Error())
		}
	})
	//
	// TAKE
	//
	button5 := widget.NewButton("TAKE", func() {
		_, err := fmt.Fprintln(conn, "TAKE "+selectedItem)
		if err != nil {
			responseLabel.SetText("Send error: " + err.Error())
		}
	})
	//
	// ATTACK
	//
	button6 := widget.NewButton("ATTACK", func() {
		_, err := fmt.Fprintln(conn, "ATTACK "+selectedMonster)
		if err != nil {
			responseLabel.SetText("Send error: " + err.Error())
		}
	})

	title := canvas.NewText("The Wanderer's Rest", color.White)
	title.TextSize = 24
	title.TextStyle = fyne.TextStyle{Bold: true}
	header := container.NewCenter(title)

	monsterText := canvas.NewText("Monsters", color.White)
	monsterText.TextSize = 18
	monsterText.TextStyle = fyne.TextStyle{Bold: true}

	monsterTitle := container.NewCenter(
		monsterText,
	)

	monsterContainer := container.NewBorder(
		monsterTitle,
		nil,
		nil,
		nil,
		monstersList,
	)

	itemText := canvas.NewText("Items", color.White)
	itemText.TextSize = 18
	itemText.TextStyle = fyne.TextStyle{Bold: true}

	itemTitle := container.NewCenter(
		itemText,
	)

	itemContainer := container.NewBorder(
		itemTitle,
		nil,
		nil,
		nil,
		itemsList,
	)

	playerText := canvas.NewText("Players", color.White)
	playerText.TextSize = 18
	playerText.TextStyle = fyne.TextStyle{Bold: true}

	playerTitle := container.NewCenter(
		playerText,
	)

	playerContainer := container.NewBorder(
		playerTitle,
		nil,
		nil,
		nil,
		playersList,
	)

	midcontainer := container.New(layout.NewGridLayout(3), monsterContainer, itemContainer, playerContainer)

	center_res := container.NewCenter(responseLabel)
	hContainer := container.New(layout.NewCenterLayout(), header)
	// name_container := container.NewVBox(entry)
	buttons := container.New(layout.NewGridLayout(4), button1, button2, button3, button4, button5, button6)
	center_zone := container.NewVBox(center_res, fourButtons, entry)
	content := container.NewBorder(hContainer, buttons, roomDescription, midcontainer, center_zone)
	myWindow.Resize(fyne.NewSize(800, 800))
	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
