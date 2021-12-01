package main

import (
	"fmt"
	"log"
	"encoding/json"
	"net/http"
	"io/ioutil"

	"github.com/jroimartin/gocui"
)

type boardResponse struct {
	Boards []board `json:"boards"`
}

type board struct {
	Title string `json:"title"`
}

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("boards", 1, 1, maxX/6, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Boards"
		getBoards()
	}
	if v, err := g.SetView("threads", maxX/6+1, 1, maxX/6*3, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Theads"
	}
	if v, err := g.SetView("thread", maxX/6*3+1, 1, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Thead"
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func getBoards() {
	resp, err := http.Get("https://a.4cdn.org/boards.json")
	if err != nil {
	  log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
	  log.Fatalln(err)
	}

	boardRes := boardResponse{}
	jsonErr := json.Unmarshal(body, &boardRes)
	if jsonErr != nil {
		log.Fatalln(jsonErr)
	}

	fmt.Println(boardRes.Boards)
}
