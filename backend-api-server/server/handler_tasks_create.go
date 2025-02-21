package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	statusQueued     = "queued"
	statusInProgress = "in_progress"
	statusFinished   = "finished"
)

type Task struct {
	ID         uuid.UUID  `json:"id"`
	Command    string     `json:"command"`
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	Status     string     `json:"status"`
	Stdout     *string    `json:"stdout"`
	Stderr     *string    `json:"stderr"`
	ExitCode   *int       `json:"exit_code"`
}

type TaskCreate struct {
	Command string `json:"command"`
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	log.Info("Creating new task")
	var taskCreate TaskCreate
	if err := json.NewDecoder(r.Body).Decode(&taskCreate); err != nil {
		log.Error("invalid request payload: " + err.Error())
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}
	task := Task{Command: taskCreate.Command}
	task.ID = uuid.New()
	task.Status = statusQueued

	taskData := task.toTaskData()
	if err := s.db.Create(&taskData).Error; err != nil {
		log.Error("failed to save task: " + err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(task); err != nil {
		log.Error("failed to encode response: " + err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
