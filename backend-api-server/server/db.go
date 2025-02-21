package server

import (
	"time"

	"github.com/google/uuid"
)

type TaskData struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	Command    string     `json:"command"`
	Date       time.Time  `json:"date" gorm:"autoCreateTime"`
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	Status     string     `json:"status"`
	Stdout     *string    `json:"stdout"`
	Stderr     *string    `json:"stderr"`
	ExitCode   *int       `json:"exit_code"`
}

func (t *Task) toTaskData() TaskData {
	return TaskData{
		ID:         t.ID,
		Command:    t.Command,
		StartedAt:  t.StartedAt,
		FinishedAt: t.FinishedAt,
		Status:     t.Status,
		Stdout:     t.Stdout,
		Stderr:     t.Stderr,
		ExitCode:   t.ExitCode,
	}
}

func (d *TaskData) toTask() Task {
	return Task{
		ID:         d.ID,
		Command:    d.Command,
		StartedAt:  d.StartedAt,
		FinishedAt: d.FinishedAt,
		Status:     d.Status,
		Stdout:     d.Stdout,
		Stderr:     d.Stderr,
		ExitCode:   d.ExitCode,
	}
}

func (d *TaskData) finish(u TaskResult) {
	now := time.Now()
	d.FinishedAt = &now
	d.Status = u.Status
	d.Stdout = u.Stdout
	d.Stderr = u.Stderr
	d.ExitCode = u.ExitCode
}
