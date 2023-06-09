package main

import (
	"flag"
	"fmt"
	. "nl/vdb/dagstertui/app"
	s "nl/vdb/dagstertui/internal"
	"os"
	"strings"

	c "github.com/jroimartin/gocui"
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
	g.InputEsc = true

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

	EnvironmentInfoView.RenderContent([]string{strings.TrimPrefix(Overview.Url, "https://")})

	// Start main loop
	err = g.MainLoop()
	if err != nil && err != c.ErrQuit {
		panic(err)
	}
}
