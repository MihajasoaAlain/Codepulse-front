package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"codepulse/api"
	"codepulse/bindings/ui"
	"codepulse/config"
	"codepulse/impls/scheduler"
)

func main() {
	// ── 1. Load configuration ────────────────────────────────────────────────
	cfg := config.Default()
	// In production you might do:
	//   cfg, err := config.LoadFromFile("config.yaml")

	// ── 2. Build the repository ──────────────────────────────────────────────
	// Switch to NewRemoteRepository(api.NewClient(cfg.BaseURL, cfg.APIKey))
	// once your backend is running.
	var repo api.Repository
	if cfg.APIKey == "YOUR_API_KEY_HERE" {
		log.Println("main: using MockRepository (no real API key configured)")
		repo = api.NewMockRepository()
	} else {
		client := api.NewClient(cfg.BaseURL, cfg.APIKey)
		repo = api.NewRemoteRepository(client)
	}

	// ── 3. Create the Fyne application ──────────────────────────────────────
	// app.NewWithID gives the app a stable identifier used for:
	//   • Storing preferences (window size, reminder hour, …)
	//   • System notification source identification on some platforms
	fyneApp := app.NewWithID("com.yourname.coding-reminder")
	fyneApp.SetIcon(resourceIconPng()) // see icon.go for the bundled asset

	window := fyneApp.NewWindow("Coding Reminder")
	window.Resize(fyne.NewSize(820, 560))
	window.SetFixedSize(false)

	// ── 4. Initialise shared state & scheduler ───────────────────────────────
	state := scheduler.NewState()

	// Restore the saved reminder hour from Fyne preferences (persisted to disk).
	savedHour := fyneApp.Preferences().IntWithFallback("reminder_hour", cfg.DefaultReminderHour)

	mon := scheduler.NewMonitor(repo, state, fyneApp, cfg.PollInterval, savedHour)

	// Persist preference changes whenever the user moves the slider.
	// We do this by wrapping the monitor's hour in a listener on State.StatusMsg
	// (an easy hook); in a real app you'd expose a dedicated callback on Monitor.
	// For clarity we just save on window close below.

	// ── 5. Build the UI ─────────────────────────────────────────────────────
	overviewTab := container.NewTabItem(
		"📊 Overview",
		ui.OverviewPage(repo, state),
	)
	schedulerTab := container.NewTabItem(
		"⏰ Scheduler",
		ui.SchedulerPage(mon, state),
	)

	tabs := container.NewAppTabs(overviewTab, schedulerTab)
	tabs.SetTabLocation(container.TabLocationTop)

	window.SetContent(tabs)

	// ── 6. Window lifecycle hooks ────────────────────────────────────────────
	window.SetOnClosed(func() {
		// Persist the user's chosen reminder hour before exit.
		fyneApp.Preferences().SetInt("reminder_hour", mon.ReminderHour())
		// Gracefully stop the background goroutine.
		mon.Stop()
	})

	// ── 7. Start the background monitor ─────────────────────────────────────
	// Must be called BEFORE app.Run() blocks the main goroutine.
	mon.Start()

	// ── 8. Hand control to Fyne ──────────────────────────────────────────────
	// app.Run() blocks here until the window is closed.
	// It MUST be called from the main goroutine.
	window.ShowAndRun()
}

// resourceIconPng returns a placeholder icon.
// In a real project use `fyne bundle icon.png > bundled.go` and reference the
// generated variable.  Here we return nil so Fyne uses its default icon.
func resourceIconPng() fyne.Resource {
	return nil
}

// Sentinel to ensure the widget package is used (avoids blank import if the
// compiler is very aggressive about pruning; normally not needed).
var _ = widget.NewLabel
