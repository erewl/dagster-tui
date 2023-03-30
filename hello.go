package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	s "nl/vdb/dagstertui/datastructures"

	"github.com/jroimartin/gocui"
)

var (
	RepositoriesView *gocui.View
	JobsView         *gocui.View
	RunsView         *gocui.View
	KeyMappingsView  *gocui.View
	LaunchRunWindow  *gocui.View

	data *s.Overview

	currentRepositoriesList []string
	currentJobsList         []string

	runs     []s.Run
	runNames []string

	userHomeDir string
)

const (
	REPOSITORIES_VIEW = "repositories"
	JOBS_VIEW         = "jobs"
	RUNS_VIEW         = "runs"
	KEY_MAPPINGS_VIEW = "keymaps"
	LAUNCH_RUN_VIEW   = "launch_run"
)

func FillViewWithItems(v *gocui.View, items []string) {
	v.Clear()

	for _, item := range items {
		fmt.Fprintln(v, item)
	}

}

type ConfigState struct {
	Repositories []string `json:"Repositories"`
}

func LoadStateFromConfig(dir string) {

	// Open our jsonFile
	jsonFile, err := os.Open(fmt.Sprintf("%s/.dagstertui/state.json", dir))
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	var state ConfigState
	json.Unmarshal(byteValue, &state)
}

func main() {

	data = new(s.Overview)
	data.Repositories = make(map[string]*s.RepositoryRepresentation, 0)
	data.Url = ""

	home, err := os.UserHomeDir()
	userHomeDir = home

	// Initialize gocui
	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		panic(err)
	}

	defer g.Close()

	// Set layout function
	g.SetManagerFunc(layout)

	// Set keybindings
	err = setKeybindings(g)
	if err != nil {
		panic(err)
	}

	g.SelFgColor = gocui.ColorGreen | gocui.AttrBold
	g.BgColor = gocui.ColorDefault
	g.Highlight = true

	// called once
	setupViews(g)

	// loading latest saved config
	LoadStateFromConfig(userHomeDir)
	if len(data.Repositories) == 0 {
		data.AppendRepositories(GetRepositories())
	}

	setWindowColors(g, REPOSITORIES_VIEW, "red")
	currentRepositoriesList = data.GetRepositoryNames()
	FillViewWithItems(RepositoriesView, currentRepositoriesList)

	RepositoriesView.SelFgColor = gocui.AttrBold
	RepositoriesView.SelBgColor = gocui.ColorRed
	RepositoriesView.Wrap = true

	JobsView.SelFgColor = gocui.AttrBold
	JobsView.SelBgColor = gocui.ColorRed
	JobsView.Wrap = true

	RunsView.SelFgColor = gocui.AttrBold
	RunsView.SelBgColor = gocui.ColorRed
	RunsView.Wrap = true

	// Start main loop
	err = g.MainLoop()
	if err != nil && err != gocui.ErrQuit {
		panic(err)
	}
}

