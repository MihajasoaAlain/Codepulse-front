package ui

import (
	"image/color"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"codepulse/auth"
)

const (
	callbackPort = 19876
	backendURL   = "http://localhost:8080"
)

func LoginPage(onLoginSuccess func(token string)) fyne.CanvasObject {
	ghIcon := canvas.NewImageFromResource(GitHubIconRes)
	ghIcon.FillMode = canvas.ImageFillContain
	ghIcon.SetMinSize(fyne.NewSize(96, 96))

	title := canvas.NewText("CodePulse", color.NRGBA{R: 230, G: 237, B: 243, A: 255})
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 28
	title.Alignment = fyne.TextAlignCenter

	subtitle := canvas.NewText("Track your GitHub activity", color.NRGBA{R: 125, G: 133, B: 144, A: 255})
	subtitle.TextSize = 14
	subtitle.Alignment = fyne.TextAlignCenter

	statusLabel := canvas.NewText("", color.NRGBA{R: 248, G: 81, B: 73, A: 255})
	statusLabel.TextSize = 12
	statusLabel.Alignment = fyne.TextAlignCenter

	signInBtn := widget.NewButton("Sign in with GitHub", nil)
	signInBtn.Importance = widget.HighImportance

	signInBtn.OnTapped = func() {
		signInBtn.Disable()
		signInBtn.SetText("Waiting for GitHub...")
		statusLabel.Text = ""
		statusLabel.Refresh()

		resultCh, shutdown := auth.StartCallbackServer(callbackPort)

		go func() {
			redirectURL := "http://127.0.0.1:19876/auth/callback"
			loginURL := backendURL + "/auth/github/login?redirect=" + url.QueryEscape(redirectURL)

			parsed, _ := url.Parse(loginURL)
			fyne.CurrentApp().OpenURL(parsed)

			result := <-resultCh
			shutdown()

			if result.Err != nil {
				statusLabel.Text = "Authentication failed"
				statusLabel.Refresh()
				signInBtn.Enable()
				signInBtn.SetText("Sign in with GitHub")
				return
			}

			onLoginSuccess(result.Token)
		}()
	}

	cardContent := container.NewVBox(
		container.NewCenter(ghIcon),
		spacer(12),
		title,
		subtitle,
		spacer(32),
		signInBtn,
		spacer(4),
		statusLabel,
	)

	card := widget.NewCard("", "", cardContent)

	return container.NewCenter(
		container.NewPadded(card),
	)
}

func spacer(height float32) fyne.CanvasObject {
	s := canvas.NewRectangle(color.Transparent)
	s.SetMinSize(fyne.NewSize(0, height))
	return s
}
