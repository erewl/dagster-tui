package main

import (
	"flag"
	"fmt"
	c "github.com/jroimartin/gocui"
	. "nl/vdb/dagstertui/app"
	s "nl/vdb/dagstertui/internal"
	"os"
	"strings"
)

var (
	userHomeDir string
)

func main() {

	home, err := os.UserHomeDir()
	userHomeDir = home
	LoadConfig(home)

	environmentFlag := flag.String("e", "default", "sets the home url of the dagster environment")

	// Parse the command-line arguments to set the value of environmentFlag
	flag.Parse()

	Overview = &s.Overview{
		Repositories: make(map[string]*s.RepositoryRepresentation, 0),
		Url:          Conf.Environments[*environmentFlag],
	}
	if *environmentFlag == "default" {
		Overview.Url = Conf.Environments[Conf.Environments[*environmentFlag]]
	}

	Client = &GraphQLClient{
		Url: fmt.Sprintf("%s/graphql", Overview.Url),
	}

	State = &ApplicationState{
		PreviousActiveWindow: "",
		SelectedRepo:         "",
		SelectedJob:          "",
		SelectedRun:          "",
		RepoFilter:           "",
	}

	// Initialize gocui
	g, err := c.NewGui(c.Output256)
	if err != nil {
		panic(err)
	}

	defer g.Close()

	// Set layout function
	g.SetManagerFunc(Layout)

	// Set keybindings
	err = SetKeybindings(g)
	if err != nil {
		panic(err)
	}

	g.SelFgColor = c.ColorGreen | c.AttrBold
	g.BgColor = c.ColorDefault
	g.Highlight = true

	// called once
	InitializeViews(g)

	SetWindowColors(g, REPOSITORIES_VIEW, "red")

	repos := Client.LoadRepositories()
	Overview.AppendRepositories(repos)
	RepoWindow.RenderItems(Overview.GetRepositoryList())

	environmentInfo := []string{strings.TrimPrefix(Overview.Url, "https://")}
	FillViewWithItems(EnvironmentInfoView, environmentInfo)

	SetViewStyles(RepoWindow.Base.View)
	SetViewStyles(JobsWindow.Base.View)
	SetViewStyles(RunsWindow.Base.View)
	FilterView.Editable = true
	FilterView.Editor = DefaultEditor

	// Start main loop
	err = g.MainLoop()
	if err != nil && err != c.ErrQuit {
		panic(err)
	}
}
