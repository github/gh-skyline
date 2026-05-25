package skyline

import (
	"testing"

	"github.com/github/gh-skyline/internal/github"
	"github.com/github/gh-skyline/internal/testutil/fixtures"
	"github.com/github/gh-skyline/internal/testutil/mocks"
	"github.com/github/gh-skyline/internal/types"
)

func TestGenerateSkyline(t *testing.T) {
	// Save original initializer
	originalInit := github.InitializeGitHubClient
	defer func() {
		github.InitializeGitHubClient = originalInit
	}()

	tests := []struct {
		name       string
		startYear  int
		endYear    int
		targetUser string
		full       bool
		mockClient *mocks.MockGitHubClient
		wantErr    bool
	}{
		{
			name:       "single year",
			startYear:  2024,
			endYear:    2024,
			targetUser: "testuser",
			full:       false,
			mockClient: &mocks.MockGitHubClient{
				Username: "testuser",
				JoinYear: 2020,
				MockData: fixtures.GenerateContributionsResponse("testuser", 2024),
			},
			wantErr: false,
		},
		{
			name:       "year range",
			startYear:  2020,
			endYear:    2024,
			targetUser: "testuser",
			full:       false,
			mockClient: &mocks.MockGitHubClient{
				Username: "testuser",
				JoinYear: 2020,
				MockData: fixtures.GenerateContributionsResponse("testuser", 2024),
			},
			wantErr: false,
		},
		{
			name:       "full range",
			startYear:  2008,
			endYear:    2024,
			targetUser: "testuser",
			full:       true,
			mockClient: &mocks.MockGitHubClient{
				Username: "testuser",
				JoinYear: 2008,
				MockData: fixtures.GenerateContributionsResponse("testuser", 2024),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			github.InitializeGitHubClient = func() (*github.Client, error) {
				return github.NewClient(tt.mockClient), nil
			}

			err := GenerateSkyline(tt.startYear, tt.endYear, tt.targetUser, "", tt.full, "", false)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSkyline() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateSkylineWithOrg(t *testing.T) {
	originalInit := github.InitializeGitHubClient
	defer func() {
		github.InitializeGitHubClient = originalInit
	}()

	tests := []struct {
		name       string
		startYear  int
		endYear    int
		targetUser string
		org        string
		mockClient *mocks.MockGitHubClient
		wantErr    bool
	}{
		{
			name:       "org filter single year",
			startYear:  2024,
			endYear:    2024,
			targetUser: "testuser",
			org:        "testorg",
			mockClient: &mocks.MockGitHubClient{
				Username: "testuser",
				JoinYear: 2020,
			},
			wantErr: false,
		},
		{
			name:       "org filter year range",
			startYear:  2023,
			endYear:    2024,
			targetUser: "testuser",
			org:        "testorg",
			mockClient: &mocks.MockGitHubClient{
				Username: "testuser",
				JoinYear: 2020,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			github.InitializeGitHubClient = func() (*github.Client, error) {
				return github.NewClient(tt.mockClient), nil
			}

			err := GenerateSkyline(tt.startYear, tt.endYear, tt.targetUser, tt.org, false, "", true)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSkyline() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildContributionGrid(t *testing.T) {
	tests := []struct {
		name        string
		year        int
		dailyCounts map[string]int
		wantWeeks   int
	}{
		{
			name:        "empty contributions",
			year:        2024,
			dailyCounts: map[string]int{},
			wantWeeks:   53,
		},
		{
			name: "single day contribution",
			year: 2024,
			dailyCounts: map[string]int{
				"2024-06-15": 5,
			},
			wantWeeks: 53,
		},
		{
			name: "multiple days",
			year: 2024,
			dailyCounts: map[string]int{
				"2024-01-15": 3,
				"2024-06-15": 5,
				"2024-12-25": 1,
			},
			wantWeeks: 53,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := buildContributionGrid(tt.year, tt.dailyCounts)

			if len(grid) < 52 || len(grid) > 54 {
				t.Errorf("expected ~53 weeks, got %d", len(grid))
			}

			for date, expectedCount := range tt.dailyCounts {
				found := false
				for _, week := range grid {
					for _, day := range week {
						if day.Date == date {
							found = true
							if day.ContributionCount != expectedCount {
								t.Errorf("date %s: expected count %d, got %d", date, expectedCount, day.ContributionCount)
							}
						}
					}
				}
				if !found {
					t.Errorf("date %s not found in grid", date)
				}
			}
		})
	}
}

func TestFetchOrgContributionDataFiltering(t *testing.T) {
	originalInit := github.InitializeGitHubClient
	defer func() {
		github.InitializeGitHubClient = originalInit
	}()

	mockOrgData := &types.OrgContributionsResponse{}
	mockOrgData.User.Login = "testuser"
	mockOrgData.User.ContributionsCollection.CommitContributionsByRepository = []struct {
		Repository struct {
			Name  string `json:"name"`
			Owner struct {
				Login string `json:"login"`
			} `json:"owner"`
		} `json:"repository"`
		Contributions struct {
			TotalCount int `json:"totalCount"`
			Nodes      []struct {
				OccurredAt string `json:"occurredAt"`
			} `json:"nodes"`
		} `json:"contributions"`
	}{
		{
			Repository: struct {
				Name  string `json:"name"`
				Owner struct {
					Login string `json:"login"`
				} `json:"owner"`
			}{
				Name: "included-repo",
				Owner: struct {
					Login string `json:"login"`
				}{Login: "targetorg"},
			},
			Contributions: struct {
				TotalCount int `json:"totalCount"`
				Nodes      []struct {
					OccurredAt string `json:"occurredAt"`
				} `json:"nodes"`
			}{
				TotalCount: 2,
				Nodes: []struct {
					OccurredAt string `json:"occurredAt"`
				}{
					{OccurredAt: "2024-03-15T10:00:00Z"},
					{OccurredAt: "2024-03-15T14:00:00Z"},
				},
			},
		},
		{
			Repository: struct {
				Name  string `json:"name"`
				Owner struct {
					Login string `json:"login"`
				} `json:"owner"`
			}{
				Name: "excluded-repo",
				Owner: struct {
					Login string `json:"login"`
				}{Login: "otherorg"},
			},
			Contributions: struct {
				TotalCount int `json:"totalCount"`
				Nodes      []struct {
					OccurredAt string `json:"occurredAt"`
				} `json:"nodes"`
			}{
				TotalCount: 5,
				Nodes: []struct {
					OccurredAt string `json:"occurredAt"`
				}{
					{OccurredAt: "2024-03-16T10:00:00Z"},
				},
			},
		},
	}

	mockClient := &mocks.MockGitHubClient{
		Username:    "testuser",
		MockOrgData: mockOrgData,
	}

	github.InitializeGitHubClient = func() (*github.Client, error) {
		return github.NewClient(mockClient), nil
	}

	client, _ := github.InitializeGitHubClient()
	grid, err := fetchOrgContributionData(client, "testuser", "targetorg", 2024)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	march15Count := 0
	march16Count := 0
	for _, week := range grid {
		for _, day := range week {
			if day.Date == "2024-03-15" {
				march15Count = day.ContributionCount
			}
			if day.Date == "2024-03-16" {
				march16Count = day.ContributionCount
			}
		}
	}

	// With quarterly queries, the mock returns data 4 times (once per quarter), so counts are 4x
	// The key test is that targetorg contributions are included and otherorg contributions are filtered
	if march15Count == 0 {
		t.Errorf("expected contributions on 2024-03-15 (targetorg), got %d", march15Count)
	}
	if march16Count != 0 {
		t.Errorf("expected 0 contributions on 2024-03-16 (otherorg should be filtered), got %d", march16Count)
	}
}
