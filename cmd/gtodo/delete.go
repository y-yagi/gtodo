package main

import (
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/y-yagi/gtodo"
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
