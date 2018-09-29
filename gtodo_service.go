package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/y-yagi/configure"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	tasks "google.golang.org/api/tasks/v1"
)

// GTodoService is a todo module.
type GTodoService struct {
	Service *tasks.Service
}

func NewGTodoService() (*GTodoService, error) {
	gt := &GTodoService{}

	if err := gt.buildTaskService(); err != nil {
		return nil, err
	}

	return gt, nil
}

func (gt *GTodoService) buildTaskService() error {
	b, err := ioutil.ReadFile(filepath.Join(os.Getenv("HOME"), ".credentials.json"))
	if err != nil {
		return errors.Wrap(err, "Unable to read client secret file")
	}

	config, err := google.ConfigFromJSON(b, tasks.TasksScope)
	if err != nil {
		return errors.Wrap(err, "Unable to parse client secret file to config")
	}

	client, err := gt.getClient(config)
	if err != nil {
		return errors.Wrap(err, "Unable to get Client")
	}

	gt.Service, err = tasks.New(client)
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve tasks Client")
	}

	return nil
}

func (gt *GTodoService) Tasklists() *tasks.TasklistsService {
	return gt.Service.Tasklists
}

func (gt *GTodoService) Tasks() *tasks.TasksService {
	return gt.Service.Tasks
}

func (gt *GTodoService) getClient(config *oauth2.Config) (*http.Client, error) {
	tokFile, err := gt.tokenFile()
	if err != nil {
		return nil, err
	}

	tok, err := gt.tokenFromFile(tokFile)
	if err != nil {
		if tok, err = gt.getTokenFromWeb(config); err != nil {
			return nil, err
		}
		gt.saveToken(tokFile, tok)
	}

	return config.Client(context.Background(), tok), err
}

func (gt *GTodoService) tokenFile() (string, error) {
	dir := configure.ConfigDir("gtodo")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", errors.Wrap(err, "Unable to open file")
	}

	return filepath.Join(dir, "token.json"), nil
}

func (gt *GTodoService) getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
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

func (gt *GTodoService) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func (gt *GTodoService) saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrap(err, "Unable to cache oauth token")
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)

	return nil
}
