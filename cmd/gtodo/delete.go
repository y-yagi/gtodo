package main

import (
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/y-yagi/gtodo"
	tasks "google.golang.org/api/tasks/v1"
)

func cmdDelete(c *cli.Context) error {
	srv, err := gtodo.NewService()
	if err != nil {
		return err
	}

	taskListID, err := selectTaskList(srv)
	if err != nil {
		return err
	}

	task, err := selectTask(srv, taskListID)
	if err != nil {
		return err
	}

	label := "Will delete " + "'" + task.Title + "'" + ". Are you sure"
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
	}

	if _, err = prompt.Run(); err != nil {
		// NOTE: confirm canceled
		return nil
	}

	if err = srv.Tasks().Delete(taskListID, task.Id).Do(); err != nil {
		return errors.Wrap(err, "Task delete failed")
	}

	return nil
}

func selectTask(srv *gtodo.Service, taskListID string) (tasks.Task, error) {
	var task tasks.Task

	taskSrv, err := srv.Tasks().List(taskListID).MaxResults(50).Do()
	if err != nil {
		return task, errors.Wrap(err, "Unable to retrieve tasks")
	}

	if len(taskSrv.Items) == 0 {
		return task, errors.New("No tasks found")
	}

	if len(taskSrv.Items) == 1 {
		task.Id = taskSrv.Items[0].Id
	} else {
		var selectItems []string
		titleListWithID := map[string]*tasks.Task{}

		for _, t := range taskSrv.Items {
			if t.Title == "" {
				continue
			}
			selectItems = append(selectItems, t.Title)
			titleListWithID[t.Title] = t
		}

		pSelect := promptui.Select{
			Label: "Select Task",
			Items: selectItems,
		}

		_, result, err := pSelect.Run()
		if err != nil {
			return task, errors.Wrap(err, "Prompt canceled")
		}
		task = *titleListWithID[result]
	}

	return task, nil
}
