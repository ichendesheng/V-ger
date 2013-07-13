package task

import (
	// "encoding/base64"
	// "encoding/json"
	// "errors"
	"fmt"
	"util"
	// "io"
	"io/ioutil"
	"log"
	// "native"
	// "net/http"
	"os"
	"path"
	// "strconv"
	// "regexp"
	"strings"
	"time"
)

var TaskDir string

type Task struct {
	URL  string
	Size int64
	Name string //identifier (a little unsafe but more readable than url)
	// seconds from 1970-1-1
	StartTime int64

	DownloadedSize int64
	ElapsedTime    time.Duration
	IsNew          bool

	LimitSpeed int64
	Speed      float64
	Status     string
	NameHash   string
	Est        time.Duration

	Autoshutdown bool
}

func taskInfoFileName(name string) string {
	if !strings.HasSuffix(name, ".vger-task.txt") {
		name = fmt.Sprint(name, ".vger-task.txt")
	}
	return path.Join(TaskDir, name)
}

// func hashName(name string) string {
// 	return strings.TrimRight(base64.URLEncoding.EncodeToString([]byte(name)), "=")
// }

func GetTask(name string) (*Task, error) {
	t := new(Task)
	err := util.ReadJson(taskInfoFileName(name), t)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func GetTasks() []*Task {
	fileInfoes, err := ioutil.ReadDir(TaskDir)
	if err != nil {
		log.Print(err)
		return make([]*Task, 0)
	}

	tasks := make([]*Task, 0, len(fileInfoes))
	for _, f := range fileInfoes {
		name := f.Name()

		if strings.HasPrefix(name, ".") || f.IsDir() || !strings.HasSuffix(name, ".vger-task.txt") { //exculding hidden files
			continue
		}

		if t, err := GetTask(name); err == nil {
			tasks = append(tasks, t)
		}
	}

	return tasks
}

func GetDownloadingTask() (*Task, bool) {
	for _, t := range GetTasks() {
		if t.Status == "Downloading" {
			return t, true
		}
	}

	return nil, false
}

func SaveTask(t *Task) (err error) {
	_, err = ioutil.ReadDir(TaskDir)
	if os.IsNotExist(err) {
		os.Mkdir(TaskDir, 0777)
	}

	err = util.WriteJson(taskInfoFileName(t.Name), t)
	if err == nil {
		writeChangeEvent()
	}

	return
}

func RemoveTask(name string) error {
	err := os.Remove(taskInfoFileName(name))
	if err != nil {
		fmt.Printf("Remove task [%s] failed: %s\n", name, err)
		return err
	}

	writeChangeEvent()

	return nil
}

var chTaskChange chan []*Task

func WatchChange(ch chan []*Task) {
	chTaskChange = ch
}

func writeChangeEvent() {
	if chTaskChange != nil {
		chTaskChange <- GetTasks()
	}
}