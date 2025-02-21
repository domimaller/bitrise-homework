package executor

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

type Executor struct {
	client *http.Client
	cfg    *Config
}

func New(cfg *Config) *Executor {
	e := Executor{
		cfg: cfg,
		client: &http.Client{
			Transport: &http.Transport{},
			Timeout:   10 * time.Second,
		},
	}

	return &e
}

const (
	pickTaskPath = "/tasks/pick"
)

type Task struct {
	ID      string `json:"id"`
	Command string `json:"command"`
}

type TaskResult struct {
	Status   string  `json:"status"`
	Stdout   *string `json:"stdout"`
	Stderr   *string `json:"stderr"`
	ExitCode *int    `json:"exit_code"`
}

func (e *Executor) Run() {
	log.Info("Executor started running")
	ticker := time.NewTicker(e.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Info("Picking a task")
			pickURL := "http://" + e.cfg.BackendHost + ":" + e.cfg.BackendPort + pickTaskPath
			resp, err := e.client.Get(pickURL)
			if err != nil {
				log.Errorf("error picking task: %v", err)
				continue
			}

			if resp.StatusCode != http.StatusOK {
				log.Debugf("Picking returned status code: %d", resp.StatusCode)
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				continue
			}

			var task Task
			if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
				log.Errorf("error decoding picked task: %v", err)
				resp.Body.Close()
				continue
			}
			resp.Body.Close()

			log.Infof("Executing task %s: %s", task.ID, task.Command)
			result := executeCommand(task.Command)

			e.finishTask(task.ID, result)
		}
	}
}

func executeCommand(command string) TaskResult {
	var stdoutStr, stderrStr *string
	exitCode := 0

	cmd := exec.Command("sh", "-c", command)
	log.Debugf("Executing command: %+v", cmd)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		errMsg := "failed to get stdout pipe: " + err.Error()
		log.Error(errMsg)
		return TaskResult{Status: "failed", Stderr: &errMsg, ExitCode: intPointer(1)}
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		errMsg := "failed to get stderr pipe: " + err.Error()
		log.Error(errMsg)
		return TaskResult{Status: "failed", Stderr: &errMsg, ExitCode: intPointer(1)}
	}

	if err := cmd.Start(); err != nil {
		errMsg := "failed to start command: " + err.Error()
		log.Error(errMsg)
		return TaskResult{Status: "failed", Stderr: &errMsg, ExitCode: intPointer(1)}
	}

	stdoutBytes, _ := io.ReadAll(stdoutPipe)
	stderrBytes, _ := io.ReadAll(stderrPipe)

	err = cmd.Wait()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	sOut := string(stdoutBytes)
	sErr := string(stderrBytes)
	stdoutStr = &sOut
	stderrStr = &sErr

	taskResult := TaskResult{
		Status:   "finished",
		Stdout:   stdoutStr,
		Stderr:   stderrStr,
		ExitCode: intPointer(exitCode),
	}
	log.Debugf("Task has successfuly executed: %+v", taskResult)

	return taskResult
}

func intPointer(i int) *int {
	return &i
}

func (e *Executor) finishTask(taskID string, result TaskResult) {
	finishURL := "http://" + e.cfg.BackendHost + ":" + e.cfg.BackendPort + "/tasks/" + taskID + "/finish"

	body, err := json.Marshal(result)
	if err != nil {
		log.Errorf("error marshaling task result: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, finishURL, bytes.NewReader(body))
	if err != nil {
		log.Errorf("error creating finish request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	log.Debugf("Sending finished task: %+v", result)
	resp, err := e.client.Do(req)
	if err != nil {
		log.Errorf("error sending finish request: %v", err)
		return
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Finish request for task %s returned status: %s", taskID, resp.Status)
	}
}
