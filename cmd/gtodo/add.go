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

	prompt.Label = "Notes"
	prompt.Validate = func(input string) error {
		return nil
	}
	task.Notes, err = prompt.Run()
	if err != nil {
		return errors.Wrap(err, "Prompt failed")
	}

	return nil
}
