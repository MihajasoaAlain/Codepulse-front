package api

import "time"

type CommitStatus struct {
	HasCommitted bool      `json:"has_committed"`
	CommitCount  int       `json:"commit_count"`
	LastCommitAt time.Time `json:"last_commit_at"`
	Streak       int       `json:"streak_days"`
}

type CodingStats struct {
	TotalCommits      int            `json:"total_commits"`
	TotalLinesAdded   int            `json:"total_lines_added"`
	TotalLinesRemoved int            `json:"total_lines_removed"`
	ActiveDays        int            `json:"active_days"`
	CurrentStreak     int            `json:"current_streak"`
	LongestStreak     int            `json:"longest_streak"`
	TopLanguages      []LanguageStat `json:"top_languages"`
	LastUpdated       time.Time      `json:"last_updated"`
}

type LanguageStat struct {
	Name       string  `json:"name"`
	Percentage float64 `json:"percentage"`
}
