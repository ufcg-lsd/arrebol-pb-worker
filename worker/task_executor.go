package worker

import (
	"bytes"
	"fmt"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/ufcg-lsd/arrebol-pb-worker/utils"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	TaskScriptExecutorFileName  = "task-script-executor.sh"
	RunTaskScriptCommandPattern = "/bin/bash %s -d -tsf=%s"
	DefaultWorkerDockerImage    = "ubuntu"
)

type TaskExecutor struct {
	Cli client.Client
	Cid string
}

func (e *TaskExecutor) Execute(task *Task, containerSpawnWarner chan<- interface{}) error {
	image := task.Image

	if image == "" {
		image = DefaultWorkerDockerImage
	}
	log.Println("Creating container with image: " + image)

	config := utils.ContainerConfig{
		Name:   "",
		Image:  image,
		Mounts: []mount.Mount{},
	}

	if err := e.initiate(config); err != nil {
		return err
	}
	if err := e.send(task); err != nil {
		return err
	}
	// It allows the system to know when the container
	// is ready.
	containerSpawnWarner <- "spawned"
	if err := e.run(task.Id); err != nil {
		task.State = TaskFailed
		return err
	}
	if err := e.stop(); err != nil {
		return err
	}
	task.State = TaskFinished
	return nil
}

func (e *TaskExecutor) initiate(config utils.ContainerConfig) error {
	exists, err := utils.CheckImage(&e.Cli, config.Image)
	if !exists {
		if _, err = utils.Pull(&e.Cli, config.Image); err != nil {
			return err
		}
	}
	cid, err := utils.CreateContainer(&e.Cli, config)

	if err != nil {
		return err
	}
	err = utils.StartContainer(&e.Cli, cid)

	if err != nil {
		return err
	}

	err = utils.Exec(&e.Cli, cid, "mkdir /arrebol")

	if err != nil {
		log.Println("Error on creating /arrebol folder")
		return err
	}

	taskScriptExecutorPath := os.Getenv("BIN_PATH") + "/" + TaskScriptExecutorFileName

	err = utils.Copy(&e.Cli, cid, taskScriptExecutorPath, "/arrebol/"+TaskScriptExecutorFileName)

	e.Cid = cid
	return err
}

func (e *TaskExecutor) stop() error {
	err := utils.StopContainer(&e.Cli, e.Cid)
	if err != nil {
		return err
	}
	err = utils.RemoveContainer(&e.Cli, e.Cid)
	return err
}

func (e *TaskExecutor) send(task *Task) error {
	taskScriptFileName := "task-id.ts"
	rawCmdsStr := task.Commands
	err := utils.Write(&e.Cli, e.Cid, rawCmdsStr, "/arrebol/"+taskScriptFileName)
	return err
}

func (e *TaskExecutor) run(taskId string) error {
	taskScriptFilePath := "/arrebol/task-id.ts"
	cmd := fmt.Sprintf(RunTaskScriptCommandPattern, "/arrebol/"+TaskScriptExecutorFileName, taskScriptFilePath)
	err := utils.Exec(&e.Cli, e.Cid, cmd)
	return err
}

func (e *TaskExecutor) Track() (int, error) {
	err := utils.Exec(&e.Cli, e.Cid, "touch /arrebol/task-id.ts.ec")

	if err != nil {
		log.Println(err)
	}

	ec, err := e.getExitCodes()

	if err != nil {
		log.Println(err)
		return 0, err
	}

	return len(ec), nil
}

func (e *TaskExecutor) getExitCodes() ([]int8, error) {
	ecFilePath := "/arrebol/task-id" + ".ts.ec"
	dat, err := utils.Cat(&e.Cli, e.Cid, ecFilePath)
	if err != nil {
		return nil, err
	}
	dat = bytes.TrimFunc(dat, isNotUTFNumber)
	content := string(dat[:])
	log.Println("Content: " + content)
	exitCodesStr := strings.Split(content, "\r\n")
	log.Println("ExitCodes String Array: ", exitCodesStr)
	exitCodes := toIntArray(exitCodesStr)
	return exitCodes, nil
}

func toIntArray(strs []string) []int8 {
	ints := make([]int8, 0)
	for _, s := range strs {
		x, err := strconv.Atoi(s)
		if err == nil {
			ints = append(ints, int8(x))
		}
	}
	return ints
}

func isNotUTFNumber(r rune) bool {
	if r >= 48 && r <= 57 {
		return false
	}
	return true
}
