package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type GraphQLQuery struct {
	Query string `json:"query"`
}

// ContributionDay represents a day's contributions
type ContributionDay struct {
	Date              string `json:"date"`
	ContributionCount int    `json:"contributionCount"`
}

// ContributionWeek represents a week's contributions
type ContributionWeek struct {
	ContributionDays []ContributionDay `json:"contributionDays"`
}

// ContributionData is the structure for the response
type ContributionData struct {
	Data struct {
		User struct {
			ContributionsCollection struct {
				ContributionCalendar struct {
					Weeks []ContributionWeek `json:"weeks"`
				} `json:"contributionCalendar"`
			} `json:"contributionsCollection"`
			CreatedAt string `json:"createdAt"`
			TotalContributions int `json:"totalContributions"`
		} `json:"user"`
	} `json:"data"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		http.Error(w, "GITHUB_TOKEN is not set", http.StatusInternalServerError)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	query := `
	query {
	  user(login: "` + username + `") {
	    contributionsCollection {
	      contributionCalendar {
	        weeks {
	          contributionDays {
	            date
	            contributionCount
	          }
	        }
	      }
				totalContributions
	    }
	    createdAt
	  }
	}`

	graphqlQuery := GraphQLQuery{
		Query: query,
	}

	// Marshal the query into JSON
	payload, err := json.Marshal(graphqlQuery)
	if err != nil {
		fmt.Println("Error marshaling query:", err)
		os.Exit(1)
	}

	// Make the HTTP request
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(payload))
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

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		os.Exit(1)
	}

	// Unmarshal the response into ContributionData
	var data ContributionData
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error unmarshaling response:", err)
		os.Exit(1)
	}

	totalContributions := data.Data.User.TotalContributions

	streak := calculateStreak(data.Data.User.ContributionsCollection.ContributionCalendar.Weeks)
	startedDate, err := time.Parse(time.RFC3339, data.Data.User.CreatedAt)
	if err != nil {
		fmt.Println("Error parsing started date:", err)
		os.Exit(1)
	}

	// Create a response map to include streak and started date information
	// response := map[string]interface{}{
	// 	"streak":      streak,
	// 	"startedDate": startedDate,
	// }

	// jsonData, err := json.MarshalIndent(response, "", "  ")
	// if err != nil {
	// 	fmt.Println("Error marshaling data to JSON:", err)
	// 	os.Exit(1)
	// }

	w.Header().Set("Content-Type", "image/svg+xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(streakSVG(streak, startedDate, totalContributions)))
}

func calculateStreak(weeks []ContributionWeek) int {
	var streak int
	var previousDate time.Time

	// Iterate backward through weeks and days to calculate streak
	for i := len(weeks) - 1; i >= 0; i-- {
		week := weeks[i]
		for j := len(week.ContributionDays) - 1; j >= 0; j-- {
			day := week.ContributionDays[j]
			if day.ContributionCount == 0 {
				return streak // Streak ends when a day with no contributions is encountered
			}

			date, err := time.Parse("2006-01-02", day.Date)
			if err != nil {
				fmt.Println("Error parsing date:", err)
				os.Exit(1)
			}

			// Check for consecutive days
			if !previousDate.IsZero() && !date.AddDate(0, 0, 1).Equal(previousDate) {
				return streak
			}

			streak++
			previousDate = date
		}
	}

	return streak
}

func streakSVG(streak int, startedDate time.Time, totalContributions int) string {
	return `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" style="isolation: isolate" viewBox="0 0 495 195" width="495px" height="195px" direction="ltr">
        <style>
            @keyframes currstreak {
                0% { font-size: 3px; opacity: 0.2; }
                80% { font-size: 34px; opacity: 1; }
                100% { font-size: 28px; opacity: 1; }
            }
            @keyframes fadein {
                0% { opacity: 0; }
                100% { opacity: 1; }
            }
        </style>
        <defs>
            <clipPath id="outer_rectangle">
                <rect width="495" height="195" rx="4.5"/>
            </clipPath>
            <mask id="mask_out_ring_behind_fire">
                <rect width="495" height="195" fill="white"/>
                <ellipse id="mask-ellipse" cx="247.5" cy="32" rx="13" ry="18" fill="black"/>
            </mask>
            
        </defs>
        <g clip-path="url(#outer_rectangle)">
            <g style="isolation: isolate">
                <rect stroke="#E4E2E2" fill="#FFFEFE" rx="4.5" x="0.5" y="0.5" width="494" height="194"/>
            </g>
            <g style="isolation: isolate">
                <line x1="165" y1="28" x2="165" y2="170" vector-effect="non-scaling-stroke" stroke-width="1" stroke="#E4E2E2" stroke-linejoin="miter" stroke-linecap="square" stroke-miterlimit="3"/>
                <line x1="330" y1="28" x2="330" y2="170" vector-effect="non-scaling-stroke" stroke-width="1" stroke="#E4E2E2" stroke-linejoin="miter" stroke-linecap="square" stroke-miterlimit="3"/>
            </g>
            <g style="isolation: isolate">
                <!-- Total Contributions big number -->
                <g transform="translate(82.5, 48)">
                    <text x="0" y="32" stroke-width="0" text-anchor="middle" fill="#151515" stroke="none" font-family="&quot;Segoe UI&quot;, Ubuntu, sans-serif" font-weight="700" font-size="28px" font-style="normal" style="opacity: 0; animation: fadein 0.5s linear forwards 0.6s">
                        ` + strconv.Itoa(totalContributions) + `
                    </text>
                </g>

                <!-- Total Contributions label -->
                <g transform="translate(82.5, 84)">
                    <text x="0" y="32" stroke-width="0" text-anchor="middle" fill="#151515" stroke="none" font-family="&quot;Segoe UI&quot;, Ubuntu, sans-serif" font-weight="400" font-size="14px" font-style="normal" style="opacity: 0; animation: fadein 0.5s linear forwards 0.7s">
                        Total Contributions
                    </text>
                </g>

                <!-- Total Contributions range -->
                <g transform="translate(82.5, 114)">
                    <text x="0" y="32" stroke-width="0" text-anchor="middle" fill="#464646" stroke="none" font-family="&quot;Segoe UI&quot;, Ubuntu, sans-serif" font-weight="400" font-size="12px" font-style="normal" style="opacity: 0; animation: fadein 0.5s linear forwards 0.8s">
                        ` + startedDate.Format("DD MMM, YYYY") + ` - Present
                    </text>
                </g>
            </g>
            <g style="isolation: isolate">
                <!-- Current Streak big number -->
                <g transform="translate(247.5, 48)">
                    <text x="0" y="32" stroke-width="0" text-anchor="middle" fill="#151515" stroke="none" font-family="&quot;Segoe UI&quot;, Ubuntu, sans-serif" font-weight="700" font-size="28px" font-style="normal" style="animation: currstreak 0.6s linear forwards">
                        ` + strconv.Itoa(streak) + `
                    </text>
                </g>

                <!-- Current Streak label -->
                <g transform="translate(247.5, 108)">
                    <text x="0" y="32" stroke-width="0" text-anchor="middle" fill="#FB8C00" stroke="none" font-family="&quot;Segoe UI&quot;, Ubuntu, sans-serif" font-weight="700" font-size="14px" font-style="normal" style="opacity: 0; animation: fadein 0.5s linear forwards 0.9s">
                        Current Streak
                    </text>
                </g>

                <!-- Current Streak range -->
                <g transform="translate(247.5, 145)">
                    <text x="0" y="21" stroke-width="0" text-anchor="middle" fill="#464646" stroke="none" font-family="&quot;Segoe UI&quot;, Ubuntu, sans-serif" font-weight="400" font-size="12px" font-style="normal" style="opacity: 0; animation: fadein 0.5s linear forwards 0.9s">
                        Nov 13 - Dec 4
                    </text>
                </g>

                <!-- Ring around number -->
                <g mask="url(#mask_out_ring_behind_fire)">
                    <circle cx="247.5" cy="71" r="40" fill="none" stroke="#FB8C00" stroke-width="5" style="opacity: 0; animation: fadein 0.5s linear forwards 0.4s"/>
                </g>
                <!-- Fire icon -->
                <g transform="translate(247.5, 19.5)" stroke-opacity="0" style="opacity: 0; animation: fadein 0.5s linear forwards 0.6s">
                    <path d="M -12 -0.5 L 15 -0.5 L 15 23.5 L -12 23.5 L -12 -0.5 Z" fill="none"/>
                    <path d="M 1.5 0.67 C 1.5 0.67 2.24 3.32 2.24 5.47 C 2.24 7.53 0.89 9.2 -1.17 9.2 C -3.23 9.2 -4.79 7.53 -4.79 5.47 L -4.76 5.11 C -6.78 7.51 -8 10.62 -8 13.99 C -8 18.41 -4.42 22 0 22 C 4.42 22 8 18.41 8 13.99 C 8 8.6 5.41 3.79 1.5 0.67 Z M -0.29 19 C -2.07 19 -3.51 17.6 -3.51 15.86 C -3.51 14.24 -2.46 13.1 -0.7 12.74 C 1.07 12.38 2.9 11.53 3.92 10.16 C 4.31 11.45 4.51 12.81 4.51 14.2 C 4.51 16.85 2.36 19 -0.29 19 Z" fill="#FB8C00" stroke-opacity="0"/>
                </g>

            </g>
            <g style="isolation: isolate">
                <!-- Longest Streak big number -->
                <g transform="translate(412.5, 48)">
                    <text x="0" y="32" stroke-width="0" text-anchor="middle" fill="#151515" stroke="none" font-family="&quot;Segoe UI&quot;, Ubuntu, sans-serif" font-weight="700" font-size="28px" font-style="normal" style="opacity: 0; animation: fadein 0.5s linear forwards 1.2s">
                        ` + strconv.Itoa(streak) + `
                    </text>
                </g>

                <!-- Longest Streak label -->
                <g transform="translate(412.5, 84)">
                    <text x="0" y="32" stroke-width="0" text-anchor="middle" fill="#151515" stroke="none" font-family="&quot;Segoe UI&quot;, Ubuntu, sans-serif" font-weight="400" font-size="14px" font-style="normal" style="opacity: 0; animation: fadein 0.5s linear forwards 1.3s">
                        Longest Streak
                    </text>
                </g>

                <!-- Longest Streak range -->
                <g transform="translate(412.5, 114)">
                    <text x="0" y="32" stroke-width="0" text-anchor="middle" fill="#464646" stroke="none" font-family="&quot;Segoe UI&quot;, Ubuntu, sans-serif" font-weight="400" font-size="12px" font-style="normal" style="opacity: 0; animation: fadein 0.5s linear forwards 1.4s">
                        Nov 13 - Dec 4
                    </text>
                </g>
            </g>
            
        </g>
    </svg>`
}
