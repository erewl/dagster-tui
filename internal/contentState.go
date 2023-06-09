package internal

import (
	"sort"
	"strings"
)

type RunRepresentation struct {
	RunId         string
	StartTime     float64
	EndTime       float64
	Status        string
	RunconfigYaml string
}

type JobRepresentation struct {
	Name                 string
	JobId                string
	Description          string
	DefaultRunConfigYaml string
	Runs                 []*RunRepresentation
}

type RepositoryRepresentation struct {
	Name     string
	Location string
	Jobs     map[string]*JobRepresentation
}

type Overview struct {
	Url          string
	Repositories map[string]*RepositoryRepresentation
}

func (o *Overview) GetRepositoryList() []RepositoryRepresentation {
	var repoReps []RepositoryRepresentation
	for _, v := range o.Repositories {
		repoReps = append(repoReps, *v)
	}
	return repoReps
}

func (o *Overview) GetRepositoryNames() []string {
	var names []string
	for k := range o.Repositories {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func (o *Overview) GetJobNamesInRepository(repo string) []JobRepresentation {
	var names []JobRepresentation
	for _, v := range o.Repositories[repo].Jobs {
		names = append(names, *v)
	}
	return names
}

func (o *Overview) AppendRepositories(repos []Repository) {
	for _, node := range repos {
		rep := new(RepositoryRepresentation)
		rep.Name = node.Name
		rep.Location = node.Location.Name
		rep.Jobs = make(map[string]*JobRepresentation, 0)

		o.Repositories[rep.Location] = rep
	}
}

func (o *Overview) AppendJobsToRepository(location string, Jobs []Job) {

	for _, job := range Jobs {
		jobRep := new(JobRepresentation)
		jobRep.Name = job.Name
		jobRep.Description = job.Description
		jobRep.DefaultRunConfigYaml = ""
		jobRep.JobId = job.JobId
		jobRep.Runs = make([]*RunRepresentation, 0)
		o.Repositories[location].Jobs[jobRep.Name] = jobRep
	}

}

func (o *Overview) UpdatePipelineAndRuns(location string, pipeline PipelineOrError) {
	SelectedJob := o.Repositories[location].Jobs[pipeline.Name]
	if len(pipeline.Presets) > 0 {
		SelectedJob.DefaultRunConfigYaml = pipeline.Presets[0].RunConfigYaml
	}
	SelectedJob.Runs = make([]*RunRepresentation, 0)
	for _, run := range pipeline.Runs {
		runRep := new(RunRepresentation)
		runRep.RunId = run.RunId
		runRep.StartTime = run.StartTime
		runRep.EndTime = run.EndTime
		runRep.RunconfigYaml = run.RunConfigYaml
		runRep.Status = run.Status

		SelectedJob.Runs = append(SelectedJob.Runs, runRep)
	}
}

func (o *Overview) GetSortedRunNamesFor(location string, pipelineName string) []string {
	runNames := make([]string, 0)
	for _, run := range (o.Repositories[location].Jobs[pipelineName]).Runs {
		runNames = append(runNames, run.RunId)
	}
	return runNames
}

func (o *Overview) FindRunIdBySubstring(location string, pipelineName string, info string) RunRepresentation {
	for _, run := range o.Repositories[location].Jobs[pipelineName].Runs {
		if strings.Contains(info, run.RunId) {
			return *run
		}
	}
	panic("Couldnt find any run")
}

func (o *Overview) GetRunsFor(location string, pipelineName string) []RunRepresentation {
	runs := make([]RunRepresentation, 0)
	for _, run := range o.Repositories[location].Jobs[pipelineName].Runs {
		runs = append(runs, *run)
	}
	return runs
}

func (o *Overview) GetRepoByLocation(location string) RepositoryRepresentation {
	return *o.Repositories[location]
}
