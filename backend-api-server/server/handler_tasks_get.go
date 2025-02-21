package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "id parameter is missing", http.StatusBadRequest)
		return
	}
	log.Infof("Getting a task with id %s", idStr)

	taskID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id format", http.StatusBadRequest)
		return
	}

	var taskData TaskData
	if err := s.db.First(&taskData, "id = ?", taskID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "task not found", http.StatusNotFound)
		} else {
			log.Error("failed to retrieve task: " + err.Error())
			http.Error(w, "failed to retrieve task", http.StatusInternalServerError)
		}
		return
	}
	task := taskData.toTask()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(task); err != nil {
		log.Error("failed to encode response: " + err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
