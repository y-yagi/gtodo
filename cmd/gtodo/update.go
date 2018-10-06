package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/y-yagi/gtodo"
)

func cmdUpdate(c *cli.Context) error {
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

	if err := buildTask(&task); err != nil {
		return err
	}

	_, err = srv.Tasks().Update(taskListID, task.Id, &task).Do()
	if err != nil {
		return errors.Wrap(err, "Task update failed")
	}

	return nil
}
