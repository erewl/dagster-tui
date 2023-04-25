package datastructures

type Repository struct {
	Name     string `json:"name"`
	Location struct {
		Name string `json:"name"`
	} `json:"location"`
}

type RepositoriesResponse struct {
	Data struct {
		RepositoriesOrError struct {
			Nodes []Repository `json:"nodes"`
		} `json:"repositoriesOrError"`
	} `json:"data"`
}

type LaunchRunResponse struct {
	Data struct {
		LaunchRun struct {
			Run struct {
				RunId string `json:"runId"`
			} `json:"run"`
		} `json:"launchRun"`
	} `json:"data"`
}

type TerminateRunResponse struct {
	Data struct {
		TerminateRun struct {
			Run struct {
				RunId string `json:"runId"`
			} `json:"run"`
		} `json:"terminateRun"`
	} `json:"data"`
}

type Jobs struct {
	name string
}

type Job struct {
	Name        string `json:"name"`
	JobId       string `json:"id"`
	Description string `json:"description"`
}

type JobsResponse struct {
	Data struct {
		RepositoriesOrError struct {
			Jobs []Job `json:"jobs"`
		} `json:"repositoryOrError"`
	} `json:"data"`
}

type Preset struct {
	RunConfigYaml string `json:"runConfigYaml"`
}

type PipelineOrError struct {
	Id      string   `json:"id"`
	Name    string   `json:"name"`
	Presets []Preset `json:"presets"`
	Runs    []Run    `json:"runs"`
}

type RunsResponse struct {
	Data struct {
		PipelineOrError PipelineOrError `json:"pipelineOrError"`
	} `json:"data"`
}

type Run struct {
	RunId         string  `json:"runId"`
	StartTime     float64 `json:"startTime"`
	EndTime       float64 `json:"endTime"`
	Status        string  `json:"status"`
	RunConfigYaml string  `json:"runConfigYaml"`
}
