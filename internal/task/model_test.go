package task

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProject(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "My Project", false},
		{"empty name", "", true},
		{"name with spaces", "  Project Name  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, err := NewProject(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, project)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, project)
				assert.Equal(t, tt.input, project.GetName())
				assert.Empty(t, project.GetPhases())
			}
		})
	}
}

func TestProject_AddPhase(t *testing.T) {
	project, err := NewProject("Test Project")
	require.NoError(t, err)

	phase1, err := NewPhase("Backlog", "Test Project")
	require.NoError(t, err)

	phase2, err := NewPhase("In Progress", "Test Project")
	require.NoError(t, err)

	project.AddPhase(phase1)
	project.AddPhase(phase2)

	phases := project.GetPhases()
	assert.Len(t, phases, 2)
	assert.Equal(t, "Backlog", phases[0].GetName())
	assert.Equal(t, "In Progress", phases[1].GetName())
}

func TestNewPhase(t *testing.T) {
	tests := []struct {
		name        string
		phaseName   string
		projectName string
		wantStatus  Status
		wantErr     bool
	}{
		{"pending phase", "Backlog", "Project A", StatusPending, false},
		{"running phase", "🏃 In Progress", "Project A", StatusRunning, false},
		{"scheduled phase", "🗓️ Scheduled", "Project A", StatusScheduled, false},
		{"completed phase", "✅ Done", "Project A", StatusCompleted, false},
		{"empty name", "", "Project A", StatusPending, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phase, err := NewPhase(tt.phaseName, tt.projectName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, phase)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, phase)
				assert.Equal(t, tt.phaseName, phase.GetName())
				assert.Equal(t, tt.wantStatus, phase.GetAssociatedStatus())
			}
		})
	}
}

func TestNewTask(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		wantErr bool
	}{
		{"valid title", "Implement feature", false},
		{"empty title", "", true},
		{"title with spaces", "  Task Title  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := NewTask(tt.title)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, task)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, task)
				assert.Equal(t, tt.title, task.GetTitle())
				assert.Empty(t, task.GetTags())
				assert.Empty(t, task.GetActiveReminders())
			}
		})
	}
}

func TestTask_AddComment(t *testing.T) {
	task, err := NewTask("Test Task")
	require.NoError(t, err)

	t.Run("add valid comment", func(t *testing.T) {
		err := task.AddComment("This is a comment")
		assert.NoError(t, err)
	})

	t.Run("add empty comment", func(t *testing.T) {
		err := task.AddComment("")
		assert.Error(t, err)
	})

	t.Run("add multiple comments", func(t *testing.T) {
		err := task.AddComment("Second comment")
		assert.NoError(t, err)
	})
}

func TestTask_AddReminder(t *testing.T) {
	task, err := NewTask("Test Task")
	require.NoError(t, err)

	reminder := &Reminder{
		Label:        "@remind (25-01-01 10:00:00)",
		Time:         time.Now().Add(24 * time.Hour),
		Acknowledged: false,
	}

	task.AddReminder(reminder)

	activeReminders := task.GetActiveReminders()
	assert.Len(t, activeReminders, 1)
	assert.Equal(t, reminder, activeReminders[0])
}

func TestTask_SetDueDate(t *testing.T) {
	task, err := NewTask("Test Task")
	require.NoError(t, err)

	dueDate := time.Now().Add(7 * 24 * time.Hour)

	t.Run("set due date first time", func(t *testing.T) {
		err := task.SetDueDate(&dueDate)
		assert.NoError(t, err)
	})

	t.Run("set due date second time", func(t *testing.T) {
		newDueDate := time.Now().Add(14 * 24 * time.Hour)
		err := task.SetDueDate(&newDueDate)
		assert.Error(t, err)
	})
}

func TestTask_AddTag(t *testing.T) {
	task, err := NewTask("Test Task")
	require.NoError(t, err)

	t.Run("add unique tags", func(t *testing.T) {
		err := task.AddTag("urgent")
		assert.NoError(t, err)

		err = task.AddTag("important")
		assert.NoError(t, err)

		tags := task.GetTags()
		assert.Len(t, tags, 2)
		assert.ElementsMatch(t, []string{"urgent", "important"}, tags)
	})

	t.Run("add duplicate tag", func(t *testing.T) {
		err := task.AddTag("urgent")
		assert.Error(t, err)
	})
}

