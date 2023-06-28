package app

import (
	c "github.com/jroimartin/gocui"
	// l "nl/vdb/dagstertui/log"
)

func Quit(g *c.Gui, v *c.View) error {
	return c.ErrQuit
}

func CursorDown(g *c.Gui, v *c.View) error {
	items := GetContentByView(v)

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

func CursorUp(g *c.Gui, v *c.View) error {
	items := GetContentByView(v)
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

func SetFocus(g *c.Gui, newViewName string, oldViewName string) error {
	// Set focus on next view
	_, err := g.SetCurrentView(newViewName)
	if err != nil {
		return err
	}

	// Set background color of active window to red, and background color of inactive windows to default
	if err := SetWindowColors(g, newViewName, "red"); err != nil {
		return err
	}
	if err := SetWindowColors(g, oldViewName, ""); err != nil {
		return err
	}

	return nil
}

func SwitchFocusRight(g *c.Gui, v *c.View) error {
	// Get current view name
	currentViewName := v.Name()
	State.PreviousActiveWindow = currentViewName

	// Get next view name
	nextViewName := ""
	switch currentViewName {
	case REPOSITORIES_VIEW:
		nextViewName = JOBS_VIEW
	case JOBS_VIEW:
		nextViewName = RUNS_VIEW
	case RUNS_VIEW:
		nextViewName = REPOSITORIES_VIEW
	default:
		nextViewName = currentViewName
	}

	return SetFocus(g, nextViewName, currentViewName)
}

func SwitchFocusLeft(g *c.Gui, v *c.View) error {
	// Get current view name
	currentViewName := v.Name()
	State.PreviousActiveWindow = currentViewName

	// Get previous view name
	previousViewName := ""
	switch currentViewName {
	case REPOSITORIES_VIEW:
		previousViewName = RUNS_VIEW
	case JOBS_VIEW:
		previousViewName = REPOSITORIES_VIEW
	case RUNS_VIEW:
		previousViewName = JOBS_VIEW
	default:
		previousViewName = currentViewName
	}

	return SetFocus(g, previousViewName, currentViewName)
}
