package server

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	log.Info("Listing tasks")
	var tasksData []TaskData
	if err := s.db.Find(&tasksData).Error; err != nil {
		log.Error("failed to retrieve tasks: " + err.Error())
		http.Error(w, "failed to retrieve tasks", http.StatusInternalServerError)
		return
	}

	tasks := make([]Task, len(tasksData))
	for i, td := range tasksData {
		tasks[i] = td.toTask()
	}

	response := map[string]interface{}{
		"tasks": tasks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(response); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
