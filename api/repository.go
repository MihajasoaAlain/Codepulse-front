package api

import (
	"fmt"
	"math/rand"
	"time"
)

type Repository interface {
	GetCommitStatus() (*CommitStatus, error)
	GetCodingStats() (*CodingStats, error)
}
type RemoteRepository struct {
	client *Client
}

func NewRemoteRepository(client *Client) Repository {
	return &RemoteRepository{client: client}
}

func (r *RemoteRepository) GetCommitStatus() (*CommitStatus, error) {
	var status CommitStatus
	if err := r.client.get("/api/v1/commits/today", &status); err != nil {
		return nil, fmt.Errorf("get commit status: %w", err)
	}
	return &status, nil
}

func (r *RemoteRepository) GetCodingStats() (*CodingStats, error) {
	var stats CodingStats
	if err := r.client.get("/api/v1/stats", &stats); err != nil {
		return nil, fmt.Errorf("get coding stats: %w", err)
	}
	return &stats, nil
}

type MockRepository struct{}

func NewMockRepository() Repository { return &MockRepository{} }

func (m *MockRepository) GetCommitStatus() (*CommitStatus, error) {
	committed := rand.Float32() > 0.3
	count := 0
	var last time.Time
	if committed {
		count = rand.Intn(5) + 1
		last = time.Now().Add(-time.Duration(rand.Intn(6)) * time.Hour)
	}
	return &CommitStatus{
		HasCommitted: committed,
		CommitCount:  count,
		LastCommitAt: last,
		Streak:       rand.Intn(30),
	}, nil
}

func (m *MockRepository) GetCodingStats() (*CodingStats, error) {
	return &CodingStats{
		TotalCommits:      rand.Intn(5000) + 100,
		TotalLinesAdded:   rand.Intn(200000) + 10000,
		TotalLinesRemoved: rand.Intn(50000) + 1000,
		ActiveDays:        rand.Intn(365) + 30,
		CurrentStreak:     rand.Intn(30),
		LongestStreak:     rand.Intn(60) + 10,
		TopLanguages: []LanguageStat{
			{Name: "Go", Percentage: 45.2},
			{Name: "TypeScript", Percentage: 30.1},
			{Name: "Python", Percentage: 15.7},
			{Name: "Bash", Percentage: 9.0},
		},
		LastUpdated: time.Now(),
	}, nil
}