func TestPhase_AddTask(t *testing.T) {
	phase, err := NewPhase("Backlog", "Test Project")
	require.NoError(t, err)

	task1, err := NewTask("Task 1")
	require.NoError(t, err)

	task2, err := NewTask("Task 2")
	require.NoError(t, err)

	phase.AddTask(task1)
	phase.AddTask(task2)

	tasks := phase.GetTasks()
	assert.Len(t, tasks, 2)
	assert.Equal(t, "Task 1", tasks[0].GetTitle())
	assert.Equal(t, "Task 2", tasks[1].GetTitle())
}

func TestProject_GetPhaseByName(t *testing.T) {
	project, err := NewProject("Test Project")
	require.NoError(t, err)

	backlogPhase, err := NewPhase("Backlog", "Test Project")
	require.NoError(t, err)

	inProgressPhase, err := NewPhase("In Progress", "Test Project")
	require.NoError(t, err)

	project.AddPhase(backlogPhase)
	project.AddPhase(inProgressPhase)

	t.Run("get existing phase", func(t *testing.T) {
		phase, err := project.GetPhaseByName("Backlog")
		assert.NoError(t, err)
		assert.Equal(t, backlogPhase, phase)
	})

	t.Run("get non-existing phase", func(t *testing.T) {
		phase, err := project.GetPhaseByName("Non-existent")
		assert.Error(t, err)
		assert.Nil(t, phase)
	})
}

func TestPhase_GetTaskTitles(t *testing.T) {
	phase, err := NewPhase("Backlog", "Test Project")
	require.NoError(t, err)

	task1, err := NewTask("Task 1")
	require.NoError(t, err)

	task2, err := NewTask("Task 2")
	require.NoError(t, err)

	phase.AddTask(task1)
	phase.AddTask(task2)

	titles := phase.GetTaskTitles()
	assert.Len(t, titles, 2)
	assert.Equal(t, []string{"Task 1", "Task 2"}, titles)
}

func TestTask_GetOverdueReminders(t *testing.T) {
	task, err := NewTask("Test Task")
	require.NoError(t, err)

	pastTime := time.Now().Add(-24 * time.Hour)
	futureTime := time.Now().Add(24 * time.Hour)

	// Add overdue reminder (unacknowledged, past time)
	task.AddReminder(&Reminder{
		Label:        "Overdue",
		Time:         pastTime,
		Acknowledged: false,
	})

	// Add future reminder (unacknowledged, future time)
	task.AddReminder(&Reminder{
		Label:        "Future",
		Time:         futureTime,
		Acknowledged: false,
	})

	// Add acknowledged reminder (acknowledged, past time)
	task.AddReminder(&Reminder{
		Label:        "Acknowledged",
		Time:         pastTime,
		Acknowledged: true,
	})

	overdueReminders := task.GetOverdueReminders()
	assert.Len(t, overdueReminders, 1)
	assert.Equal(t, "Overdue", overdueReminders[0].Label)
	assert.Equal(t, "Test Task", overdueReminders[0].TaskTitle)
}

func TestPhase_GetOverdueReminders(t *testing.T) {
	phase, err := NewPhase("Backlog", "Test Project")
	require.NoError(t, err)

	task1, err := NewTask("Task 1")
	require.NoError(t, err)

	task2, err := NewTask("Task 2")
	require.NoError(t, err)

	// Add overdue reminder to task1
	task1.AddReminder(&Reminder{
		Label:        "Task1 Overdue",
		Time:         time.Now().Add(-24 * time.Hour),
		Acknowledged: false,
	})

	// Add overdue reminder to task2
	task2.AddReminder(&Reminder{
		Label:        "Task2 Overdue",
		Time:         time.Now().Add(-12 * time.Hour),
		Acknowledged: false,
	})

	phase.AddTask(task1)
	phase.AddTask(task2)

	overdueReminders := phase.GetOverdueReminders()
	assert.Len(t, overdueReminders, 2)
	assert.Equal(t, "Task1 Overdue", overdueReminders[0].Label)
	assert.Equal(t, "Task2 Overdue", overdueReminders[1].Label)
}

