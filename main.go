package main

import (
	"fmt"
	"github.com/aquasecurity/table"
	"github.com/eiannone/keyboard"
	"github.com/godbus/dbus/v5"
	"os"
	"time"
)

func main() {

	c := make(chan *dbus.Message)
	clear := make(chan bool)

	go readKeyboardEvents(clear)
	go listenNotifications(c)
	go printNotifications(c, clear)
	select {}
}

func readKeyboardEvents(ch chan<- bool) {
	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer func() {
		_ = keyboard.Close()
	}()

	fmt.Println("Press ESC to quit")
	fmt.Println("Press 'c' to clear")
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		}
		if key == keyboard.KeyEsc {
			os.Exit(0)
			keyboard.Close()
			break
		} else if char == 'c' {
			ch <- true
		}
	}
}

func listenNotifications(ch chan *dbus.Message) {

	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		os.Exit(1)
	}
	defer conn.Close()

	call := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		"eavesdrop='true',path='/org/freedesktop/Notifications',type='method_call',member='Notify'")
	if call.Err != nil {
		fmt.Fprintln(os.Stderr, "Failed to add match:", call.Err)
		os.Exit(1)
	}
	conn.Eavesdrop(ch)
	select {}
}

func printNotifications(ch chan *dbus.Message, clear chan bool) {
	tab := table.New(os.Stdout)
	tab.SetRowLines(true)
	tab.SetHeaders("Notification", "Sender", "Raised at")
	tab.SetHeaderStyle(table.StyleBold)
	tab.SetDividers(table.UnicodeRoundedDividers)
	tab.SetFooters("Notification", "Sender", "Raised at")

	for {
		select {
		case v := <-ch:
			fmt.Println("\033[2J")
			tab.AddRow(v.Body[3].(string), v.Body[0].(string), time.Now().Format(time.RFC1123))
			tab.Render()
		case <-clear:
			fmt.Println("\033[2J")
			fmt.Println("Clearing")
			tab.Clear()
		}
	}
}
