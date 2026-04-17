package main

import (
	"log"
	"strings"

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
	cfg, err := config.LoadFromFile()
	if err != nil {
		log.Printf("main: error loading config: %v, using default", err)
		cfg = config.Default()
	}

	fyneApp := app.NewWithID("com.codepulse.app")
	fyneApp.Settings().SetTheme(ui.NewPulseTheme())
	fyneApp.SetIcon(resourceIconPng())

	window := fyneApp.NewWindow("CodePulse")

	if !cfg.IsAuthenticated() {
		showLoginPage(fyneApp, window, cfg)
		return
	}

	buildMainContent(fyneApp, window, cfg)
	window.ShowAndRun()
}

func showLoginPage(fyneApp fyne.App, window fyne.Window, cfg *config.Config) {
	window.Resize(fyne.NewSize(480, 520))
	window.SetFixedSize(true)
	window.CenterOnScreen()

	loginContent := ui.LoginPage(func(bearerToken string) {
		jwt := strings.TrimPrefix(bearerToken, "Bearer ")
		cfg.GitHubToken = jwt
		cfg.APIKey = jwt

		if err := cfg.Save(); err != nil {
			log.Printf("main: failed to save config: %v", err)
		}

		window.SetFixedSize(false)
		buildMainContent(fyneApp, window, cfg)
	})

	window.SetContent(loginContent)
	window.ShowAndRun()
}

func buildMainContent(fyneApp fyne.App, window fyne.Window, cfg *config.Config) {
	var repo api.Repository
	if cfg.APIKey == "YOUR_API_KEY_HERE" {
		log.Println("main: using MockRepository (no real API key configured)")
		repo = api.NewMockRepository()
	} else {
		client := api.NewClient(cfg.BaseURL, cfg.APIKey)
		repo = api.NewRemoteRepository(client)
	}

	window.Resize(fyne.NewSize(820, 560))

	state := scheduler.NewState()
	savedHour := fyneApp.Preferences().IntWithFallback("reminder_hour", cfg.DefaultReminderHour)
	mon := scheduler.NewMonitor(repo, state, fyneApp, cfg.PollInterval, savedHour)

	overviewTab := container.NewTabItem(
		"Overview",
		ui.OverviewPage(repo, state),
	)
	schedulerTab := container.NewTabItem(
		"Scheduler",
		ui.SchedulerPage(mon, state),
	)

	tabs := container.NewAppTabs(overviewTab, schedulerTab)
	tabs.SetTabLocation(container.TabLocationTop)

	window.SetContent(tabs)

	window.SetOnClosed(func() {
		fyneApp.Preferences().SetInt("reminder_hour", mon.ReminderHour())
		mon.Stop()
	})

	mon.Start()
}

func resourceIconPng() fyne.Resource {
	return nil
}

var _ = widget.NewLabel
