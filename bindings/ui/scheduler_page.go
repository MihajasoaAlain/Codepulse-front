package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"

	"codepulse/impls/scheduler"
)

func SchedulerPage(mon *scheduler.Monitor, state *scheduler.State) fyne.CanvasObject {
	hourBinding := binding.NewFloat()
	_ = hourBinding.Set(float64(mon.ReminderHour()))

	hourLabel := widget.NewLabel(formatHour(mon.ReminderHour()))

	slider := widget.NewSliderWithData(0, 23, hourBinding)
	slider.Step = 1

	hourBinding.AddListener(binding.NewDataListener(func() {
		h, _ := hourBinding.Get()
		hour := int(h)
		hourLabel.SetText(formatHour(hour))
		mon.SetReminderHour(hour)
	}))

	reminderCard := widget.NewCard(
		"Reminder Time",
		"Send a notification if no commit is detected by this hour.",
		container.NewVBox(
			widget.NewLabel("Trigger hour (0 = midnight, 23 = 11 PM):"),
			slider,
			hourLabel,
		),
	)

	statusCard := widget.NewCard(
		"Monitor Status",
		"The background monitor polls your API on a fixed interval.",
		container.NewVBox(
			widget.NewLabelWithData(state.StatusMsg),
			widget.NewSeparator(),
			makeStatusGrid(state),
		),
	)

	logEntries := binding.NewStringList()

	logList := widget.NewListWithData(
		logEntries,
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(item binding.DataItem, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			s, _ := item.(binding.String).Get()
			label.SetText(s)
		},
	)

	state.StatusMsg.AddListener(binding.NewDataListener(func() {
		msg, _ := state.StatusMsg.Get()
		if msg == "" {
			return
		}
		existing, _ := logEntries.Get()
		entry := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg)
		updated := append([]string{entry}, existing...)
		if len(updated) > 50 {
			updated = updated[:50]
		}
		_ = logEntries.Set(updated)
	}))

	logCard := widget.NewCard(
		"Activity Log",
		"Recent monitor events (newest first).",
		container.NewVScroll(logList),
	)

	return container.NewVBox(
		widget.NewLabelWithStyle(
			"⏰ Scheduler",
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		),
		widget.NewSeparator(),
		reminderCard,
		statusCard,
		logCard,
	)
}

func makeStatusGrid(state *scheduler.State) fyne.CanvasObject {
	committedStr := binding.NewString()
	state.HasCommitted.AddListener(binding.NewDataListener(func() {
		v, _ := state.HasCommitted.Get()
		if v {
			_ = committedStr.Set("✅  Yes")
		} else {
			_ = committedStr.Set("❌  Not yet")
		}
	}))

	return container.NewGridWithColumns(2,
		widget.NewLabel("Committed today:"),
		widget.NewLabelWithData(committedStr),
		widget.NewLabel("Last commit:"),
		widget.NewLabelWithData(state.LastCommitAt),
		widget.NewLabel("Current streak:"),
		widget.NewLabelWithData(streakStr(state)),
	)
}

func streakStr(state *scheduler.State) binding.String {
	out := binding.NewString()
	state.Streak.AddListener(binding.NewDataListener(func() {
		n, _ := state.Streak.Get()
		_ = out.Set(fmt.Sprintf("%d days 🔥", n))
	}))
	return out
}

func formatHour(h int) string {
	period, display := "AM", h
	switch {
	case h == 0:
		display = 12
	case h == 12:
		period = "PM"
	case h > 12:
		display, period = h-12, "PM"
	}
	return fmt.Sprintf("Notify at %02d:00 %s  (hour %d)", display, period, h)
}
