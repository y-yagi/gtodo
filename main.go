package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/y-yagi/configure"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/tasks/v1"
)

func msg(err error) int {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		return 1
	}
	return 0
}

func tokenFile() (string, error) {
	dir := configure.ConfigDir("gtodo")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", errors.Wrap(err, "Unable to open file")
	}

	return filepath.Join(dir, "token.json"), nil
}

func getClient(config *oauth2.Config) (*http.Client, error) {
	tokFile, err := tokenFile()
	if err != nil {
		return nil, err
	}

	tok, err := tokenFromFile(tokFile)
	if err != nil {
		if tok, err = getTokenFromWeb(config); err != nil {
			return nil, err
		}
		saveToken(tokFile, tok)
	}

	return config.Client(context.Background(), tok), err
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
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

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrap(err, "Unable to cache oauth token")
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)

	return nil
}

func run() error {
	b, err := ioutil.ReadFile(filepath.Join(os.Getenv("HOME"), ".credentials.json"))
	if err != nil {
		return errors.Wrap(err, "Unable to read client secret file")
	}

	config, err := google.ConfigFromJSON(b, tasks.TasksScope)
	if err != nil {
		return errors.Wrap(err, "Unable to parse client secret file to config")
	}

	client, err := getClient(config)
	if err != nil {
		return errors.Wrap(err, "Unable to get Client")
	}

	srv, err := tasks.New(client)
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve tasks Client")
	}

	tList, err := srv.Tasklists.List().MaxResults(10).Do()
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve task lists")
	}

	if len(tList.Items) > 0 {
		for _, i := range tList.Items {
			tasks, err := srv.Tasks.List(i.Id).MaxResults(50).Do()
			if err != nil {
				return errors.Wrap(err, "Unable to retrieve tasks")
			}

			fmt.Printf("## %s\n", i.Title)

			for _, task := range tasks.Items {
				due := ""
				if task.Due != "" {
					time, _ := time.Parse(time.RFC3339, task.Due)
					due = "(" + time.Format("2006/1/2") + ")"
				}
				fmt.Printf("* %s %s\n", task.Title, due)
			}
			fmt.Printf("\n")
		}
	} else {
		fmt.Print("No task lists found.")
	}

	return nil
}

func main() {
	os.Exit(msg(run()))
}
