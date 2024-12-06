package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github-met/infrastructure/utils"
	"github-met/types"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var token string
var githubGraphQLURL string

func init() {
	token = os.Getenv("GITHUB_TOKEN")
	githubGraphQLURL = os.Getenv("GITHUB_GRAPHQL_URL")
	if token == "" {
		log.Fatal("GITHUB_TOKEN is not set")
		os.Exit(1)
	}
	if githubGraphQLURL == "" {
		githubGraphQLURL = "https://api.github.com/graphql"
	}
}

func GetUserCreatedAt(username *string) time.Time {
	query := `
	query {
	  user(login: "` + *username + `") {
	    createdAt
	  }
	}`

	graphqlQuery := types.GraphQLQuery{
		Query: query,
	}

	payload, err := json.Marshal(graphqlQuery)
	if err != nil {
		fmt.Println("Error marshaling query:", err)
		os.Exit(1)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", githubGraphQLURL, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		os.Exit(1)
	}

	var data types.User
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error unmarshaling response:", err)
		os.Exit(1)
	}

	createdAt, err := time.Parse(time.RFC3339, data.Data.User.CreatedAt)
	if err != nil {
		fmt.Println("Error parsing created at:", err)
		os.Exit(1)
	}
	return createdAt
}

// func GetContributionsStreak(username string, startDate *time.Time) types.StreakData {
// 	query := `
// 	query {
// 	  user(login: "` + username + `") {
// 	    contributionsCollection {
// 	      contributionCalendar {
// 	        weeks {
// 	          contributionDays {
// 	            date
// 	            contributionCount
// 	          }
// 	        }
// 	      }
// 	    }
// 	  }
// 	}`

// 	graphqlQuery := types.GraphQLQuery{
// 		Query: query,
// 	}

// 	payload, err := json.Marshal(graphqlQuery)
// 	if err != nil {
// 		fmt.Println("Error marshaling query:", err)
// 		os.Exit(1)
// 	}

// 	client := &http.Client{}
// 	req, err := http.NewRequest("POST", githubGraphQLURL, bytes.NewBuffer(payload))
// 	if err != nil {
// 		fmt.Println("Error creating request:", err)
// 		os.Exit(1)
// 	}

// 	req.Header.Set("Authorization", "Bearer "+token)
// 	req.Header.Set("Content-Type", "application/json")

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		fmt.Println("Error making request:", err)
// 		os.Exit(1)
// 	}
// 	defer resp.Body.Close()

// 	// Read the response
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		fmt.Println("Error reading response body:", err)
// 		os.Exit(1)
// 	}

// 	// Unmarshal the response into ContributionData
// 	var data types.ContributionData
// 	err = json.Unmarshal(body, &data)
// 	if err != nil {
// 		fmt.Println("Error unmarshaling response:", err)
// 		os.Exit(1)
// 	}
// 	streak, startedDateStreak, endDateStreak := utils.CalculateStreak(data.Data.User.ContributionsCollection.ContributionCalendar.Weeks)
// 	startedDate, err := time.Parse(time.RFC3339, data.Data.User.CreatedAt)
// 	if err != nil {
// 		fmt.Println("Error parsing started date:", err)
// 		os.Exit(1)
// 	}

// 	return types.StreakData{
// 		Streak:          streak,
// 		StreakStartDate: startedDateStreak,
// 		StreakEndDate:   endDateStreak,
// 		StartedDate:     startedDate,
// 	}
// }

func GetContributionsForYear(username string, start *time.Time, end *time.Time) (types.ContributionData, error) {
	contributionsCollectionParams := ""

	if start != nil && end != nil {
		contributionsCollectionParams = `(from: "` + start.Format(time.RFC3339) + `", to: "` + end.Format(time.RFC3339) + `")`
	} else if start != nil {
		contributionsCollectionParams = `(from: "` + start.Format(time.RFC3339) + `")`
	}

	query := `
	query {
		user(login: "` + username + `") {
			contributionsCollection` + contributionsCollectionParams + ` {
				contributionCalendar {
					weeks {
						contributionDays {
							date
							contributionCount
						}
					}
					totalContributions
				}
			}
		}
	}`

	fmt.Println("query:", query)

	graphqlQuery := types.GraphQLQuery{
		Query: query,
	}

	payload, err := json.Marshal(graphqlQuery)
	if err != nil {
		fmt.Println("Error marshaling query:", err)
		return types.ContributionData{}, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", githubGraphQLURL, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return types.ContributionData{}, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return types.ContributionData{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return types.ContributionData{}, err
	}

	var data types.ContributionData
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error unmarshaling response:", err)
		return types.ContributionData{}, err
	}

	prettyJSON, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error generating pretty JSON:", err)
	} else {
		fmt.Printf("data: %v\n", string(prettyJSON))
	}

	return data, nil
}

func GetAllContributions(username string, start *time.Time) (int, types.CalculatedStreakData) {
	yearRanges := utils.RangeOfYears(start)

	contributionsDataChan := make(chan types.ContributionData, len(yearRanges))
	errorsChan := make(chan error, len(yearRanges))

	for _, yearRange := range yearRanges {
		go func(start, end time.Time) {
			contributionData, err := GetContributionsForYear(username, &start, &end)
			if err != nil {
				errorsChan <- err
				return
			}
			contributionsDataChan <- contributionData
		}(yearRange[0], yearRange[1])
	}

	totalContributions := 0
	weeks := []types.ContributionWeek{}
	for i := 0; i < len(yearRanges); i++ {
		select {
		case contributionData := <-contributionsDataChan:
			totalContributions += contributionData.Data.User.ContributionsCollection.ContributionCalendar.TotalContributions
			weeks = append(weeks, contributionData.Data.User.ContributionsCollection.ContributionCalendar.Weeks...)
		case err := <-errorsChan:
			fmt.Println("Error fetching contributions:", err)
			return 0, types.CalculatedStreakData{}
		}
	}

	calculatedStreakData := utils.CalculateStreak(weeks)

	return totalContributions, calculatedStreakData
}
