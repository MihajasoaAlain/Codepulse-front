package scheduler

import (
	"log"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"

	"codepulse/api"
)

type State struct {
	HasCommitted binding.Bool
	CommitCount  binding.Int
	LastCommitAt binding.String
	Streak       binding.Int
	StatusMsg    binding.String
}

func NewState() *State {
	return &State{
		HasCommitted: binding.NewBool(),
		CommitCount:  binding.NewInt(),
		LastCommitAt: binding.NewString(),
		Streak:       binding.NewInt(),
		StatusMsg:    binding.NewString(),
	}
}

type Monitor struct {
	repo         api.Repository
	state        *State
	app          fyne.App
	pollInterval time.Duration

	mu           sync.Mutex
	reminderHour int
	notified     bool
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

func NewMonitor(
	repo api.Repository,
	state *State,
	app fyne.App,
	pollInterval time.Duration,
	defaultReminderHour int,
) *Monitor {
	return &Monitor{
		repo:         repo,
		state:        state,
		app:          app,
		pollInterval: pollInterval,
		reminderHour: defaultReminderHour,
		stopCh:       make(chan struct{}),
	}
}

func (m *Monitor) SetReminderHour(hour int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reminderHour = hour
	m.notified = false
}

func (m *Monitor) ReminderHour() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.reminderHour
}
func (m *Monitor) Start() {
	m.wg.Add(1)
	go m.run()
}

func (m *Monitor) Stop() {
	close(m.stopCh)
	m.wg.Wait()
}

func (m *Monitor) run() {
	defer m.wg.Done()

	m.tick()

	midnightTicker := time.NewTicker(1 * time.Minute)
	defer midnightTicker.Stop()

	pollTicker := time.NewTicker(m.pollInterval)
	defer pollTicker.Stop()

	for {
		select {
		case <-m.stopCh:
			return

		case <-pollTicker.C:
			m.tick()

		case t := <-midnightTicker.C:
			if t.Hour() == 0 && t.Minute() == 0 {
				m.mu.Lock()
				m.notified = false
				m.mu.Unlock()
			}
		}
	}
}

func (m *Monitor) tick() {
	_ = m.state.StatusMsg.Set("Checking commit status…")

	status, err := m.repo.GetCommitStatus()
	if err != nil {
		log.Printf("scheduler: poll error: %v", err)
		_ = m.state.StatusMsg.Set("⚠ Could not reach API: " + err.Error())
		return
	}

	_ = m.state.HasCommitted.Set(status.HasCommitted)
	_ = m.state.CommitCount.Set(status.CommitCount)
	_ = m.state.Streak.Set(status.Streak)

	if status.HasCommitted && !status.LastCommitAt.IsZero() {
		_ = m.state.LastCommitAt.Set(status.LastCommitAt.Format("15:04:05"))
	} else {
		_ = m.state.LastCommitAt.Set("—")
	}

	if status.HasCommitted {
		_ = m.state.StatusMsg.Set("✓ Committed today")
	} else {
		_ = m.state.StatusMsg.Set("✗ No commit yet today")
	}

	m.mu.Lock()
	hour := m.reminderHour
	alreadyNotified := m.notified
	m.mu.Unlock()

	now := time.Now()
	if !status.HasCommitted && !alreadyNotified && now.Hour() >= hour {
		m.sendNotification()
		m.mu.Lock()
		m.notified = true
		m.mu.Unlock()
	}
}

func (m *Monitor) sendNotification() {
	m.app.SendNotification(&fyne.Notification{
		Title:   "⚠ Coding Reminder",
		Content: "You haven't committed today yet! Open your editor and get coding. 🚀",
	})
	log.Println("scheduler: reminder notification sent")
}
