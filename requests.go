package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

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

type PipelineOrError struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Presets struct {
		// solidSelection ???
		RunConfigYaml string `json:"runConfigYaml"`
	} `json:"presets"`
	Runs []Run `json:"runs"`
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

const (
	URL = "https://dagster.test-backend.vdbinfra.nl/graphql"
)

func GetRepositories() []Repository {

	query := "query RepositoriesQuery { repositoriesOrError { ... on RepositoryConnection { nodes { name location { name }}}}}"
	var reqStr = []byte(fmt.Sprintf(`{ 
		"query": "%s"
		}`, query))
	req, reqErr := http.NewRequest("POST", URL, bytes.NewBuffer(reqStr))
	if reqErr != nil {
		panic(reqErr)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	resp, respErr := client.Do(req)
	if respErr != nil {
		log.Fatalf("Failed POST request: %v", respErr)
	}
	defer resp.Body.Close()

	jsonData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	var response RepositoriesResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		log.Fatalf("Failed to parse JSON: %v, %s", err, string(jsonData))
	}

	repos := response.Data.RepositoriesOrError.Nodes
	return repos
}

func GetJobsInRepository(repository RepositoryRepresentation) []Job {
	re := regexp.MustCompile(`[\s]`)
	query := `query JobsQuery($repositoryLocationName: String!, $repositoryName: String!) {
	repositoryOrError(
		repositorySelector: {
		repositoryLocationName: $repositoryLocationName 
		repositoryName: $repositoryName 
		}
	) {
		... on Repository {
		jobs {
			name 
			id 
			description 
		}
		}
	}}`
	query = re.ReplaceAllString(query, " ")
	str := fmt.Sprintf(`{
		"query": "%s",
		"variables": { "repositoryName": "%s", "repositoryLocationName": "%s" }
	}`, query, repository.name, repository.location)

	var reqStr = []byte(str)
	req, reqErr := http.NewRequest("POST", URL, bytes.NewBuffer(reqStr))
	if reqErr != nil {
		panic(reqErr)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	resp, respErr := client.Do(req)
	if respErr != nil {
		log.Fatalf("Failed POST request: %v", respErr)
	}
	defer resp.Body.Close()

	jsonData, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	var response JobsResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		log.Fatalln(string(jsonData))
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	jobs := response.Data.RepositoriesOrError.Jobs
	return jobs
}

func GetPipelineRuns(repository RepositoryRepresentation, jobName string, limit int) PipelineOrError {
	query := fmt.Sprintf(`query RunIdsQuery {
	pipelineOrError(
		params: {
		repositoryName: "%s"
		pipelineName:"%s"
		repositoryLocationName:"%s"
		}
	) {
		...on Pipeline {
		id
		name
		presets {
				solidSelection
				runConfigYaml
		}
		runs(
			limit: %d
		) {
			runId
			status
			startTime
			endTime
			runConfigYaml
		}
		}
		
		...on PipelineNotFoundError {
		message
		}
	}}`, repository.name, jobName, repository.location, limit)
	query = regexp.MustCompile(`[\s]`).ReplaceAllString(query, " ")
	query = regexp.MustCompile(`"`).ReplaceAllString(query, `\"`)

	var reqStr = []byte(fmt.Sprintf(`{ 
		"query": "%s"
		}`, query))
	req, reqErr := http.NewRequest("POST", URL, bytes.NewBuffer(reqStr))
	if reqErr != nil {
		panic(reqErr)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	resp, respErr := client.Do(req)
	if respErr != nil {
		log.Fatalf("Failed POST request: %v", respErr)
	}
	defer resp.Body.Close()

	jsonData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	var response RunsResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		log.Fatalf("Failed to parse JSON: %v, %s", err, string(jsonData))
	}

	pipelineOrError := response.Data.PipelineOrError

	return pipelineOrError
}

func LaunchRunForJob(repository Repository, jobName string, runConfigYaml string) error {
	query := `mutation LaunchRunMutation(
		$repositoryLocationName: String!
		$repositoryName: String!
		$jobName: String!
		$runConfigData: RunConfigData!
	) {
		launchRun(
			executionParams: {
			selector: {
				repositoryLocationName: $repositoryLocationName
				repositoryName: $repositoryName
				jobName: $jobName
			}
			runConfigData: $runConfigData
			}
		) {
			__typename
			... on LaunchRunSuccess {
			run {
				runId
			}
			}
			... on RunConfigValidationInvalid {
			errors {
				message
				reason
			}
			}
			... on PythonError {
			message
			}
		}
	}`
	query = regexp.MustCompile(`[\s]`).ReplaceAllString(query, " ")
	query = regexp.MustCompile(`"`).ReplaceAllString(query, `\"`)

	str := fmt.Sprintf(`{
		"query": "%s",
		"variables": { "repositoryName": "%s", "repositoryLocationName": "%s" , "jobName": "%s", "runConfigData": "%s"}
	}`, query, repository.Name, repository.Location.Name, jobName, runConfigYaml)

	var reqStr = []byte(str)
	req, reqErr := http.NewRequest("POST", URL, bytes.NewBuffer(reqStr))
	if reqErr != nil {
		panic(reqErr)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	resp, respErr := client.Do(req)
	if respErr != nil {
		log.Fatalf("Failed POST request: %v", respErr)
	}
	defer resp.Body.Close()

	jsonData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	var response string
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		log.Fatalf("Failed to parse JSON: %v, %s", err, string(jsonData))
	}

	fmt.Println(response)

	return nil
}
