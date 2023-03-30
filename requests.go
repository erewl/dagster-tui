package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	s"nl/vdb/dagstertui/datastructures"
)

const (
	URL = "https://dagster.test-backend.vdbinfra.nl/graphql"
)

func GetRepositories() []s.Repository {

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

	var response s.RepositoriesResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		log.Fatalf("Failed to parse JSON: %v, %s", err, string(jsonData))
	}

	repos := response.Data.RepositoriesOrError.Nodes
	return repos
}

func GetJobsInRepository(repository s.RepositoryRepresentation) []s.Job {
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
	}`, query, repository.Name, repository.Location)

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

	var response s.JobsResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		log.Fatalln(string(jsonData))
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	jobs := response.Data.RepositoriesOrError.Jobs
	return jobs
}

func GetPipelineRuns(repository s.RepositoryRepresentation, jobName string, limit int) s.PipelineOrError {
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
	}}`, repository.Name, jobName, repository.Location, limit)
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

	var response s.RunsResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		log.Fatalf("Failed to parse JSON: %v, %s", err, string(jsonData))
	}

	pipelineOrError := response.Data.PipelineOrError

	return pipelineOrError
}

func LaunchRunForJob(repository s.Repository, jobName string, runConfigYaml string) error {
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
