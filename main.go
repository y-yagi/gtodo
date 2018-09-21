package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/y-yagi/configure"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/tasks/v1"
)

func tokenFile() string {
	dir := configure.ConfigDir("gtodo")
	if err := os.MkdirAll(dir, 0700); err != nil {
		log.Fatalf("Unable to create config directory: %v", err)
	}

	return filepath.Join(dir, "token.json")
}

func getClient(config *oauth2.Config) *http.Client {
	tokFile := tokenFile()
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
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

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	b, err := ioutil.ReadFile(filepath.Join(os.Getenv("HOME"), ".credentials.json"))
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, tasks.TasksScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := tasks.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve tasks Client %v", err)
	}

	tList, err := srv.Tasklists.List().MaxResults(10).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve task lists. %v", err)
	}

	if len(tList.Items) > 0 {
		for _, i := range tList.Items {
			fmt.Printf("## %s\n", i.Title)
			tasks, err := srv.Tasks.List(i.Id).MaxResults(50).Do()
			if err != nil {
				log.Fatalf("Unable to retrieve tasks. %v", err)
			}
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
}