func setupViews(g *gocui.Gui) error {
	// Set window sizes and positions
	maxX, maxY := g.Size()
	windowWidth := maxX / 3
	windowHeight := maxY - 2
	window1X := 0
	window2X := windowWidth
	window3X := windowWidth * 2

	// Create windows
	var err error
	RepositoriesView, err = g.SetView(REPOSITORIES_VIEW, window1X, 1, window1X+windowWidth, windowHeight+1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	RepositoriesView.Title = "Repositories"

	JobsView, err = g.SetView(JOBS_VIEW, window2X, 1, window2X+windowWidth, windowHeight+1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	JobsView.Title = "Jobs"

	RunsView, err = g.SetView(RUNS_VIEW, window3X, 1, window3X+windowWidth, windowHeight+1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	RunsView.Title = "Runs"

	// Set focus on first window
	if _, err := g.SetCurrentView(REPOSITORIES_VIEW); err != nil {
		panic(err)
	}

	return nil
}

// repeated call with every change (render)
func layout(g *gocui.Gui) error {
	// Set window sizes and positions
	maxX, maxY := g.Size()
	windowWidth := maxX / 3
	windowHeight := maxY - 2
	window1X := 0
	window2X := windowWidth
	window3X := windowWidth * 2

	// Create windows
	if _, err := g.SetView(REPOSITORIES_VIEW, window1X, 1, window1X+windowWidth, windowHeight+1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	if _, err := g.SetView(JOBS_VIEW, window2X, 1, window2X+windowWidth, windowHeight+1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	if _, err := g.SetView(RUNS_VIEW, window3X, 1, window3X+windowWidth, windowHeight+1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}

	return nil
}

func OpenKeyMaps(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := g.Size()

	var err error
	KeyMappingsView, err = g.SetView(KEY_MAPPINGS_VIEW, int(float64(maxX)*0.2), int(float64(maxY)*0.2), int(float64(maxX)*0.8), int(float64(maxY)*0.8))
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	KeyMappingsView.Title = "KeyMaps"

	g.SetCurrentView(KEY_MAPPINGS_VIEW)
	return nil
}

func OpenLaunchWindow(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := g.Size()

	var err error
	LaunchRunWindow, err = g.SetView(LAUNCH_RUN_VIEW, int(float64(maxX)*0.2), int(float64(maxY)*0.2), int(float64(maxX)*0.8), int(float64(maxY)*0.8))
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	LaunchRunWindow.Editable = true
	LaunchRunWindow.Title = "Launch Run For"

	g.SetCurrentView(LAUNCH_RUN_VIEW)
	return nil
}

func ClosePopupView(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView(v.Name()); err != nil {
		return err
	}
	g.SetCurrentView(REPOSITORIES_VIEW)
	return nil
}

func setKeybindings(g *gocui.Gui) error {
	// Set keybindings to switch focus between windows
	if err := g.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, SwitchFocusRight); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, SwitchFocusLeft); err != nil {
		return err
	}

	// Quit
	if err := g.SetKeybinding("", 'q', gocui.ModNone, Quit); err != nil {
		return err
	}
	// Open Controls window
	if err := g.SetKeybinding("", 'x', gocui.ModNone, OpenKeyMaps); err != nil {
		return err
	}
	if err := g.SetKeybinding(KEY_MAPPINGS_VIEW, 'X', gocui.ModNone, ClosePopupView); err != nil {
		return err
	}
	if err := g.SetKeybinding(LAUNCH_RUN_VIEW, 'X', gocui.ModNone, ClosePopupView); err != nil {
		return err
	}

	// define keybindings for moving between items
	if err := g.SetKeybinding(REPOSITORIES_VIEW, gocui.KeyArrowDown, gocui.ModNone, CursorDown); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(REPOSITORIES_VIEW, gocui.KeyArrowUp, gocui.ModNone, CursorUp); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(REPOSITORIES_VIEW, gocui.KeyEnter, gocui.ModNone, LoadJobsForRepository); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding(JOBS_VIEW, gocui.KeyArrowDown, gocui.ModNone, CursorDown); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(JOBS_VIEW, gocui.KeyArrowUp, gocui.ModNone, CursorUp); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(JOBS_VIEW, gocui.KeyEnter, gocui.ModNone, LoadRunsForJob); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(JOBS_VIEW, 'l', gocui.ModNone, OpenLaunchWindow); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding(RUNS_VIEW, gocui.KeyArrowDown, gocui.ModNone, CursorDown); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(RUNS_VIEW, gocui.KeyArrowUp, gocui.ModNone, CursorUp); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(RUNS_VIEW, 'l', gocui.ModNone, OpenLaunchWindow); err != nil {
		panic(err)
	}

	return nil
}

func setWindowColors(g *gocui.Gui, viewName string, bgColor string) error {
	view, err := g.View(viewName)
	if err != nil {
		return err
	}

	if bgColor == "" {
		view.Highlight = false
		view.FgColor = gocui.Attribute(gocui.ColorDefault)
	} else {
		view.Highlight = true
		// view.FgColor = gocui.Attribute(gocui.ColorGreen) | gocui.AttrBold
	}

	return nil
}

func getContentByView(v *gocui.View) []string {

	name := v.Name()
	switch name {
	case REPOSITORIES_VIEW:
		return currentRepositoriesList
	case JOBS_VIEW:
		return currentJobsList
	case RUNS_VIEW:
		return runNames
	}
	return []string{}

}

func LoadJobsForRepository(g *gocui.Gui, v *gocui.View) error {

	locationName := getElementByCursor(v)

	repo := data.GetRepoByLocation(locationName)

	data.AppendJobsToRepository(repo.Location, GetJobsInRepository(repo))

	JobsView.Title = fmt.Sprintf("%s - Jobs", locationName)
	JobsView.Clear()
	currentJobsList = data.GetJobNamesInRepository(locationName)
	FillViewWithItems(JobsView, currentJobsList)

	resetCursor(g, JOBS_VIEW)
	return SetFocus(g, JOBS_VIEW, v.Name())

}

func getElementByCursor(v *gocui.View) string {
	_, oy := v.Origin()
	_, vy := v.Cursor()

	items := getContentByView(v)

	return items[vy+oy]
}

func LoadRunsForJob(g *gocui.Gui, v *gocui.View) error {

	locationName := getElementByCursor(RepositoriesView)
	jobName := getElementByCursor(v)

	repo := data.GetRepoByLocation(locationName)

	pipelineRuns := GetPipelineRuns(repo, jobName, 10)
	data.UpdatePipelineAndRuns(repo.Location, pipelineRuns)
	runNames = data.GetSortedRunNamesFor(locationName, jobName)

	RunsView.Title = fmt.Sprintf("%s - Runs", jobName)
	RunsView.Clear()
	FillViewWithItems(RunsView, runNames)

	resetCursor(g, RUNS_VIEW)
	return SetFocus(g, RUNS_VIEW, v.Name())
}

func resetCursor(g *gocui.Gui, name string) error {
	v, err := g.View(name)
	if err != nil {
		return err
	}
	v.SetCursor(0, 0)
	v.SetOrigin(0, 0)

	return nil
}
