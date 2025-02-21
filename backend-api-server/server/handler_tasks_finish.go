package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TaskResult struct {
	Status   string  `json:"status"`
	Stdout   *string `json:"stdout"`
	Stderr   *string `json:"stderr"`
	ExitCode *int    `json:"exit_code"`
}

func (s *Server) handleFinishTask(w http.ResponseWriter, r *http.Request) {
	log.Info("Finishing a task")
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "id parameter is missing", http.StatusBadRequest)
		return
	}
	taskID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id format", http.StatusBadRequest)
		return
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		log.Error("failed to start transaction: " + tx.Error.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var taskData TaskData
	err = tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&taskData, "id = ?", taskID).Error
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "task not found", http.StatusNotFound)
		} else {
			log.Error("failed to retrieve task: " + err.Error())
			http.Error(w, "failed to retrieve task", http.StatusInternalServerError)
		}
		return
	}

	var taskResult TaskResult
	if err := json.NewDecoder(r.Body).Decode(&taskResult); err != nil {
		tx.Rollback()
		log.Error("failed to decode request body: " + err.Error())
		http.Error(w, "failed to decode request body", http.StatusBadRequest)
		return
	}

	taskData.finish(taskResult)
	if err := tx.Save(&taskData).Error; err != nil {
		tx.Rollback()
		log.Error("failed to update task: " + err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		log.Error("failed to commit transaction: " + err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	updatedTask := taskData.toTask()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(updatedTask); err != nil {
		log.Error("failed to encode response: " + err.Error())
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
