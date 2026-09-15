package dashboard

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"tracker_cli/internal/domain/entity"
)

func TestDashboardModel_TabsSwitching(t *testing.T) {
	m := NewDashboardModel()
	if m.activeTab != tabTimer {
		t.Fatalf("expected initial tab tabTimer, got %v", m.activeTab)
	}

	// Test switching to tab 2 (Plan)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	dm2 := m2.(DashboardModel)
	if dm2.activeTab != tabPlan {
		t.Errorf("expected tabPlan, got %v", dm2.activeTab)
	}

	// Test switching to tab 3 (Analytics)
	m3, _ := dm2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	dm3 := m3.(DashboardModel)
	if dm3.activeTab != tabAnalytics {
		t.Errorf("expected tabAnalytics, got %v", dm3.activeTab)
	}

	// Test switching to tab 4 (Evening)
	m4, _ := dm3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	dm4 := m4.(DashboardModel)
	if dm4.activeTab != tabEvening {
		t.Errorf("expected tabEvening, got %v", dm4.activeTab)
	}
}

func TestDashboardModel_DataLoadedMsg(t *testing.T) {
	m := NewDashboardModel()

	dataMsg := dataLoadedMsg{
		runningTask: entity.RunningTask{
			TaskName:       "work",
			Role:           "work",
			IsRunning:      true,
			TargetDuration: 25,
		},
		taskList: []entity.TaskList{
			{Name: "work", Role: "work", TimeDuration: 240, TimeDone: 60, Priority: 1},
			{Name: "english", Role: "learn", TimeDuration: 20, TimeDone: 0, Priority: 2},
		},
		restUnits:     1500, // 15.0 min
		scheduledTime: 260,
		completedTime: 60,
	}

	mUpdated, _ := m.Update(dataMsg)
	dm := mUpdated.(DashboardModel)

	if dm.loading {
		t.Errorf("expected loading false")
	}
	if dm.runningTask.TaskName != "work" {
		t.Errorf("expected running task work, got %v", dm.runningTask.TaskName)
	}
	if len(dm.taskList) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(dm.taskList))
	}
	if dm.restUnits != 1500 {
		t.Errorf("expected rest units 1500, got %d", dm.restUnits)
	}
}

