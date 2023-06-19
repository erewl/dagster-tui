package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	s "nl/vdb/dagstertui/internal"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	c "github.com/jroimartin/gocui"
)

type TransformFunc[T any] func(T) string

type Window[T any] struct {
	View              *c.View
	StartX, StartY    int
	EndX, EndY        int
	Title             string
	Elements          []string
	RawElements       []T
	TransformRawToStr TransformFunc[T]
}

func (w *Window[T]) RenderView(g *c.Gui) error {
	_, err := g.SetView(w.View.Name(), w.StartX, w.StartY, w.EndX, w.EndY)
	w.View.Title = w.Title
	return err
}

func (w *Window[T]) SetView(g *c.Gui, viewName string) error {
	tempView, err := g.SetView(viewName, 0, 0, 1, 1)
	if err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	w.View = tempView
	return nil
}

// Replacing FillViewWithItems
func (w *Window[T]) RenderItems(items []T) {
	w.RawElements = items
	w.Elements = make([]string, 0)
	w.View.Clear()
	for _, item := range items {
		itemStr := w.TransformRawToStr(item)
		w.Elements = append(w.Elements, itemStr)
		fmt.Fprintln(w.View, itemStr)
	}
}

var (
	// RepositoriesView    *c.View
	JobsView            *c.View
	RunsView            *c.View
	RunInfoView         *c.View
	KeyMappingsView     *c.View
	LaunchRunWindow     *c.View
	FeedbackView        *c.View
	FilterView          *c.View
	EnvironmentInfoView *c.View

	RepoWindow *Window[s.RepositoryRepresentation]

	Overview *s.Overview
	State    *ApplicationState
	Conf     Config

	CurrentJobsList []string

	runInfos []string

	userHomeDir string

	Client *GraphQLClient
)

const (
	// Views
	REPOSITORIES_VIEW = "repositories"
	JOBS_VIEW         = "jobs"
	RUNS_VIEW         = "runs"
	RUN_INFO_VIEW     = "run_info"
	KEY_MAPPINGS_VIEW = "keymaps"
	LAUNCH_RUN_VIEW   = "launch_run"
	FILTER_VIEW       = "filter"
	ENVIRONMENT_INFO  = "environment"

	FEEDBACK_VIEW     = "feedback"
	CONFIRMATION_VIEW = "confirmation"
)

type ApplicationState struct {
	PreviousActiveWindow string
	SelectedRepo         string
	SelectedJob          string
	SelectedRun          string

	RepoFilter string
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
	json.Unmarshal(byteValue, &Conf)
}

func SetViewStyles(v *c.View) {
	v.SelFgColor = c.AttrBold
	v.SelBgColor = c.ColorRed
	v.Wrap = true
}

func initializeView(g *c.Gui, viewRep **c.View, viewName string, viewTitle string) error {
	view, err := g.SetView(viewName, 0, 0, 1, 1)
	if err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	view.Title = viewTitle

	*viewRep = view
	return nil
}

func RepoToStr(a s.RepositoryRepresentation) string {
	return a.Name
}

