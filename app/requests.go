package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	s "nl/vdb/dagstertui/internal"
	"regexp"
	"strings"
)

type GraphQLClient struct {
	Url string
}

func (c *GraphQLClient) LoadRepositories() []s.Repository {

	query := "query RepositoriesQuery { repositoriesOrError { ... on RepositoryConnection { nodes { name location { name }}}}}"
	var reqStr = []byte(fmt.Sprintf(`{ 
		"query": "%s"
		}`, query))
	req, reqErr := http.NewRequest("POST", c.Url, bytes.NewBuffer(reqStr))
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

func (c *GraphQLClient) GetJobsInRepository(repository s.RepositoryRepresentation) []s.Job {
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
	req, reqErr := http.NewRequest("POST", c.Url, bytes.NewBuffer(reqStr))
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

func (c *GraphQLClient) GetPipelineRuns(repository s.RepositoryRepresentation, jobName string, limit int) s.PipelineOrError {
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
	req, reqErr := http.NewRequest("POST", c.Url, bytes.NewBuffer(reqStr))
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

func (c *GraphQLClient) LaunchRunForJob(repository s.RepositoryRepresentation, jobName string, runConfigYamlLines []string) string {
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
	}`, query, repository.Name, repository.Location, jobName, strings.Join(runConfigYamlLines, "\\n"))

	var reqStr = []byte(str)
	req, reqErr := http.NewRequest("POST", c.Url, bytes.NewBuffer(reqStr))
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

	var response s.LaunchRunResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		log.Fatalf("Failed to parse JSON: %v, %s", err, string(jsonData))
	}

	return response.Data.LaunchRun.Run.RunId
}

func (c *GraphQLClient) TerminateRun(runId string) s.TerminateRunResponse{
	re := regexp.MustCompile(`[\s]`)
	query := `mutation TerminateRun($runId: String!) {
				terminateRun(runId: $runId){
					__typename
					... on TerminateRunSuccess{
					run {
						runId
					}
					}
					... on TerminateRunFailure {
					message
					}
					... on RunNotFoundError {
					runId
					}
					... on PythonError {
					message
					stack
					}
				}
			}`

	query = re.ReplaceAllString(query, " ")
	formattedQuery := fmt.Sprintf(`{
		"query": "%s",
		"variables": { "runId": "%s"}
	}`, query, runId)

	req, reqErr := http.NewRequest("POST", c.Url, bytes.NewBuffer([]byte(formattedQuery)))
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

	var response s.TerminateRunResponse

	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		log.Fatalf("Failed to parse JSON: %v, %s", err, string(jsonData))
	}

	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}
	return response
}

func (c *GraphQLClient) GetLogs(runId string) {
	// example on how to fetch logs
	// 	query LogsForRun {
	//   logsForRun(
	//     runId: "4c322239-fac1-45e1-bcdb-a0e5a4c27a08"
	//   ) {
	//     ...on EventConnection {
	//       events {
	//         ...on LogsCapturedEvent {
	//           logKey
	//           fileKey
	//           message
	//         }
	//         # ...on LogMessageEvent{
	//         #   level
	//         #   timestamp
	//         #   message
	//         # }
	//       }
	//       cursor
	//       hasMore
	//     }
	//   }
	// }
}
