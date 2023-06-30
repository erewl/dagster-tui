
package app

import (
	c "github.com/jroimartin/gocui"
)

// repeated call with every change (render)
func Layout(g *c.Gui) error {
	// Set window sizes and positions
	maxX, maxY := g.Size()
	windowWidth := maxX / 2
	windowHeight := maxY - 2
	window1X := 0
	window2X := windowWidth / 2
	window3X := windowWidth

	yOffset := 3

	// most left
	RepoWindow.Base.RenderView(g, window1X, yOffset, window1X+windowWidth/2, windowHeight+1)

	// left of REPOSITORIES_VIEW
	JobsWindow.Base.RenderView(g, window2X, yOffset, window2X+windowWidth/2, windowHeight+1)

	// left of JOBS_VIEW /  most right
	RunsWindow.Base.RenderView(g, window3X, yOffset, window3X+windowWidth, int(2.0*(windowHeight/3.0)))

	// below RUNS_VIEW
	RunInfoWindow.Base.RenderView(g, window3X, int(2.0*(windowHeight/3.0))+1, window3X+windowWidth, windowHeight+1)

	// top right corner
	// TODO ok for now, but could be more content-agnostic
	EnvironmentInfoView.Base.RenderView(g, window3X+windowWidth-len(Overview.Url)-1, 0, int(float64(window3X+windowWidth)), yOffset-1)

	// on top of REPOSITORIES_VIEW
	FilterView.Base.RenderView(g, 0, 0, window1X+windowWidth/2, yOffset-1)

	return nil
}

func SetKeybindings(g *c.Gui) error {
	// Set keybindings to switch focus between windows
	if err := g.SetKeybinding("", c.KeyArrowRight, c.ModNone, SwitchFocusRight); err != nil {
		return err
	}
	if err := g.SetKeybinding("", c.KeyArrowLeft, c.ModNone, SwitchFocusLeft); err != nil {
		return err
	}

	// Quit
	if err := g.SetKeybinding("", 'q', c.ModNone, Quit); err != nil {
		return err
	}
	// Open Controls window
	if err := g.SetKeybinding("", 'x', c.ModNone, OpenPopupKeyMaps); err != nil {
		return err
	}

	if err := g.SetKeybinding("", 'O', c.ModNone, OpenInBrowser); err != nil {
		return err
	}
	if err := g.SetKeybinding(KEY_MAPPINGS_VIEW, c.KeyEsc, c.ModNone, ClosePopupView); err != nil {
		return err
	}
	if err := g.SetKeybinding(LAUNCH_RUN_VIEW, c.KeyEsc, c.ModNone, ClosePopupView); err != nil {
		return err
	}
	if err := g.SetKeybinding(LAUNCH_RUN_VIEW, c.KeyCtrlL, c.ModNone, ValidateAndLaunchRun); err != nil {
		return err
	}

	// define keybindings for moving between items
	if err := g.SetKeybinding(REPOSITORIES_VIEW, c.KeyArrowDown, c.ModNone, CursorDown); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(REPOSITORIES_VIEW, c.KeyArrowUp, c.ModNone, CursorUp); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(REPOSITORIES_VIEW, c.KeyEnter, c.ModNone, LoadJobsForRepository); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(REPOSITORIES_VIEW, 'f', c.ModNone, SwitchToFilterView); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding(JOBS_VIEW, c.KeyArrowDown, c.ModNone, CursorDown); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(JOBS_VIEW, c.KeyArrowUp, c.ModNone, CursorUp); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(JOBS_VIEW, c.KeyEnter, c.ModNone, LoadRunsForJob); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(JOBS_VIEW, 'l', c.ModNone, OpenPopupLaunchWindow); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding(RUNS_VIEW, c.KeyArrowDown, c.ModNone, CursorDownAndUpdateRunInfo); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(RUNS_VIEW, c.KeyArrowUp, c.ModNone, CursorUpAndUpdateRunInfo); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(RUNS_VIEW, 'l', c.ModNone, OpenPopupLaunchWindow); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(RUNS_VIEW, 't', c.ModNone, ShowTerminationOptions); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(RUNS_VIEW, 'T', c.ModNone, TerminateRunByRunId); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(RUNS_VIEW, 'i', c.ModNone, GetLogsByRunId); err != nil {
		panic(err)
	}
	// if err := g.SetKeybinding(RUNS_VIEW, 'i', c.ModNone, InspectCurrentRunConfig); err != nil {
	// panic(err)
	// }
	// if err := g.SetKeybinding(FILTER_VIEW, c.KeyEnter, c.ModNone, FilterItemsInView); err != nil {
	// 	panic(err)
	// }
	if err := g.SetKeybinding(FILTER_VIEW, c.KeyArrowDown, c.ModNone, SwitchFocusDown); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding(CONFIRMATION_VIEW, c.KeyArrowDown, c.ModNone, CursorDown); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(CONFIRMATION_VIEW, c.KeyArrowUp, c.ModNone, CursorUp); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(CONFIRMATION_VIEW, c.KeyEsc, c.ModNone, ClosePopupView); err != nil {
		return err
	}
	if err := g.SetKeybinding(CONFIRMATION_VIEW, c.KeyEnter, c.ModNone, TerminateRunWithConfirmationByRunId); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(FEEDBACK_VIEW, c.KeyEsc, c.ModNone, ClosePopupView); err != nil {
		return err
	}

	return nil
}