package main

import (
	"time"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/y-yagi/gtodo"
	tasks "google.golang.org/api/tasks/v1"
)

func cmdAdd(c *cli.Context) error {
	srv, err := gtodo.NewService()
	if err != nil {
		return err
	}

	taskListID, err := selectTaskList(srv)
	if err != nil {
		return err
	}

	var task tasks.Task
	if err := buildTask(&task); err != nil {
		return err
	}

	_, err = srv.Tasks().Insert(taskListID, &task).Do()
	if err != nil {
		return errors.Wrap(err, "Task insert failed")
	}

	return nil
}

func selectTaskList(srv *gtodo.Service) (string, error) {
	var taskListID string

	tList, err := srv.Tasklists().List().MaxResults(10).Do()
	if err != nil {
		return "", errors.Wrap(err, "Unable to retrieve task lists")
	}

	if len(tList.Items) == 0 {
		return "", errors.New("No task lists found")
	}

	if len(tList.Items) == 1 {
		taskListID = tList.Items[0].Id
	} else {
		var selectItems []string
		// TODO: Add care about the same task list name
		titleListWithID := map[string]string{}

		for _, i := range tList.Items {
			selectItems = append(selectItems, i.Title)
			titleListWithID[i.Title] = i.Id
		}

		pSelect := promptui.Select{
			Label: "Select Task List",
			Items: selectItems,
		}
		_, result, err := pSelect.Run()

		if err != nil {
			return "", errors.Wrap(err, "Prompt canceled")
		}
		taskListID = titleListWithID[result]
	}

	return taskListID, nil
}

func buildTask(task *tasks.Task) error {
	var err error

	validate := func(input string) error {
		if len(input) == 0 {
			return errors.New("Title can not be empty")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Title",
		Validate: validate,
	}

	task.Title, err = prompt.Run()
	if err != nil {
		return errors.Wrap(err, "Prompt canceled")
	}

	prompt.Label = "Due(yyyy-MM-dd)"
	prompt.Validate = func(input string) error {
		if len(input) == 0 {
			return nil
		}

		_, err := time.Parse("2006-01-02", input)
		if err != nil {
			return errors.New("Invalid format")
		}

		return nil
	}

	due, err := prompt.Run()
	if err != nil {
		return errors.Wrap(err, "Prompt failed")
	}

	if len(due) != 0 {
		t, _ := time.Parse("2006-01-02", due)
		task.Due = t.Format(time.RFC3339)
	}

	return nil
}
