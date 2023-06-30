package app

import (
	"encoding/json"
	"fmt"
	c "github.com/jroimartin/gocui"
	"io/ioutil"
	s "nl/vdb/dagstertui/internal"
	// l "nl/vdb/dagstertui/log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var (
	ConfirmationView    *s.ListView[string]
	LaunchRunWindow     *s.InfoView
	FeedbackView        *s.InfoView
	KeyMappingsView     *s.InfoView
	EnvironmentInfoView *s.InfoView
	RunInfoWindow       *s.InfoView
	FilterView          *s.InfoView

	RunsWindow *s.ListView[s.RunRepresentation]
	RepoWindow *s.ListView[s.RepositoryRepresentation]
	JobsWindow *s.ListView[s.JobRepresentation]

	Overview *s.Overview
	State    *ApplicationState
	Conf     Config

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

func (a *ApplicationState) SetNewActiveWindow(g *c.Gui, previousWindow string, currentWindow string) error {
	a.PreviousActiveWindow = previousWindow
	return SetFocus(g, currentWindow, previousWindow)
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

func InitializeViews(g *c.Gui) error {

	RepoWindow = &s.ListView[s.RepositoryRepresentation]{}
	JobsWindow = &s.ListView[s.JobRepresentation]{}
	RunsWindow = &s.ListView[s.RunRepresentation]{}

	RunInfoWindow = &s.InfoView{}
	FeedbackView = &s.InfoView{}
	EnvironmentInfoView = &s.InfoView{}
	FilterView = &s.InfoView{}
	ConfirmationView = &s.ListView[string]{}
	LaunchRunWindow = &s.InfoView{}
	KeyMappingsView = &s.InfoView{}

	RepoWindow.Initialize(g, "Repositories", REPOSITORIES_VIEW,
		func(a s.RepositoryRepresentation) string { return a.Location },
		func(a s.RepositoryRepresentation) string { return a.Location })
	JobsWindow.Initialize(g, "Jobs", JOBS_VIEW,
		func(a s.JobRepresentation) string { return a.Name },
		func(a s.JobRepresentation) string { return a.Name })
	RunsWindow.Initialize(g, "Runs", RUNS_VIEW,
		func(a s.RunRepresentation) string { return fmt.Sprintf("%s \t %s", a.Status, a.RunId) },
		func(a s.RunRepresentation) string { return fmt.Sprint(a.StartTime) })

	RunInfoWindow.Initialize(g, "Run Info", RUN_INFO_VIEW)
	EnvironmentInfoView.Initialize(g, "Dagster Info", ENVIRONMENT_INFO)
	FilterView.Initialize(g, "Filter", FILTER_VIEW)

	FilterView.Base.View.Editable = true
	FilterView.Base.View.Editor = FilterEditor

	RepoWindow.Base.SetNavigableFeedback(g)
	JobsWindow.Base.SetNavigableFeedback(g)
	RunsWindow.Base.SetNavigableFeedback(g)

	// Set focus on main window
	if _, err := g.SetCurrentView(REPOSITORIES_VIEW); err != nil {
		panic(err)
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
			runId = Overview.FindRunIdBySubstring(State.SelectedRepo, State.SelectedJob, RunsWindow.GetElementOnCursorPosition()).RunId
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

	KeyMappingsView.Initialize(g, "Key Map", KEY_MAPPINGS_VIEW)
	KeyMappingsView.Base.RenderView(g, int(float64(maxX)*0.2), 1, int(float64(maxX)*0.8), maxY+1)
	KeyMappingsView.RenderContent([]string{s.KeyMap})
	return State.SetNewActiveWindow(g, v.Name(), KEY_MAPPINGS_VIEW)
}

var DefaultEditor c.Editor = c.EditorFunc(simpleEditor)
var FilterEditor c.Editor = c.EditorFunc(filterEditor)

func filterEditor(v *c.View, key c.Key, ch rune, mod c.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == c.KeyBackspace || key == c.KeyBackspace2:
		v.EditDelete(true)
	}
	FilterItemsInView(v)
}

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

	LaunchRunWindow.Initialize(g, "Launch Run For", LAUNCH_RUN_VIEW)
	LaunchRunWindow.Base.RenderView(g, int(float64(maxX)*0.2), int(float64(maxY)*0.2), int(float64(maxX)*0.8), int(float64(maxY)*0.8))

	LaunchRunWindow.Base.View.Editable = true
	LaunchRunWindow.Base.View.Editor = DefaultEditor
	LaunchRunWindow.Base.View.Highlight = true
	LaunchRunWindow.Base.View.SelBgColor = c.ColorBlue
	LaunchRunWindow.Base.View.SetCursor(0, 0)

	runConfig := ""
	if v.Name() == RUNS_VIEW {
		SelectedRun := RunsWindow.GetElementOnCursorPosition()
		runConfig = Overview.FindRunIdBySubstring(State.SelectedRepo, State.SelectedJob, SelectedRun).RunconfigYaml
	} else {
		runConfig = Overview.Repositories[State.SelectedRepo].Jobs[State.SelectedJob].DefaultRunConfigYaml
	}

	LaunchRunWindow.RenderContent([]string{runConfig})

	return State.SetNewActiveWindow(g, v.Name(), LAUNCH_RUN_VIEW)
}

func ClosePopupView(g *c.Gui, v *c.View) error {
	err := State.SetNewActiveWindow(g, v.Name(), State.PreviousActiveWindow)
	if err = g.DeleteView(v.Name()); err != nil {
		return err
	}
	return err
}

func OpenConfirmationWindow(g *c.Gui, message string, options []string) error {
	ConfirmationView.Initialize(g, message, CONFIRMATION_VIEW, s.Identity[string], s.Identity[string])
	ConfirmationView.Base.RenderView(g, 10, 10, 40, 40)
	ConfirmationView.Base.SetNavigableFeedback(g)

	ConfirmationView.RenderItems(options, false)

	return nil
}

func OpenFeedbackWindow(g *c.Gui, v *c.View, message string) error {
	lengthOfMessage := len(message)
	FeedbackView.Initialize(g, "Termination Response", FEEDBACK_VIEW)
	FeedbackView.Base.RenderView(g, 10, 10, 10+lengthOfMessage, 20)
	FeedbackView.RenderContent([]string{message})
	return State.SetNewActiveWindow(g, RUNS_VIEW, FEEDBACK_VIEW)
}

func ShowTerminationOptions(g *c.Gui, v *c.View) error {
	OpenConfirmationWindow(g, "Terminate run?", []string{"Yes", "No"})
	return State.SetNewActiveWindow(g, v.Name(), CONFIRMATION_VIEW)
}

func TerminateRunWithConfirmationByRunId(g *c.Gui, v *c.View) error {
	d := ConfirmationView.GetElementOnCursorPosition()
	if d == "Yes" {
		ClosePopupView(g, ConfirmationView.Base.View)
		return TerminateRunByRunId(g, v)
	}
	return nil
}

func TerminateRunByRunId(g *c.Gui, v *c.View) error {
	SelectedRun := RunsWindow.GetElementOnCursorPosition()
	run := Overview.FindRunIdBySubstring(State.SelectedRepo, State.SelectedJob, SelectedRun)
	resp := Client.TerminateRun(run.RunId)

	respStr := fmt.Sprintf("Termination Request of type: %s \n\n %s", resp.Data.TerminateRun.TypeName, resp.Data.TerminateRun.Message)
	LoadRuns(g, JobsWindow.Base.View)
	OpenFeedbackWindow(g, v, respStr)
	return nil
}

func setRunInformation(v *c.View) {
	if v.Name() == RUNS_VIEW && len(v.ViewBufferLines()) > 0 {
		SelectedRun := RunsWindow.GetElementOnCursorPosition()
		run := Overview.FindRunIdBySubstring(State.SelectedRepo, State.SelectedJob, SelectedRun)
		runInfo := make([]string, 0)

		s := time.Unix(int64(run.StartTime), 0)
		e := time.Unix(int64(run.EndTime), 0)
		duration := e.Sub(s)

		runInfo = append(runInfo, fmt.Sprintf("Start\t\t %s", s.Local().Format(("2006-01-02 15:04:05"))))
		runInfo = append(runInfo, fmt.Sprintf("End\t\t %s", e.Local().Format("2006-01-02 15:04:05")))
		runInfo = append(runInfo, fmt.Sprintf("Duration\t\t %s", duration.String()))
		runInfo = append(runInfo, fmt.Sprintf("Status\t\t %s", run.Status))

		RunInfoWindow.RenderContent(runInfo)
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

func FilterItemsInView(v *c.View) error {
	switch State.PreviousActiveWindow {
	case REPOSITORIES_VIEW:
		filterTerm := FilterView.Base.View.BufferLines()[0]
		cond_contains_term := func(repo s.RepositoryRepresentation) bool { return strings.Contains(repo.Location, filterTerm) }
		currentRepositoriesList := s.Filter(Overview.GetRepositoryList(), cond_contains_term)
		RepoWindow.RenderItems(currentRepositoriesList)
	default:
		return nil

	}
	return nil
}

func SwitchToFilterView(g *c.Gui, v *c.View) error {
	FilterView.Base.Title = fmt.Sprintf("Filter %s", v.Title)
	return State.SetNewActiveWindow(g, v.Name(), FILTER_VIEW)
}


func LoadJobsForRepository(g *c.Gui, v *c.View) error {

	locationName := RepoWindow.GetElementOnCursorPosition()
	State.SelectedRepo = locationName

	repo := Overview.GetRepoByLocation(locationName)

	Overview.AppendJobsToRepository(repo.Location, Client.GetJobsInRepository(repo))

	JobsWindow.Base.Title = fmt.Sprintf("%s - Jobs", locationName)
	JobsWindow.RenderItems(Overview.GetJobNamesInRepository(locationName))

	JobsWindow.ResetCursor()
	return SetFocus(g, JOBS_VIEW, v.Name())

}


func LoadRuns(g *c.Gui, v* c.View) {
	jobName := JobsWindow.GetElementOnCursorPosition()
	State.SelectedJob = jobName

	repo := Overview.GetRepoByLocation(State.SelectedRepo)

	pipelineRuns := Client.GetPipelineRuns(repo, State.SelectedJob, 10)
	Overview.UpdatePipelineAndRuns(repo.Location, pipelineRuns)
	runs := Overview.GetRunsFor(State.SelectedRepo, State.SelectedJob)
	// TODO make headers skippable in navigation
	// runInfos = append(runInfos, "Status \t RunId \t Time")

	RunsWindow.Base.Title = fmt.Sprintf("%s - Runs", State.SelectedJob)
	RunsWindow.RenderItems(runs)
	RunsWindow.ResetCursor()

	setRunInformation(RunsWindow.Base.View)
}


func LoadRunsForJob(g *c.Gui, v *c.View) error {
	LoadRuns(g, v)
	return SetFocus(g, RUNS_VIEW, v.Name())
}

func ValidateAndLaunchRun(g *c.Gui, v *c.View) error {

	Client.LaunchRunForJob(*Overview.Repositories[State.SelectedRepo], State.SelectedJob, LaunchRunWindow.Base.View.BufferLines())
	ClosePopupView(g, LaunchRunWindow.Base.View)
	LoadRuns(g, JobsWindow.Base.View)

	return nil
}
