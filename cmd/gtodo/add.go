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
	gtSrv, err := gtodo.NewGTodoService()
	if err != nil {
		return err
	}

	taskListId, err := selectTaskList(gtSrv)
	if err != nil {
		return err
	}

	var task tasks.Task
	if err := buildTask(&task); err != nil {
		return err
	}

	_, err = gtSrv.Tasks().Insert(taskListId, &task).Do()
	if err != nil {
		return errors.Wrap(err, "Task insert failed")
	}

	return nil
}

func selectTaskList(gtSrv *gtodo.GTodoService) (string, error) {
	var taskListId string

	tList, err := gtSrv.Tasklists().List().MaxResults(10).Do()
	if err != nil {
		return "", errors.Wrap(err, "Unable to retrieve task lists")
	}

	if len(tList.Items) == 0 {
		return "", errors.New("No task lists found")
	}

	if len(tList.Items) == 1 {
		taskListId = tList.Items[0].Id
	} else {
		var selectItems []string
		// TODO: Add care about the same task list name
		titleListWithId := map[string]string{}

		for _, i := range tList.Items {
			selectItems = append(selectItems, i.Title)
			titleListWithId[i.Title] = i.Id
		}

		pSelect := promptui.Select{
			Label: "Select Task List",
			Items: selectItems,
		}
		_, result, err := pSelect.Run()

		if err != nil {
			return "", errors.Wrap(err, "Prompt canceled")
		}
		taskListId = titleListWithId[result]
	}

	return taskListId, nil
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
