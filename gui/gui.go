package main

import (
	"fmt"
	"net"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func gui(conn net.Conn, incoming chan <-string) {
	myApp := app.New()
	myWindow := myApp.NewWindow("Grid Layout")

	var data = []string{"goblin", "steven"}

	list := widget.NewList(
		func() int {
			return len(data)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(data[i])
		})
	listheader := widget.NewLabel("Monsters")

	var data2 = []string{"sword", "silver coin"}

	list2 := widget.NewList(
		func() int {
			return len(data2)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(data2[i])
		})
	listheader2 := widget.NewLabel("Items")

	button1 := widget.NewButton("LOOK", func() {
		fmt.Println("LOOK")
	})
	button2 := widget.NewButton("TAKE", func() {
		fmt.Println("TAKE")
	})
	button3 := widget.NewButton("MOVE", func() {
		fmt.Println("MOVE")
	})
	button4 := widget.NewButton("USE", func() {
		fmt.Println("USE")
	})
	button5 := widget.NewButton("INVENTORY", func() {
		fmt.Println("INVENTORY")
	})
	button6 := widget.NewButton("TALK", func() {
		fmt.Println("TALK")
	})
	header := container.NewCenter(
		widget.NewLabel("The Wanderer's Rest"),
	)

	leftContainer := container.NewBorder(listheader, nil, nil, nil, list)
	rightContainer := container.NewBorder(listheader2, nil, nil, nil, list2)
	hContainer := container.New(layout.NewCenterLayout(), header)
	buttons := container.New(layout.NewGridLayout(3), button1, button2, button3, button4, button5, button6)
	content := container.NewBorder(hContainer, buttons, leftContainer, rightContainer, nil)
	myWindow.Resize(fyne.NewSize(800, 800))
	myWindow.SetContent(content)
	// myWindow.SetContent(list)
	myWindow.ShowAndRun()
}
