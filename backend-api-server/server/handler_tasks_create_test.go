package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestHandlerTaskCreateOK(t *testing.T) {
	assert := assert.New(t)

	// Mock
	mockDB, mock, err := sqlmock.New()
	assert.NoError(err)
	defer func() { _ = mockDB.Close() }()

	dialector := postgres.New(postgres.Config{
		Conn:                 mockDB,
		DriverName:           "postgres",
		PreferSimpleProtocol: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(err)

	// Create task OK
	server := Server{db: db}

	taskPayload := map[string]interface{}{
		"command": "test command",
	}
	payloadBytes, err := json.Marshal(taskPayload)
	assert.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mock.ExpectBegin()
	insertQuery := regexp.QuoteMeta(`INSERT INTO "task_data"`)
	mock.ExpectExec(insertQuery).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	server.handleCreateTask(w, req)

	resp := w.Result()
	assert.Equal(http.StatusCreated, resp.StatusCode)

	var returnedTask map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&returnedTask)
	assert.NoError(err)
	assert.Equal("test command", returnedTask["command"])
	assert.Equal("queued", returnedTask["status"])
	assert.Contains(returnedTask, "id")

	// Create task - custom payload: only command can be given
	taskCustomPayload := map[string]interface{}{
		"id":          "my_custom_id",
		"command":     "some command",
		"started_at":  "2025-02-21T18:16:36.756116Z",
		"finished_at": "2025-02-21T18:16:46.771742Z",
		"status":      "finished",
		"stdout":      "random",
		"stderr":      "also",
		"exit_code":   110,
	}
	payloadBytes, err = json.Marshal(taskCustomPayload)
	assert.NoError(err)

	req = httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	mock.ExpectBegin()
	mock.ExpectExec(insertQuery).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	server.handleCreateTask(w, req)

	resp = w.Result()
	assert.Equal(http.StatusCreated, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&returnedTask)
	assert.NoError(err)
	assert.Equal("some command", returnedTask["command"])
	assert.Equal("queued", returnedTask["status"])
	assert.Contains(returnedTask, "id")
	assert.NotEqual("my_custom_id", returnedTask["id"])
	assert.Nil(returnedTask["started_at"])
	assert.Nil(returnedTask["finished_at"])
	assert.Nil(returnedTask["stdout"])
	assert.Nil(returnedTask["stderr"])
	assert.Nil(returnedTask["exit_code"])

	err = mock.ExpectationsWereMet()
	assert.NoError(err)
}

func TestHandlerTaskCreateNOK(t *testing.T) {
	assert := assert.New(t)

	// Mock
	mockDB, mock, err := sqlmock.New()
	assert.NoError(err)
	defer mockDB.Close()

	dialector := postgres.New(postgres.Config{
		Conn:                 mockDB,
		DriverName:           "postgres",
		PreferSimpleProtocol: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(err)

	mock.ExpectBegin()
	insertQuery := regexp.QuoteMeta(`INSERT INTO "task_data"`)
	mock.ExpectExec(insertQuery).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	// Init
	server := Server{db: db}

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader([]byte("not a json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Create task - invalid json
	server.handleCreateTask(w, req)

	resp := w.Result()
	assert.Equal(http.StatusBadRequest, resp.StatusCode)

	// Validate
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(string(body), "invalid request payload")

	// Init
	taskPayload := map[string]interface{}{
		"command": "test command",
	}
	payloadBytes, err := json.Marshal(taskPayload)
	assert.NoError(err)

	req = httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	// Create task - cannot save to DB
	server.handleCreateTask(w, req)

	// Validate
	resp = w.Result()
	assert.Equal(http.StatusInternalServerError, resp.StatusCode)
	body, _ = io.ReadAll(resp.Body)
	assert.Contains(string(body), "Internal Server Error")
	assert.NoError(mock.ExpectationsWereMet())
}
