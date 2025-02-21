package server

import (
	"encoding/json"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
)

func (s *Server) handlePickTask(w http.ResponseWriter, r *http.Request) {
	tx := s.db.Begin()
	if tx.Error != nil {
		log.Error("failed to start transaction: " + tx.Error.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var taskData TaskData
	err := tx.
		Where("status = ?", "queued").
		Order("date ASC").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&taskData).Error
	if err != nil {
		tx.Rollback()
		http.Error(w, "failed to find queued task", http.StatusNotFound)
		return
	}

	taskData.Status = statusInProgress
	now := time.Now()
	taskData.StartedAt = &now

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
	task := taskData.toTask()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(task); err != nil {
		log.Error("failed to encode response: " + err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