func TestProject_GetOverdueReminders(t *testing.T) {
	project, err := NewProject("Test Project")
	require.NoError(t, err)

	backlogPhase, err := NewPhase("Backlog", "Test Project")
	require.NoError(t, err)

	inProgressPhase, err := NewPhase("In Progress", "Test Project")
	require.NoError(t, err)

	task1, err := NewTask("Task 1")
	require.NoError(t, err)

	task2, err := NewTask("Task 2")
	require.NoError(t, err)

	// Add overdue reminder to task1 in backlog
	task1.AddReminder(&Reminder{
		Label:        "Backlog Overdue",
		Time:         time.Now().Add(-24 * time.Hour),
		Acknowledged: false,
	})

	// Add overdue reminder to task2 in progress
	task2.AddReminder(&Reminder{
		Label:        "In Progress Overdue",
		Time:         time.Now().Add(-12 * time.Hour),
		Acknowledged: false,
	})

	backlogPhase.AddTask(task1)
	inProgressPhase.AddTask(task2)

	project.AddPhase(backlogPhase)
	project.AddPhase(inProgressPhase)

	overdueReminders := project.GetOverdueReminders()
	assert.Len(t, overdueReminders, 2)
}

func TestTask_GetActiveReminders(t *testing.T) {
	task, err := NewTask("Test Task")
	require.NoError(t, err)

	// Add active reminder (unacknowledged)
	task.AddReminder(&Reminder{
		Label:        "Active",
		Time:         time.Now(),
		Acknowledged: false,
	})

	// Add acknowledged reminder
	task.AddReminder(&Reminder{
		Label:        "Acknowledged",
		Time:         time.Now(),
		Acknowledged: true,
	})

	activeReminders := task.GetActiveReminders()
	assert.Len(t, activeReminders, 1)
	assert.Equal(t, "Active", activeReminders[0].Label)
}

func TestPhase_GetWarnings(t *testing.T) {
	t.Run("scheduled phase with no reminders", func(t *testing.T) {
		phase, err := NewPhase("🗓️ Scheduled", "Test Project")
		require.NoError(t, err)

		task, err := NewTask("Scheduled Task")
		require.NoError(t, err)

		phase.AddTask(task)

		warnings := phase.GetWarnings()
		assert.Len(t, warnings, 1)
		assert.Contains(t, warnings[0], "has no reminders active")
	})

	t.Run("scheduled phase with reminders", func(t *testing.T) {
		phase, err := NewPhase("🗓️ Scheduled", "Test Project")
		require.NoError(t, err)

		task, err := NewTask("Scheduled Task")
		require.NoError(t, err)

		task.AddReminder(&Reminder{
			Label:        "@remind (25-01-01 10:00:00)",
			Time:         time.Now().Add(24 * time.Hour),
			Acknowledged: false,
		})

		phase.AddTask(task)

		warnings := phase.GetWarnings()
		assert.Empty(t, warnings)
	})

	t.Run("non-scheduled phase", func(t *testing.T) {
		phase, err := NewPhase("Backlog", "Test Project")
		require.NoError(t, err)

		task, err := NewTask("Backlog Task")
		require.NoError(t, err)

		phase.AddTask(task)

		warnings := phase.GetWarnings()
		assert.Empty(t, warnings)
	})
}

func TestProject_GetWarnings(t *testing.T) {
	project, err := NewProject("Test Project")
	require.NoError(t, err)

	backlogPhase, err := NewPhase("Backlog", "Test Project")
	require.NoError(t, err)

	scheduledPhase, err := NewPhase("🗓️ Scheduled", "Test Project")
	require.NoError(t, err)

	// Add task without reminder to scheduled phase
	task, err := NewTask("Scheduled Task")
	require.NoError(t, err)

	scheduledPhase.AddTask(task)

	project.AddPhase(backlogPhase)
	project.AddPhase(scheduledPhase)

	warnings := project.GetWarnings()
	assert.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "has no reminders active")
}
