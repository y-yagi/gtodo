package main

import (
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/y-yagi/gtodo"
)

func cmdComplete(c *cli.Context) error {
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

	label := "Will complete " + "'" + task.Title + "'" + ". Are you sure"
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
	}

	if _, err = prompt.Run(); err != nil {
		// NOTE: confirm canceled
		return nil
	}

	task.Status = "completed"
	if _, err = srv.Tasks().Update(taskListID, task.Id, &task).Do(); err != nil {
		return errors.Wrap(err, "Task complete failed")
	}

	// NOTE: Clears all completed tasks
	if err = srv.Tasks().Clear(taskListID).Do(); err != nil {
		return errors.Wrap(err, "Task clear failed")
	}
	return nil
}
