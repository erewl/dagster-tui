package main

import (
	"encoding/json"
	"flag"
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
	FilterView       *c.View

	data  *s.Overview
	State *ApplicationState
	conf  Config

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
	FILTER_VIEW       = "filter"
)

type ApplicationState struct {
	previousActiveWindow string
	selectedRepo         string
	selectedJob          string
	selectedRun          string

	repoFilter string
}

type Config struct {
	Environments map[string]string `json:"environments"`
}

func LoadConfig(dir string) {

	// Open our jsonFile
	jsonFile, err := os.Open(fmt.Sprintf("%s/.dagstertui/config.json", dir))
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
	json.Unmarshal(byteValue, &conf)
}

func main() {

	home, err := os.UserHomeDir()
	userHomeDir = home
	LoadConfig(home)

	environmentFlag := flag.String("e", "test", "sets the home url of the dagster environment")

	// Parse the command-line arguments to set the value of exampleFlag
	flag.Parse()

	data = new(s.Overview)
	data.Repositories = make(map[string]*s.RepositoryRepresentation, 0)
	data.Url = conf.Environments[*environmentFlag]
	DagsterGraphQL = fmt.Sprintf("%s/graphql", data.Url)

	State = &ApplicationState{
		previousActiveWindow: "",
		selectedRepo:         "",
		selectedJob:          "",
		selectedRun:          "",
		repoFilter:           "",
	}

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

	SetWindowColors(g, REPOSITORIES_VIEW, "red")

	data.AppendRepositories(LoadRepositories())
	
	currentRepositoriesList = data.GetRepositoryNames()
	FillViewWithItems(RepositoriesView, currentRepositoriesList)

	SetViewStyles(RepositoriesView)
	SetViewStyles(JobsView)
	SetViewStyles(RunsView)
	FilterView.Editable = true
	FilterView.Editor = DefaultEditor
	FilterView.Title = "Filter"

	// OpenLaunchWindow(g, JobsView)

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

	yOffset := 5

	// Create windows
	var err error
	RepositoriesView, err = g.SetView(REPOSITORIES_VIEW, window1X, yOffset, window1X+windowWidth, windowHeight+yOffset)
	if err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	RepositoriesView.Title = "Repositories"

	JobsView, err = g.SetView(JOBS_VIEW, window2X, yOffset, window2X+windowWidth, windowHeight+yOffset)
	if err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	JobsView.Title = "Jobs"

	RunsView, err = g.SetView(RUNS_VIEW, window3X, yOffset, window3X+windowWidth, windowHeight+yOffset)
	if err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	RunsView.Title = "Runs"

	FilterView, err = g.SetView(FILTER_VIEW, 0, 0, maxX, 1)
	if err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	FilterView.Title = "Filter"

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

	yOffset := 3

	// Create windows
	if _, err := g.SetView(REPOSITORIES_VIEW, window1X, yOffset, window1X+windowWidth, windowHeight+1); err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	if _, err := g.SetView(JOBS_VIEW, window2X, yOffset, window2X+windowWidth, windowHeight+1); err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	if _, err := g.SetView(RUNS_VIEW, window3X, yOffset, window3X+windowWidth, windowHeight+1); err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	if _, err := g.SetView(FILTER_VIEW, 0, 0, maxX, yOffset-1); err != nil {
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


	switch v.Name() {
	case REPOSITORIES_VIEW:
		repo := State.selectedRepo
		if repo != "" {
			r := data.Repositories[repo]
			openbrowser(fmt.Sprintf("%s/locations/%s@%s/jobs", data.Url, r.Name, r.Location))
		}
		return nil
	case JOBS_VIEW:
		repo := State.selectedRepo
		job := State.selectedJob
		if repo != "" && job != "" {
			r := data.Repositories[repo]
			openbrowser(fmt.Sprintf("%s/locations/%s@%s/jobs/%s/playground", data.Url, r.Name, r.Location, job))
		}
		return nil
	case RUNS_VIEW:
		runId := State.selectedRun
		if runId == "" {
			runId = GetElementByCursor(RunsView)
		}
		if runId != "" {
			openbrowser(fmt.Sprintf("%s/runs/%s", data.Url, runId))
		}
		return nil
	default:
		return nil
	}
}

func OpenPopupKeyMaps(g *c.Gui, v *c.View) error {
	maxX, maxY := g.Size()

	var err error
	KeyMappingsView, err = g.SetView(KEY_MAPPINGS_VIEW, int(float64(maxX)*0.2), 1, int(float64(maxX)*0.8), maxY+1)
	if err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	KeyMappingsView.Clear()
	KeyMappingsView.Title = "KeyMaps"

	State.previousActiveWindow = v.Name()
	g.SetCurrentView(KEY_MAPPINGS_VIEW)

	fmt.Fprint(KeyMappingsView, s.KeyMap)
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

func OpenPopupLaunchWindow(g *c.Gui, v *c.View) error {
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

	// x,y := LaunchRunWindow.Cursor()
	// currentLine, _ := LaunchRunWindow.Line(y)

	// 	sampleYaml := `execution:
	//   config:
	//     job_namespace: dagster
	//     max_concurrent: 20
	//     service_account_name: dagster
	// ops:
	//   op_get_e2e_input:
	//     config:
	//       test_selection:
	//       - AllocationDataAddNewDeterministicAsset
	//       # - AllocationDataCheckMaxCapacity
	//       # - AllocationDataCheckPreviousDaysProductionActivity
	//       # - AllocationDataCheckProductionDates
	//       # - AllocationDataChecksumModelRegistry
	//       # - AllocationDataComputeEstimatedModelType
	//       # - AllocationDataMonitorAllocation
	//       # - AllocationDataSyncAllocationCassandraEcedo
	//       # - AllocationDataUpdateCorrectConnectionProperties
	//       # - Evaluate
	//       # - PredictBio
	//       # - PredictDetSolar
	//       # - PredictDetWind
	//       # - PredictMlSolar
	//       # - PredictMlWind
	//       # - PredictSolarKNMI10MinActuals
	//       # - PredictWindKNMI10MinActuals
	// resources:
	//   io_manager:
	//     config:
	//       s3_bucket: vdb-app-dagster-test-iomanager
	//       s3_prefix: e2e
	//   slack:
	//     config:
	//       token:
	//         env: SLACK_BOT_TOKEN`
	// fmt.Fprintln(LaunchRunWindow, sampleYaml)

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
	if err := g.SetKeybinding("", 'x', c.ModNone, OpenPopupKeyMaps); err != nil {
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

	if err := g.SetKeybinding(RUNS_VIEW, c.KeyArrowDown, c.ModNone, CursorDown); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(RUNS_VIEW, c.KeyArrowUp, c.ModNone, CursorUp); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(RUNS_VIEW, 'l', c.ModNone, OpenPopupLaunchWindow); err != nil {
		panic(err)
	}
	// if err := g.SetKeybinding(RUNS_VIEW, 'i', c.ModNone, InspectCurrentRunConfig); err != nil {
	// panic(err)
	// }
	if err := g.SetKeybinding(FILTER_VIEW, c.KeyEnter, c.ModNone, FilterItemsInView); err != nil {
		panic(err)
	}

	return nil
}

func FilterItemsInView(g *c.Gui, v *c.View) error {
	g.SetCurrentView(State.previousActiveWindow)
	switch State.previousActiveWindow {
	case REPOSITORIES_VIEW:
		filterTerm := FilterView.BufferLines()[0]
		cond_contains_term := func(s string) bool { return strings.Contains(s, filterTerm) }
		currentRepositoriesList = filter(data.GetRepositoryNames(), cond_contains_term)

		FillViewWithItems(RepositoriesView, currentRepositoriesList)
	default:
		return nil

	}

	return nil
}

func SwitchToFilterView(g *c.Gui, v *c.View) error {
	State.previousActiveWindow = v.Name()
	g.SetCurrentView(FILTER_VIEW)
	FilterView.Title = fmt.Sprintf("Filter %s", v.Title)
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
