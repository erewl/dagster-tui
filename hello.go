package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	s "nl/vdb/dagstertui/datastructures"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	c "github.com/jroimartin/gocui"
)

var (
	RepositoriesView *c.View
	JobsView         *c.View
	RunsView         *c.View
	KeyMappingsView  *c.View
	LaunchRunWindow  *c.View
	FeedbackView     *c.View

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
	g, err := c.NewGui(c.Output256)
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

	g.SelFgColor = c.ColorGreen | c.AttrBold
	g.BgColor = c.ColorDefault
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
	if err != nil && err != c.ErrQuit {
		panic(err)
	}
}

func SetViewStyles(v *c.View) {
	v.SelFgColor = c.AttrBold
	v.SelBgColor = c.ColorRed
	v.Wrap = true
}

func InitializeViews(g *c.Gui) error {
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
		if err != c.ErrUnknownView {
			return err
		}
	}
	RepositoriesView.Title = "Repositories"

	JobsView, err = g.SetView(JOBS_VIEW, window2X, 1, window2X+windowWidth, windowHeight+1)
	if err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	JobsView.Title = "Jobs"

	RunsView, err = g.SetView(RUNS_VIEW, window3X, 1, window3X+windowWidth, windowHeight+1)
	if err != nil {
		if err != c.ErrUnknownView {
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
func layout(g *c.Gui) error {
	// Set window sizes and positions
	maxX, maxY := g.Size()
	windowWidth := maxX / 3
	windowHeight := maxY - 2
	window1X := 0
	window2X := windowWidth
	window3X := windowWidth * 2

	// Create windows
	if _, err := g.SetView(REPOSITORIES_VIEW, window1X, 1, window1X+windowWidth, windowHeight+1); err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	if _, err := g.SetView(JOBS_VIEW, window2X, 1, window2X+windowWidth, windowHeight+1); err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	if _, err := g.SetView(RUNS_VIEW, window3X, 1, window3X+windowWidth, windowHeight+1); err != nil {
		if err != c.ErrUnknownView {
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

func OpenInBrowser(g *c.Gui, v *c.View) error {

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
			openbrowser(fmt.Sprintf("%s/locations/%s@%s/jobs/%s/playground", url, r.Name, r.Location, job))
		}
		return nil
	case RUNS_VIEW:
		runId := State.selectedRun
		if runId == "" {
			runId = GetElementByCursor(RunsView)
		}
		if runId != "" {
			openbrowser(fmt.Sprintf("%s/runs/%s", url, runId))
		}
		return nil
	default:
		return nil
	}
}

func OpenKeyMaps(g *c.Gui, v *c.View) error {
	maxX, maxY := g.Size()

	var err error
	KeyMappingsView, err = g.SetView(KEY_MAPPINGS_VIEW, int(float64(maxX)*0.2), int(float64(maxY)*0.2), int(float64(maxX)*0.8), int(float64(maxY)*0.8))
	if err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	KeyMappingsView.Title = "KeyMaps"

	g.SetCurrentView(KEY_MAPPINGS_VIEW)
	return nil
}

var DefaultEditor c.Editor = c.EditorFunc(simpleEditor)

func simpleEditor(v *c.View, key c.Key, ch rune, mod c.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == c.KeyArrowDown:
		v.MoveCursor(0, 1, true)
	case key == c.KeyArrowUp:
		v.MoveCursor(0, -1, true)
	case key == c.KeyArrowLeft:
		v.MoveCursor(-1, 0, true)
	case key == c.KeyArrowRight:
		v.MoveCursor(1, 0, true)
	case key == '/' && mod == 1:
		x, y := v.Cursor()
		fmt.Print(x, y)
		currentLine, _ := v.Line(y)

		re := regexp.MustCompile(`^[\s]*#`) // any amount of whitespaces followed by # at the beginning of a line
		if re.MatchString(currentLine) {
			index := strings.Index(currentLine, "#")
			v.SetCursor(index, y)
			v.EditDelete(false)
		} else {
			v.SetCursor(0, y)
			v.EditWrite('#')
		}
		v.SetCursor(x, y)
	case key == c.KeySpace:
		v.EditWrite(' ')
	case key == c.KeyBackspace || key == c.KeyBackspace2:
		v.EditDelete(true)
		// ...
	}
}

func OpenLaunchWindow(g *c.Gui, v *c.View) error {
	maxX, maxY := g.Size()

	var err error
	LaunchRunWindow, err = g.SetView(LAUNCH_RUN_VIEW, int(float64(maxX)*0.2), int(float64(maxY)*0.2), int(float64(maxX)*0.8), int(float64(maxY)*0.8))
	if err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	LaunchRunWindow.Editable = true
	LaunchRunWindow.Editor = DefaultEditor
	LaunchRunWindow.Title = "Launch Run For"
	LaunchRunWindow.Highlight = true
	LaunchRunWindow.SelBgColor = c.ColorBlue

	fmt.Fprintln(LaunchRunWindow, data.Repositories[State.selectedRepo].Jobs[State.selectedJob].DefaultRunConfigYaml)

	State.previousActiveWindow = v.Name()
	g.SetCurrentView(LAUNCH_RUN_VIEW)
	return nil
}

func ClosePopupView(g *c.Gui, v *c.View) error {
	if err := g.DeleteView(v.Name()); err != nil {
		return err
	}
	g.SetCurrentView(State.previousActiveWindow)
	return nil
}

func setKeybindings(g *c.Gui) error {
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
	if err := g.SetKeybinding("", 'x', c.ModNone, OpenKeyMaps); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'O', c.ModNone, OpenInBrowser); err != nil {
		return err
	}
	if err := g.SetKeybinding(KEY_MAPPINGS_VIEW, c.KeyCtrlX, c.ModNone, ClosePopupView); err != nil {
		return err
	}
	if err := g.SetKeybinding(LAUNCH_RUN_VIEW, c.KeyCtrlX, c.ModNone, ClosePopupView); err != nil {
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

	if err := g.SetKeybinding(JOBS_VIEW, c.KeyArrowDown, c.ModNone, CursorDown); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(JOBS_VIEW, c.KeyArrowUp, c.ModNone, CursorUp); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(JOBS_VIEW, c.KeyEnter, c.ModNone, LoadRunsForJob); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(JOBS_VIEW, 'l', c.ModNone, OpenLaunchWindow); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding(RUNS_VIEW, c.KeyArrowDown, c.ModNone, CursorDown); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(RUNS_VIEW, c.KeyArrowUp, c.ModNone, CursorUp); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(RUNS_VIEW, 'l', c.ModNone, OpenLaunchWindow); err != nil {
		panic(err)
	}
	// if err := g.SetKeybinding(RUNS_VIEW, 'i', c.ModNone, InspectCurrentRunConfig); err != nil {
	// panic(err)
	// }

	return nil
}

func LoadJobsForRepository(g *c.Gui, v *c.View) error {

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

func LoadRunsForJob(g *c.Gui, v *c.View) error {

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

func ValidateAndLaunchRun(g *c.Gui, v *c.View) error {

	// runId := LaunchRunForJob(*data.Repositories[State.selectedRepo], State.selectedJob, LaunchRunWindow.BufferLines())
	LaunchRunForJob(*data.Repositories[State.selectedRepo], State.selectedJob, LaunchRunWindow.BufferLines())
	ClosePopupView(g, LaunchRunWindow)

	return nil
}

func SetWindowColors(g *c.Gui, viewName string, bgColor string) error {
	view, err := g.View(viewName)
	if err != nil {
		return err
	}

	if bgColor == "" {
		view.Highlight = false
		view.FgColor = c.Attribute(c.ColorDefault)
	} else {
		view.Highlight = true
		// view.FgColor = c.Attribute(c.ColorGreen) | c.AttrBold
	}

	return nil
}

func FillViewWithItems(v *c.View, items []string) {
	v.Clear()

	for _, item := range items {
		fmt.Fprintln(v, item)
	}

}

func GetContentByView(v *c.View) []string {

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

func GetElementByCursor(v *c.View) string {
	_, oy := v.Origin()
	_, vy := v.Cursor()

	items := GetContentByView(v)

	return items[vy+oy]
}

func ResetCursor(g *c.Gui, name string) error {
	v, err := g.View(name)
	if err != nil {
		return err
	}
	v.SetCursor(0, 0)
	v.SetOrigin(0, 0)

	return nil
}
