package main

import (
	"encoding/json"
	"fmt"
	"github.com/jroimartin/gocui"
	"io/ioutil"
	s "nl/vdb/dagstertui/datastructures"
	"os"
	"os/exec"
	"runtime"
)

var (
	RepositoriesView *gocui.View
	JobsView         *gocui.View
	RunsView         *gocui.View
	KeyMappingsView  *gocui.View
	LaunchRunWindow  *gocui.View
	FeedbackView     *gocui.View

	data  *s.Overview
	State *ApplicationState

	currentRepositoriesList []string
	currentJobsList         []string

	runNames []string

	userHomeDir string
)

const (
	REPOSITORIES_VIEW = "repositories"
	JOBS_VIEW         = "jobs"
	RUNS_VIEW         = "runs"
	KEY_MAPPINGS_VIEW = "keymaps"
	LAUNCH_RUN_VIEW   = "launch_run"
	FEEDBACK_VIEW     = "feedback"
)

type ApplicationState struct {
	previousActiveWindow string
	selectedRepo         string
	selectedJob          string
	selectedRun          string
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

	State = new(ApplicationState)
	State.previousActiveWindow = ""
	State.selectedRepo = ""
	State.selectedJob = ""
	State.selectedRun = ""

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
	InitializeViews(g)

	// loading latest saved config
	LoadStateFromConfig(userHomeDir)
	if len(data.Repositories) == 0 {
		data.AppendRepositories(GetRepositories())
	}

	SetWindowColors(g, REPOSITORIES_VIEW, "red")
	currentRepositoriesList = data.GetRepositoryNames()
	FillViewWithItems(RepositoriesView, currentRepositoriesList)

	SetViewStyles(RepositoriesView)
	SetViewStyles(JobsView)
	SetViewStyles(RunsView)

	// Start main loop
	err = g.MainLoop()
	if err != nil && err != gocui.ErrQuit {
		panic(err)
	}
}

func SetViewStyles(v *gocui.View) {
	v.SelFgColor = gocui.AttrBold
	v.SelBgColor = gocui.ColorRed
	v.Wrap = true
}

func InitializeViews(g *gocui.Gui) error {
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

func openbrowser(url string) {

	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		panic(err)
	}
}

func OpenInBrowser(g *gocui.Gui, v *gocui.View) error {

	url := "https://dagster.test-backend.vdbinfra.nl"

	switch v.Name() {
	case REPOSITORIES_VIEW:
		// https://dagster.test-backend.vdbinfra.nl/locations/supply_forecasting_repo@supply-forecasting-pr-2168/jobs
		repo := State.selectedRepo
		if repo != "" {
			r := data.Repositories[repo]
			openbrowser(fmt.Sprintf("%s/locations/%s@%s/jobs", url, r.Name, r.Location))
		}
		return nil
	case JOBS_VIEW:
		// https://dagster.test-backend.vdbinfra.nl/locations/supply_forecasting_repo@supply-forecasting-pr-2168/jobs/e2e
		repo := State.selectedRepo
		job := State.selectedJob
		if repo != "" && job != "" {
			r := data.Repositories[repo]
			openbrowser(fmt.Sprintf("%s/locations/%s@%s/jobs/%s", url, r.Name, r.Location, job))
		}
		return nil
	case RUNS_VIEW:
		runId := State.selectedRun
		if runId != "" {
			openbrowser(fmt.Sprintf("%s/runs/%s", url, runId))
		}
		return nil
	default:
		return nil
	}
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
	LaunchRunWindow.Highlight = true
	LaunchRunWindow.SelFgColor = gocui.ColorYellow

	fmt.Fprintln(LaunchRunWindow, data.Repositories[State.selectedRepo].Jobs[State.selectedJob].DefaultRunConfigYaml)

	State.previousActiveWindow = v.Name()
	g.SetCurrentView(LAUNCH_RUN_VIEW)
	return nil
}

func ClosePopupView(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView(v.Name()); err != nil {
		return err
	}
	g.SetCurrentView(State.previousActiveWindow)
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
	if err := g.SetKeybinding("", 'O', gocui.ModNone, OpenInBrowser); err != nil {
		return err
	}
	if err := g.SetKeybinding(KEY_MAPPINGS_VIEW, gocui.KeyCtrlX, gocui.ModNone, ClosePopupView); err != nil {
		return err
	}
	if err := g.SetKeybinding(LAUNCH_RUN_VIEW, gocui.KeyCtrlX, gocui.ModNone, ClosePopupView); err != nil {
		return err
	}
	if err := g.SetKeybinding(LAUNCH_RUN_VIEW, gocui.KeyEnter, gocui.ModNone, ValidateAndLaunchRun); err != nil {
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
	// if err := g.SetKeybinding(RUNS_VIEW, 'i', gocui.ModNone, InspectCurrentRunConfig); err != nil {
	// panic(err)
	// }

	return nil
}

func LoadJobsForRepository(g *gocui.Gui, v *gocui.View) error {

	locationName := GetElementByCursor(v)
	State.selectedRepo = locationName

	repo := data.GetRepoByLocation(locationName)

	data.AppendJobsToRepository(repo.Location, GetJobsInRepository(repo))

	JobsView.Title = fmt.Sprintf("%s - Jobs", locationName)
	JobsView.Clear()
	currentJobsList = data.GetJobNamesInRepository(locationName)
	FillViewWithItems(JobsView, currentJobsList)

	ResetCursor(g, JOBS_VIEW)
	return SetFocus(g, JOBS_VIEW, v.Name())

}

func LoadRunsForJob(g *gocui.Gui, v *gocui.View) error {

	jobName := GetElementByCursor(v)
	State.selectedJob = jobName

	repo := data.GetRepoByLocation(State.selectedRepo)

	pipelineRuns := GetPipelineRuns(repo, State.selectedJob, 10)
	data.UpdatePipelineAndRuns(repo.Location, pipelineRuns)
	runNames = data.GetSortedRunNamesFor(State.selectedRepo, State.selectedJob)

	RunsView.Title = fmt.Sprintf("%s - Runs", State.selectedJob)
	RunsView.Clear()
	FillViewWithItems(RunsView, runNames)

	ResetCursor(g, RUNS_VIEW)
	return SetFocus(g, RUNS_VIEW, v.Name())
}

func ValidateAndLaunchRun(g *gocui.Gui, v *gocui.View) error {

	// runId := LaunchRunForJob(*data.Repositories[State.selectedRepo], State.selectedJob, LaunchRunWindow.BufferLines())
	ClosePopupView(g, LaunchRunWindow)

	return nil
}

func SetWindowColors(g *gocui.Gui, viewName string, bgColor string) error {
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

func FillViewWithItems(v *gocui.View, items []string) {
	v.Clear()

	for _, item := range items {
		fmt.Fprintln(v, item)
	}

}

func GetContentByView(v *gocui.View) []string {

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

func GetElementByCursor(v *gocui.View) string {
	_, oy := v.Origin()
	_, vy := v.Cursor()

	items := GetContentByView(v)

	return items[vy+oy]
}

func ResetCursor(g *gocui.Gui, name string) error {
	v, err := g.View(name)
	if err != nil {
		return err
	}
	v.SetCursor(0, 0)
	v.SetOrigin(0, 0)

	return nil
}
