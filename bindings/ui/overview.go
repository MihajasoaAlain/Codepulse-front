package ui

import (
	"fmt"
	"log"

	"codepulse/api"
	"codepulse/impls/scheduler"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func OverviewPage(repo api.Repository, state *scheduler.State) fyne.CanvasObject {
	commitIcon := widget.NewLabel("⏳")
	commitIcon.TextStyle = fyne.TextStyle{Bold: true}
	commitIcon.Alignment = fyne.TextAlignCenter

	statusLabel := widget.NewLabelWithData(state.StatusMsg)
	statusLabel.Alignment = fyne.TextAlignCenter
	statusLabel.Wrapping = fyne.TextWrapWord

	commitCountBinding := binding.NewString()
	commitCountLabel := widget.NewLabelWithData(commitCountBinding)
	commitCountLabel.Alignment = fyne.TextAlignCenter

	streakBinding := binding.NewString()
	streakLabel := widget.NewLabelWithData(streakBinding)
	streakLabel.Alignment = fyne.TextAlignCenter

	lastCommitLabel := widget.NewLabelWithData(state.LastCommitAt)
	lastCommitLabel.Alignment = fyne.TextAlignCenter

	state.CommitCount.AddListener(binding.NewDataListener(func() {
		n, _ := state.CommitCount.Get()
		_ = commitCountBinding.Set(fmt.Sprintf("Commits today: %d", n))
	}))
	state.Streak.AddListener(binding.NewDataListener(func() {
		n, _ := state.Streak.Get()
		_ = streakBinding.Set(fmt.Sprintf("🔥 Streak: %d days", n))
	}))
	// Update icon when committed status changes.
	state.HasCommitted.AddListener(binding.NewDataListener(func() {
		committed, _ := state.HasCommitted.Get()
		if committed {
			commitIcon.SetText("✅")
		} else {
			commitIcon.SetText("❌")
		}
	}))

	commitCard := widget.NewCard(
		"Today's Commit",
		"",
		container.NewVBox(
			commitIcon,
			statusLabel,
			commitCountLabel,
			lastCommitLabel,
			streakLabel,
		),
	)

	statsTitle := canvas.NewText("Coding Statistics", theme.ForegroundColor())
	statsTitle.TextStyle = fyne.TextStyle{Bold: true}
	statsTitle.TextSize = 14

	totalCommitsLabel := widget.NewLabel("Total commits: —")
	linesAddedLabel := widget.NewLabel("Lines added: —")
	linesRemovedLabel := widget.NewLabel("Lines removed: —")
	activeDaysLabel := widget.NewLabel("Active days: —")
	longestStreakLabel := widget.NewLabel("Longest streak: —")
	langLabel := widget.NewLabel("Top language: —")

	statsLoading := widget.NewProgressBarInfinite()
	statsLoading.Hide()

	refreshStats := func() {
		statsLoading.Show()
		go func() {
			stats, err := repo.GetCodingStats()
			if err != nil {
				log.Printf("overview: fetch stats: %v", err)
				statsLoading.Hide()
				return
			}

			totalCommitsLabel.SetText(fmt.Sprintf("Total commits: %d", stats.TotalCommits))
			linesAddedLabel.SetText(fmt.Sprintf("Lines added: %s", formatInt(stats.TotalLinesAdded)))
			linesRemovedLabel.SetText(fmt.Sprintf("Lines removed: %s", formatInt(stats.TotalLinesRemoved)))
			activeDaysLabel.SetText(fmt.Sprintf("Active days: %d", stats.ActiveDays))
			longestStreakLabel.SetText(fmt.Sprintf("Longest streak: %d days", stats.LongestStreak))
			if len(stats.TopLanguages) > 0 {
				top := stats.TopLanguages[0]
				langLabel.SetText(fmt.Sprintf("Top language: %s (%.1f%%)", top.Name, top.Percentage))
			}
			statsLoading.Hide()
		}()
	}

	refreshBtn := widget.NewButtonWithIcon("Refresh Stats", theme.ViewRefreshIcon(), refreshStats)

	statsCard := widget.NewCard(
		"All-time Stats",
		"",
		container.NewVBox(
			statsLoading,
			totalCommitsLabel,
			linesAddedLabel,
			linesRemovedLabel,
			activeDaysLabel,
			longestStreakLabel,
			langLabel,
			refreshBtn,
		),
	)

	go refreshStats()

	return container.NewVBox(
		widget.NewLabelWithStyle(
			"📊 Overview",
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			commitCard,
			statsCard,
		),
	)
}
func formatInt(n int) string {
	s := fmt.Sprintf("%d", n)
	result := make([]byte, 0, len(s)+len(s)/3)
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}
