package gtodo

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/y-yagi/cacher"
	"github.com/y-yagi/configure"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	tasks "google.golang.org/api/tasks/v1"
)

// Service is a todo module.
type Service struct {
	taskService *tasks.Service
	cacher      *cacher.Cacher
}

const (
	taskListsCacheKey = "tasklists"
	tasksCacheKey     = "tasks"
)

// NewService create a new service.
func NewService() (*Service, error) {
	// TODO(y-yagi) Consider cache path
	cacher := cacher.WithFileStore("/tmp/")
	srv := &Service{cacher: cacher}

	if err := srv.buildTaskService(); err != nil {
		return nil, err
	}

	return srv, nil
}

func (srv *Service) buildTaskService() error {
	b, err := ioutil.ReadFile(srv.credentialsPath())
	if err != nil {
		return errors.Wrap(err, "Unable to read client secret file")
	}

	config, err := google.ConfigFromJSON(b, tasks.TasksScope)
	if err != nil {
		return errors.Wrap(err, "Unable to parse client secret file to config")
	}

	client, err := srv.getClient(config)
	if err != nil {
		return errors.Wrap(err, "Unable to get Client")
	}

	ctx := context.Background()
	srv.taskService, err = tasks.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve tasks Client")
	}

	return nil
}

// TasklistsService return TasklistsService.
func (srv *Service) TasklistsService() *tasks.TasklistsService {
	return srv.taskService.Tasklists
}

// TaskLists return TaskLists.
func (srv *Service) TaskLists() (*tasks.TaskLists, error) {
	data, err := srv.cacher.Read(taskListsCacheKey)
	if data != nil && err == nil {
		var tList tasks.TaskLists
		err = json.Unmarshal(data, &tList)
		return &tList, err
	}

	tList, err := srv.taskService.Tasklists.List().MaxResults(20).Do()
	if err != nil {
		return nil, err
	}

	json, err := tList.MarshalJSON()
	if err != nil {
		return nil, err
	}
	srv.cacher.Write(taskListsCacheKey, json, cacher.Forever)

	return tList, nil
}

// InsertTaskList insert new TaskList.
func (srv *Service) InsertTaskList(taskList *tasks.TaskList) error {
	_, err := srv.TasklistsService().Insert(taskList).Do()
	srv.cacher.Delete(taskListsCacheKey)
	return err
}

// DeleteTaskList delete TaskList.
func (srv *Service) DeleteTaskList(id string) error {
	err := srv.TasklistsService().Delete(id).Do()
	srv.cacher.Delete(taskListsCacheKey)
	return err
}

// TasksService return TasksService.
func (srv *Service) TasksService() *tasks.TasksService {
	return srv.taskService.Tasks
}

// Tasks return Tasks.
func (srv *Service) Tasks(taskListID string) (*tasks.Tasks, error) {
	cacheKey := tasksCacheKey + "-" + taskListID

	data, err := srv.cacher.Read(cacheKey)
	if data != nil && err == nil {
		var tasks tasks.Tasks
		err = json.Unmarshal(data, &tasks)
		return &tasks, err
	}

	tasks, err := srv.TasksService().List(taskListID).MaxResults(50).Do()
	if err != nil {
		return nil, err
	}

	json, err := tasks.MarshalJSON()
	if err != nil {
		return nil, err
	}
	srv.cacher.Write(cacheKey, json, cacher.Forever)

	return tasks, nil
}

// InsertTask insert new Task.
func (srv *Service) InsertTask(taskListID string, task *tasks.Task) (*tasks.Task, error) {
	task, err := srv.TasksService().Insert(taskListID, task).Do()
	srv.cacher.Delete(tasksCacheKey + "-" + taskListID)
	return task, err
}

// DeleteTask delete Task.
func (srv *Service) DeleteTask(taskListID string, taskID string) error {
	err := srv.TasksService().Delete(taskListID, taskID).Do()
	srv.cacher.Delete(tasksCacheKey + "-" + taskListID)
	return err
}

// UpdateTask insert new Task.
func (srv *Service) UpdateTask(taskListID string, task *tasks.Task) (*tasks.Task, error) {
	task, err := srv.TasksService().Update(taskListID, task.Id, task).Do()
	srv.cacher.Delete(tasksCacheKey + "-" + taskListID)
	return task, err
}

// ClearTask clear Tasks.
func (srv *Service) ClearTask(taskListID string) error {
	err := srv.TasksService().Clear(taskListID).Do()
	srv.cacher.Delete(tasksCacheKey + "-" + taskListID)
	return err
}

func (srv *Service) getClient(config *oauth2.Config) (*http.Client, error) {
	tokFile, err := srv.tokenFile()
	if err != nil {
		return nil, err
	}

	tok, err := srv.tokenFromFile(tokFile)
	if err != nil {
		if tok, err = srv.getTokenFromWeb(config); err != nil {
			return nil, err
		}
		srv.saveToken(tokFile, tok)
	}

	return config.Client(context.Background(), tok), err
}

func (srv *Service) tokenFile() (string, error) {
	dir := configure.ConfigDir("gtodo")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", errors.Wrap(err, "Unable to open file")
	}

	return filepath.Join(dir, "token.json"), nil
}

func (srv *Service) getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, errors.Wrap(err, "Unable to read authorization code")
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to retrieve token from web")
	}
	return tok, nil
}

func (srv *Service) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func (srv *Service) saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrap(err, "Unable to cache oauth token")
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)

	return nil
}

func (srv *Service) credentialsPath() string {
	if path := os.Getenv("CREDENTIALS"); len(path) != 0 {
		return path
	}

	return filepath.Join(os.Getenv("HOME"), ".credentials.json")
}
