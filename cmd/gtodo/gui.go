package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/y-yagi/gocui"
	tasks "google.golang.org/api/tasks/v1"
)

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, cursorLeft); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, cursorRight); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, complete); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, enter); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlQ, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'j', gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'k', gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'h', gocui.ModNone, cursorLeft); err != nil {
		return err
	}

	return g.SetKeybinding("", 'l', gocui.ModNone, cursorRight)
}

func layout(g *gocui.Gui) error {
	var firstKey string

	maxX, maxY := g.Size()
	if v, err := g.SetView("side", 0, 0, int(0.2*float32(maxX)), maxY-1); err != nil {
		v.Title = "List"
		v.Highlight = true
		v.SelBgColor = gocui.ColorBlue
		v.SelFgColor = gocui.ColorBlack

		for k := range tasksPerList {
			if len(firstKey) == 0 {
				firstKey = k
			}
			fmt.Fprintln(v, k)
		}
	}

	if v, err := g.SetView("main", int(0.2*float32(maxX)), 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Tasks"
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		for _, t := range tasksPerList[firstKey] {
			fmt.Fprintln(v, formatTask(t))
		}
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	if v.Name() == "prompt" {
		g.Cursor = false
		g.SetCurrentView("main")
		return g.DeleteView("prompt")
	}

	return gocui.ErrQuit
}

func enter(g *gocui.Gui, v *gocui.View) error {
	if v.Name() == "prompt" {
		fmt.Printf("Comple Task\n")
		return gocui.ErrQuit
	}

	return nil
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	var err error

	if v == nil {
		if v, err = g.SetCurrentView("main"); err != nil {
			return err
		}
	}

	cx, cy := v.Cursor()
	lineCount := len(strings.Split(v.ViewBuffer(), "\n"))
	if cy+1 == lineCount-2 {
		return nil
	}
	if err := v.SetCursor(cx, cy+1); err != nil {
		ox, oy := v.Origin()
		if err := v.SetOrigin(ox, oy+1); err != nil {
			return err
		}
	}

	if v.Name() == "side" {
		refreshMainView(g, v)
	}

	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	var err error

	if v == nil {
		if v, err = g.SetCurrentView("main"); err != nil {
			return err
		}
	}

	ox, oy := v.Origin()
	cx, cy := v.Cursor()
	if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
		if err := v.SetOrigin(ox, oy-1); err != nil {
			return err
		}
	}

	if v.Name() == "side" {
		refreshMainView(g, v)
	}

	return nil
}

func cursorLeft(g *gocui.Gui, v *gocui.View) error {
	var err error
	if v, err = g.SetCurrentView("side"); err != nil {
		return err
	}

	cx, cy := v.Cursor()
	if err := v.SetCursor(cx, cy); err != nil {
		ox, oy := v.Origin()
		if err := v.SetOrigin(ox, oy); err != nil {
			return err
		}
	}
	return nil
}

func cursorRight(g *gocui.Gui, v *gocui.View) error {
	var err error
	if v, err = g.SetCurrentView("main"); err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	cx, cy := v.Cursor()
	if err := v.SetCursor(cx, cy); err != nil {
		ox, oy := v.Origin()
		if err := v.SetOrigin(ox, oy); err != nil {
			return err
		}
	}
	return nil
}

func refreshMainView(g *gocui.Gui, v *gocui.View) error {
	var l string
	var err error

	mainView, _ := g.View("main")
	_, cy := v.Cursor()

	if l, err = v.Line(cy); err != nil {
		l = ""
	}

	if len(l) != 0 {
		mainView.Clear()
		for _, t := range tasksPerList[l] {
			fmt.Fprintln(mainView, formatTask(t))
		}
	}
	return nil
}

func complete(g *gocui.Gui, v *gocui.View) error {
	if err := createPromptView(g, "Complete Task. OK?"); err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	return nil
}

func formatTask(task *tasks.Task) string {
	result := task.Title

	if task.Due != "" {
		time, _ := time.Parse(time.RFC3339, task.Due)
		result += " | " + time.Format("2006-1-2")
	}
	if len(task.Notes) > 0 {
		result += " | " + task.Notes
	}
	return result
}

func createPromptView(g *gocui.Gui, title string) error {
	tw, th := g.Size()
	_, err := g.SetView("prompt", tw/6, (th/2)-1, (tw*5)/6, (th/2)+1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	setTopWindowTitle(g, "prompt", title)

	g.Cursor = true
	_, err = g.SetCurrentView("prompt")

	return err
}

func setTopWindowTitle(g *gocui.Gui, viewName, title string) {
	v, err := g.View(viewName)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	v.Title = fmt.Sprintf("%v (Ctrl-q to close or Enter to OK)", title)
}
