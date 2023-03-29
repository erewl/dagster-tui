package main

import (
	"github.com/jroimartin/gocui"
)

func Quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	items := getContentByView(v)
	cx, cy := v.Cursor()
	_, h := v.Size()
	var height = h
	if h > len(items) {
		height = len(items)
	}

	if cy < height-1 {
		cy++
		if err := v.SetCursor(cx, cy); err != nil {
			return err
		}

	} else {
		oX, oY := v.Origin()
		if cy >= height-1 {
			v.SetOrigin(oX, oY+1)
		}
		if cy+oY > len(items)-2 {
			v.SetOrigin(0, 0)
			v.SetCursor(0, 0)
		}

	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	items := getContentByView(v)
	cx, cy := v.Cursor()
	_, h := v.Size()
	oX, oY := v.Origin()

	var height = h
	if h > len(items) {
		height = len(items)
	}

	if cy > 0 {
		cy--
		if err := v.SetCursor(cx, cy); err != nil {
			return err
		}

	} else {
		if oY > 0 {
			v.SetOrigin(oX, oY-1)
		}
		if cy+oY <= 0 {
			v.SetOrigin(oX, len(items)-height)
			v.SetCursor(cx, height-1)
		}

	}
	return nil
}


func setFocus(g *gocui.Gui, newViewName string, oldViewName string) error {
	// Set focus on next view
	_, err := g.SetCurrentView(newViewName)
	if err != nil {
		return err
	}

	// Set background color of active window to red, and background color of inactive windows to default
	if err := setWindowColors(g, newViewName, "red"); err != nil {
		return err
	}
	if err := setWindowColors(g, oldViewName, ""); err != nil {
		return err
	}

	return nil
}

func switchFocusRight(g *gocui.Gui, v *gocui.View) error {
	// Get current view name
	currentViewName := v.Name()

	// Get next view name
	nextViewName := ""
	switch currentViewName {
	case REPOSITORIES_VIEW:
		nextViewName = JOBS_VIEW
	case JOBS_VIEW:
		nextViewName = RUNS_VIEW
	case RUNS_VIEW:
		nextViewName = REPOSITORIES_VIEW
	case KEY_MAPPINGS_VIEW:
		nextViewName = KEY_MAPPINGS_VIEW
	}

	return setFocus(g, nextViewName, currentViewName)
}

func switchFocusLeft(g *gocui.Gui, v *gocui.View) error {
	// Get current view name
	currentViewName := v.Name()

	// Get previous view name
	previousViewName := ""
	switch currentViewName {
	case REPOSITORIES_VIEW:
		previousViewName = RUNS_VIEW
	case JOBS_VIEW:
		previousViewName = REPOSITORIES_VIEW
	case RUNS_VIEW:
		previousViewName = JOBS_VIEW
	case KEY_MAPPINGS_VIEW:
		previousViewName = KEY_MAPPINGS_VIEW
	}

	return setFocus(g, previousViewName, currentViewName)
}