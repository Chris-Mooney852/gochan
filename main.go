package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/jroimartin/gocui"
)

type boardResponse struct {
	Boards []board `json:"boards"`
}

type board struct {
	Board       string `json:"board"`
	Title       string `json:"title"`
	Description string `json:"meta_description"`
}

type page struct {
	Threads []thread `json:"threads"`
}

type thread struct {
	No   int    `json:"no"`
	Com  string `json:"com"`
	Sub  string `json:"sub"`
	Name string `json:"name"`
	Now string `json:"now"`
	FileName string `json:"filename"`
	Ext string `json:"ext"`
	Tim int `json:"tim"`
}

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true

	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("boards", 1, 1, maxX/6, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Boards"
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		boards := getBoards()
		for i := 0; i < len(boards); i++ {
			fmt.Fprintln(v, boards[i].Board, "-", boards[i].Title)
		}

		if _, err := g.SetCurrentView("boards"); err != nil {
			return err
		}
	}
	if v, err := g.SetView("threads", maxX/6+1, 1, maxX/6*3, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Threads"
	}
	if v, err := g.SetView("thread", maxX/6*3+1, 1, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Thead"
	}
	return nil
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("boards", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("boards", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("boards", gocui.KeyEnter, gocui.ModNone, selectBoard); err != nil {
		return err
	}

	return nil
}

func selectBoard(g *gocui.Gui, v *gocui.View) error {
	var selected string
	var err error

	_, cy := v.Cursor()
	if selected, err = v.Line(cy); err != nil {
		selected = "Get rekt fag, you did it wrong"
	}

	if err := getCatalog(g, selected); err != nil {
		log.Fatal(err)
	}

	return nil
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func getBoards() []board {
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

	return boardRes.Boards
}

func getCatalog(g *gocui.Gui, selected string) error {
	re, err := regexp.Compile("^\\w+")
	if err != nil {
		log.Fatal(err)
	}

	boardName := string(re.Find([]byte(selected)))

	url := fmt.Sprintf("https://a.4cdn.org/%s/catalog.json", boardName)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var pages []page
	jsonErr := json.Unmarshal(body, &pages)

	if jsonErr != nil {
		log.Fatalln(jsonErr)
	}

	printThreads(g, pages)

	return nil
}

func printThreads(g *gocui.Gui, pages []page) error {
	if v, err := g.View("threads"); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		if _, err := g.SetCurrentView("threads"); err != nil {
			return err
		}
	} else {
		v.Clear()
		for i := 0; i < len(pages); i++ {
			for j := 0; j < len(pages[i].Threads); j++ {
				thread := pages[i].Threads[j];

				fmt.Fprintf(v, "\033[36;1mNo.%d\033[0m\n", thread.No)
				fmt.Fprintf(v, "\033[36;4m%s%s\033[32;1m %s\033[0m %s\n", thread.FileName, thread.Ext, thread.Name, thread.Now)
				fmt.Fprintf(v, "\033[34;1m%s\033[0m\n", thread.Sub)
				fmt.Fprintln(v, thread.Com)

				fmt.Fprintln(v, "")
			}
		}
	}

	return nil
}
