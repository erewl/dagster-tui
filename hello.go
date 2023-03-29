package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/jroimartin/gocui"
)

type RunRepresentation struct {
	runId         string
	startTime     float64
	endTime       float64
	status        string
	runconfigYaml string
}

type JobRepresentation struct {
	name                 string
	jobId                string
	description          string
	defaultRunConfigYaml string
	runs                 []*RunRepresentation
}

type RepositoryRepresentation struct {
	name     string
	location string
	jobs     map[string]*JobRepresentation
}

type Overview struct {
	url          string
	repositories map[string]*RepositoryRepresentation
}

func (o *Overview) GetRepositoryNames() []string {
	var names []string
	for k := range o.repositories {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func (o *Overview) GetJobNamesInRepository(repo string) []string {
	var names []string
	for k := range o.repositories[repo].jobs {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func (o *Overview) AppendRepositories(repos []Repository) {
	for _, node := range repos {
		rep := new(RepositoryRepresentation)
		rep.name = node.Name
		rep.location = node.Location.Name
		rep.jobs = make(map[string]*JobRepresentation, 0)

		o.repositories[rep.location] = rep
	}
}

func (o *Overview) AppendJobsToRepository(location string, jobs []Job) {

	for _, job := range jobs {
		jobRep := new(JobRepresentation)
		jobRep.name = job.Name
		jobRep.description = job.Description
		jobRep.defaultRunConfigYaml = ""
		jobRep.jobId = job.JobId
		jobRep.runs = make([]*RunRepresentation, 0)
		o.repositories[location].jobs[jobRep.name] = jobRep
	}

}

func (o *Overview) UpdatePipelineAndRuns(location string, pipeline PipelineOrError) {
	selectedJob := o.repositories[location].jobs[pipeline.Name]
	if len(pipeline.Presets) > 0 {
		selectedJob.defaultRunConfigYaml = pipeline.Presets[0].RunConfigYaml
	}
	selectedJob.runs = make([]*RunRepresentation, 0)
	for _, run := range(pipeline.Runs) {
		runRep := new(RunRepresentation)
		runRep.runId = run.RunId
		runRep.startTime = run.StartTime
		runRep.endTime = run.EndTime
		runRep.runconfigYaml = run.RunConfigYaml
		runRep.status = run.Status

		selectedJob.runs = append(selectedJob.runs, runRep)
	}
}

func (o *Overview) GetSortedRunNamesFor(location string, pipelineName string) []string {
	runNames := make([]string, 0)
	for _, run := range(o.repositories[location].jobs[pipelineName]).runs {
		runNames = append(runNames, run.runId)
	}
	return runNames
}

func (o *Overview) GetRepoByLocation(location string) RepositoryRepresentation {
	return *o.repositories[location]
}

var (
	RepositoriesView *gocui.View
	JobsView         *gocui.View
	RunsView         *gocui.View
	KeyMappingsView  *gocui.View

	data Overview

	currentRepositoriesList []string
	currentJobsList         []string

	runs     []Run
	runNames []string

	userHomeDir string
)

const (
	REPOSITORIES_VIEW = "repositories"
	JOBS_VIEW         = "jobs"
	RUNS_VIEW         = "runs"
	KEY_MAPPINGS_VIEW = "keymaps"
)

func FillViewWithItems(v *gocui.View, items []string) {
	v.Clear()

	for _, item := range items {
		fmt.Fprintln(v, item)
	}

}

type ConfigState struct {
	Repositories []string `json:"repositories"`
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

	data.repositories = make(map[string]*RepositoryRepresentation, 0)
	data.url = ""

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

	if len(data.repositories) == 0 {
		data.AppendRepositories(GetRepositories())
	}

	v, err := g.View(REPOSITORIES_VIEW)
	setWindowColors(g, REPOSITORIES_VIEW, "red")
	currentRepositoriesList = data.GetRepositoryNames()
	FillViewWithItems(v, currentRepositoriesList)
	v.SelFgColor = gocui.AttrBold
	v.SelBgColor = gocui.ColorRed
	// v.Autoscroll = true
	v.Wrap = true

	v, err = g.View(JOBS_VIEW)
	v.SelFgColor = gocui.AttrBold
	v.SelBgColor = gocui.ColorRed
	v.Autoscroll = true
	v.Wrap = true

	v, err = g.View(RUNS_VIEW)
	v.SelFgColor = gocui.AttrBold
	v.SelBgColor = gocui.ColorRed
	v.Autoscroll = true
	v.Wrap = true

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

func CloseKeyMaps(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView(v.Name()); err != nil {
		return err
	}
	g.SetCurrentView(REPOSITORIES_VIEW)
	return nil
}

func setKeybindings(g *gocui.Gui) error {
	// Set keybindings to switch focus between windows
	if err := g.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, switchFocusRight); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, switchFocusLeft); err != nil {
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
	if err := g.SetKeybinding(KEY_MAPPINGS_VIEW, 'x', gocui.ModNone, CloseKeyMaps); err != nil {
		return err
	}

	// define keybindings for moving between items
	if err := g.SetKeybinding(REPOSITORIES_VIEW, gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(REPOSITORIES_VIEW, gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(REPOSITORIES_VIEW, gocui.KeyEnter, gocui.ModNone, loadJobsForRepository); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding(JOBS_VIEW, gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(JOBS_VIEW, gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(JOBS_VIEW, gocui.KeyEnter, gocui.ModNone, loadRunsForJob); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding(RUNS_VIEW, gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding(RUNS_VIEW, gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		panic(err)
	}

	return nil
}

func setWindowColors(g *gocui.Gui, viewName string, bgColor string) error {
	// Get view
	view, err := g.View(viewName)
	if err != nil {
		return err
	}

	// Set background color
	if bgColor == "" {
		view.Highlight = false
		view.FgColor = gocui.Attribute(gocui.ColorDefault)
	} else {
		view.Highlight = true
		// view.FgColor = gocui.Attribute(gocui.ColorGreen) | gocui.AttrBold
	}

	// Set border color
	// if bgColor == "" {
	// 	view.FrameBgColor = gocui.ColorDefault
	// } else {
	// 	view.FrameBgColor = gocui.Attribute(gocui.ColorRed + 1)
	// }

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

func loadJobsForRepository(g *gocui.Gui, v *gocui.View) error {

	locationName := getElementByCursor(v)

	repo := data.GetRepoByLocation(locationName)

	data.AppendJobsToRepository(repo.location, GetJobsInRepository(repo))

	JobsView.Title = fmt.Sprintf("%s - Jobs", locationName)
	JobsView.Clear()
	currentJobsList = data.GetJobNamesInRepository(locationName)
	FillViewWithItems(JobsView, currentJobsList)

	resetCursor(g, JOBS_VIEW)
	return setFocus(g, JOBS_VIEW, v.Name())

}

func getElementByCursor(v *gocui.View) string {
	_, oy := v.Origin()
	_, vy := v.Cursor()

	items := getContentByView(v)

	return items[vy+oy]
}

func loadRunsForJob(g *gocui.Gui, v *gocui.View) error {

	
	r, _ := g.View(REPOSITORIES_VIEW)
	locationName := getElementByCursor(r)
	jobName := getElementByCursor(v)

	repo := data.GetRepoByLocation(locationName)

	pipelineRuns := GetPipelineRuns(repo, jobName, 10)
	data.UpdatePipelineAndRuns(repo.location, pipelineRuns)
	runNames = data.GetSortedRunNamesFor(locationName, jobName)
	d, _ := g.View(RUNS_VIEW)
	d.Title = fmt.Sprintf("%s - Runs", jobName)
	d.Clear()
	FillViewWithItems(d, runNames)

	resetCursor(g, RUNS_VIEW)
	return setFocus(g, RUNS_VIEW, v.Name())
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