func InitializeViews(g *c.Gui) error {
	// Create windows, position is irrelevant
	RepoWindow = &Window[s.RepositoryRepresentation]{
		Title:             "Repositories",
		TransformRawToStr: RepoToStr,
	}
	RepoWindow.SetView(g, REPOSITORIES_VIEW)

	// initializeView(g, &RepositoriesView, REPOSITORIES_VIEW, "Repositories")
	initializeView(g, &JobsView, JOBS_VIEW, "Jobs")
	initializeView(g, &RunsView, RUNS_VIEW, "Runs")
	initializeView(g, &RunInfoView, RUN_INFO_VIEW, "Run Info")
	initializeView(g, &FilterView, FILTER_VIEW, "Filter")
	initializeView(g, &EnvironmentInfoView, ENVIRONMENT_INFO, "Info")

	// Set focus on main window
	if _, err := g.SetCurrentView(REPOSITORIES_VIEW); err != nil {
		panic(err)
	}

	return nil
}

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

	// Create windows
	// most left
	RepoWindow.StartX, RepoWindow.StartY, RepoWindow.EndX, RepoWindow.EndY = window1X, yOffset, window1X+windowWidth/2, windowHeight+1
	RepoWindow.RenderView(g)

	// on top of REPOSITORIES_VIEW
	if _, err := g.SetView(FILTER_VIEW, 0, 0, int(float64(window1X+windowWidth)), yOffset-1); err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	// left of REPOSITORIES_VIEW
	if _, err := g.SetView(JOBS_VIEW, window2X, yOffset, window2X+windowWidth/2, windowHeight+1); err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	// most right
	if _, err := g.SetView(RUNS_VIEW, window3X, yOffset, window3X+windowWidth, int(2.0*(windowHeight/3.0))); err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	// underneath RUNS_VIEW
	if _, err := g.SetView(RUN_INFO_VIEW, window3X, int(2.0*(windowHeight/3.0))+1, window3X+windowWidth, windowHeight+1); err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	// TODO ok for now, but could be more content-agnostic
	// top right corner
	if _, err := g.SetView(ENVIRONMENT_INFO, window3X+windowWidth-len(Overview.Url)-1, 0, int(float64(window3X+windowWidth)), yOffset-1); err != nil {
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
		repo := State.SelectedRepo
		if repo != "" {
			r := Overview.Repositories[repo]
			openbrowser(fmt.Sprintf("%s/locations/%s@%s/jobs", Overview.Url, r.Name, r.Location))
		}
		return nil
	case JOBS_VIEW:
		repo := State.SelectedRepo
		job := State.SelectedJob
		if repo != "" && job != "" {
			r := Overview.Repositories[repo]
			openbrowser(fmt.Sprintf("%s/locations/%s@%s/jobs/%s/playground", Overview.Url, r.Name, r.Location, job))
		}
		return nil
	case RUNS_VIEW:
		runId := State.SelectedRun
		if runId == "" {
			runId = Overview.FindRunIdBySubstring(State.SelectedRepo, State.SelectedJob, GetElementByCursor(v)).RunId
		}
		if runId != "" {
			openbrowser(fmt.Sprintf("%s/runs/%s", Overview.Url, runId))
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

	State.PreviousActiveWindow = v.Name()
	g.SetCurrentView(KEY_MAPPINGS_VIEW)

	fmt.Fprint(KeyMappingsView, s.KeyMap)
	return nil
}

var DefaultEditor c.Editor = c.EditorFunc(simpleEditor)

func simpleEditor(v *c.View, key c.Key, ch rune, mod c.Modifier) {
	switch {
	case key == c.KeyArrowDown:
		v.MoveCursor(0, 1, false)
	case key == c.KeyArrowUp:
		v.MoveCursor(0, -1, false)
	case key == c.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
	case key == c.KeyArrowRight:
		v.MoveCursor(1, 0, false)
	case key == c.KeyCtrlSlash:
		x, y := v.Cursor()
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
		v.SetCursor(x, y+1)
	case key == c.KeySpace:
		v.EditWrite(' ')
	case key == c.KeyBackspace || key == c.KeyBackspace2:
		v.EditDelete(true)
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
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
	LaunchRunWindow.SetCursor(0, 0)

	runConfig := ""
	if v.Name() == RUNS_VIEW {
		SelectedRun := GetElementByCursor(v)
		runConfig = Overview.FindRunIdBySubstring(State.SelectedRepo, State.SelectedJob, SelectedRun).RunconfigYaml
	} else {
		runConfig = Overview.Repositories[State.SelectedRepo].Jobs[State.SelectedJob].DefaultRunConfigYaml
	}
	fmt.Fprintln(LaunchRunWindow, runConfig)

	State.PreviousActiveWindow = v.Name()
	g.SetCurrentView(LAUNCH_RUN_VIEW)
	return nil
}

func ClosePopupView(g *c.Gui, v *c.View) error {
	if err := g.DeleteView(v.Name()); err != nil {
		return err
	}
	g.SetCurrentView(State.PreviousActiveWindow)
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

	if err := g.SetKeybinding(RUNS_VIEW, c.KeyArrowDown, c.ModNone, CursorDownAndUpdateRunInfo); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(RUNS_VIEW, c.KeyArrowUp, c.ModNone, CursorUpAndUpdateRunInfo); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(RUNS_VIEW, 'l', c.ModNone, OpenPopupLaunchWindow); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(RUNS_VIEW, 't', c.ModNone, TerminateRunWithConfirmationByRunId); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(RUNS_VIEW, 'T', c.ModNone, TerminateRunByRunId); err != nil {
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

func OpenConfirmationWindow(message string, options []string) (int, error) {

	return 0, nil
}

func OpenFeedbackWindow(message string) error {
	// lengthOfMessage :=  len(message)
	return nil
}

func TerminateRunWithConfirmationByRunId(g *c.Gui, v *c.View) error {
	return nil
}

func TerminateRunByRunId(g *c.Gui, v *c.View) error {
	SelectedRun := GetElementByCursor(v)
	run := Overview.FindRunIdBySubstring(State.SelectedRepo, State.SelectedJob, SelectedRun)
	Client.TerminateRun(run.RunId)

	LoadRunsForJob(g, JobsView)
	return nil
}

func setRunInformation(v *c.View) {
	RunInfoView.Clear()
	if v.Name() == RUNS_VIEW && len(v.ViewBufferLines()) > 0 {
		SelectedRun := GetElementByCursor(v)
		run := Overview.FindRunIdBySubstring(State.SelectedRepo, State.SelectedJob, SelectedRun)
		runInfo := make([]string, 0)

		s := time.Unix(int64(run.StartTime), 0)
		e := time.Unix(int64(run.EndTime), 0)
		duration := e.Sub(s)

		runInfo = append(runInfo, fmt.Sprintf("Start\t\t %s", s.Local().Format(("2006-01-02 15:04:05"))))
		runInfo = append(runInfo, fmt.Sprintf("End\t\t %s", e.Local().Format("2006-01-02 15:04:05")))
		runInfo = append(runInfo, fmt.Sprintf("Duration\t\t %s", duration.String()))
		runInfo = append(runInfo, fmt.Sprintf("Status\t\t %s", run.Status))

		FillViewWithItems(RunInfoView, runInfo)
	}
}

func CursorUpAndUpdateRunInfo(g *c.Gui, v *c.View) error {
	err := CursorUp(g, v)
	setRunInformation(v)
	return err
}

func CursorDownAndUpdateRunInfo(g *c.Gui, v *c.View) error {
	err := CursorDown(g, v)
	setRunInformation(v)
	return err
}

func FilterItemsInView(g *c.Gui, v *c.View) error {
	g.SetCurrentView(State.PreviousActiveWindow)
	switch State.PreviousActiveWindow {
	case REPOSITORIES_VIEW:
		filterTerm := FilterView.BufferLines()[0]
		cond_contains_term := func(repo s.RepositoryRepresentation) bool { return strings.Contains(repo.Name, filterTerm) }
		currentRepositoriesList := filter(Overview.GetRepositoryList(), cond_contains_term)
		RepoWindow.RenderItems(currentRepositoriesList)
	default:
		return nil

	}

	return nil
}

func SwitchToFilterView(g *c.Gui, v *c.View) error {
	State.PreviousActiveWindow = v.Name()
	g.SetCurrentView(FILTER_VIEW)
	FilterView.Title = fmt.Sprintf("Filter %s", v.Title)
	return nil
}

func LoadJobsForRepository(g *c.Gui, v *c.View) error {

	locationName := GetElementByCursor(v)
	State.SelectedRepo = locationName

	repo := Overview.GetRepoByLocation(locationName)

	Overview.AppendJobsToRepository(repo.Location, Client.GetJobsInRepository(repo))

	JobsView.Title = fmt.Sprintf("%s - Jobs", locationName)
	JobsView.Clear()
	CurrentJobsList = Overview.GetJobNamesInRepository(locationName)
	FillViewWithItems(JobsView, CurrentJobsList)

	ResetCursor(g, JOBS_VIEW)
	return SetFocus(g, JOBS_VIEW, v.Name())

}

func LoadRunsForJob(g *c.Gui, v *c.View) error {

	jobName := GetElementByCursor(v)
	State.SelectedJob = jobName

	repo := Overview.GetRepoByLocation(State.SelectedRepo)

	pipelineRuns := Client.GetPipelineRuns(repo, State.SelectedJob, 10)
	Overview.UpdatePipelineAndRuns(repo.Location, pipelineRuns)
	runs := Overview.GetRunsFor(State.SelectedRepo, State.SelectedJob)
	runInfos = make([]string, 0)
	// TODO make headers skippable in navigation
	// runInfos = append(runInfos, "Status \t RunId \t Time")
	for _, run := range runs {
		runInfos = append(runInfos, fmt.Sprintf("%s \t %s", run.Status, run.RunId))
	}

	RunsView.Title = fmt.Sprintf("%s - Runs", State.SelectedJob)
	RunsView.Clear()
	FillViewWithItems(RunsView, runInfos)

	ResetCursor(g, RUNS_VIEW)
	RunsView.SetCursor(0, 0)

	setRunInformation(RunsView)
	return SetFocus(g, RUNS_VIEW, v.Name())
}

func ValidateAndLaunchRun(g *c.Gui, v *c.View) error {

	Client.LaunchRunForJob(*Overview.Repositories[State.SelectedRepo], State.SelectedJob, LaunchRunWindow.BufferLines())
	ClosePopupView(g, LaunchRunWindow)
	LoadRunsForJob(g, JobsView)

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
		return Overview.GetRepositoryNames()
	case JOBS_VIEW:
		return CurrentJobsList
	case RUNS_VIEW:
		return runInfos
	}
	return []string{}

}