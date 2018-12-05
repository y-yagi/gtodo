package main

import (
	"strings"
	"time"

	"github.com/0xAX/notificator"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/y-yagi/gtodo"
	tasks "google.golang.org/api/tasks/v1"
)

func addTask(c *cli.Context) error {
	srv, err := gtodo.NewService()
	if err != nil {
		return err
	}

	tasklist, err := selectTasklist(srv)
	if err != nil {
		return err
	}

	var task tasks.Task
	if err := buildTask(&task); err != nil {
		return err
	}

	_, err = srv.Tasks().Insert(tasklist.Id, &task).Do()
	if err != nil {
		return errors.Wrap(err, "Task insert failed")
	}

	return nil
}

func deleteTask(c *cli.Context) error {
	srv, err := gtodo.NewService()
	if err != nil {
		return err
	}

	tasklist, err := selectTasklist(srv)
	if err != nil {
		return err
	}

	task, err := selectTask(srv, tasklist.Id)
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

	if err = srv.Tasks().Delete(tasklist.Id, task.Id).Do(); err != nil {
		return errors.Wrap(err, "Task delete failed")
	}

	return nil
}

func updateTask(c *cli.Context) error {
	srv, err := gtodo.NewService()
	if err != nil {
		return err
	}

	tasklist, err := selectTasklist(srv)
	if err != nil {
		return err
	}

	task, err := selectTask(srv, tasklist.Id)
	if err != nil {
		return err
	}

	if err := buildTask(&task); err != nil {
		return err
	}

	_, err = srv.Tasks().Update(tasklist.Id, task.Id, &task).Do()
	if err != nil {
		return errors.Wrap(err, "Task update failed")
	}

	return nil
}

func completeTask(c *cli.Context) error {
	srv, err := gtodo.NewService()
	if err != nil {
		return err
	}

	tasklist, err := selectTasklist(srv)
	if err != nil {
		return err
	}

	task, err := selectTask(srv, tasklist.Id)
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
	if _, err = srv.Tasks().Update(tasklist.Id, task.Id, &task).Do(); err != nil {
		return errors.Wrap(err, "Task complete failed")
	}

	// NOTE: Clears all completed tasks
	if err = srv.Tasks().Clear(tasklist.Id).Do(); err != nil {
		return errors.Wrap(err, "Task clear failed")
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
		task = *taskSrv.Items[0]
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

func buildTask(task *tasks.Task) error {
	var err error

	validate := func(input string) error {
		if len(input) == 0 {
			return errors.New("Title can not be empty")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:     "Title",
		Default:   task.Title,
		Validate:  validate,
		AllowEdit: true,
	}

	task.Title, err = prompt.Run()
	if err != nil {
		return errors.Wrap(err, "Prompt canceled")
	}

	prompt.Label = "Due(yyyy-MM-dd)"
	if len(task.Due) != 0 {
		time, _ := time.Parse(time.RFC3339, task.Due)
		due := time.Format("2006-1-2")
		prompt.Default = due
	} else {
		prompt.Default = ""
	}

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
	prompt.Default = task.Notes
	prompt.Validate = func(input string) error {
		return nil
	}
	task.Notes, err = prompt.Run()
	if err != nil {
		return errors.Wrap(err, "Prompt failed")
	}

	return nil
}

func notifyTask(c *cli.Context) error {
	notify := notificator.New(notificator.Options{
		AppName: "gtodo",
	})

	srv, err := gtodo.NewService()
	if err != nil {
		return err
	}

	tList, err := srv.Tasklists().List().MaxResults(10).Do()
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve task lists")
	}

	if len(tList.Items) == 0 {
		return nil
	}

	for _, i := range tList.Items {
		var msg string

		tasks, err := srv.Tasks().List(i.Id).MaxResults(50).Do()
		if err != nil {
			return errors.Wrap(err, "Unable to retrieve tasks")
		}

		for _, task := range tasks.Items {
			if task.Title == "" {
				continue
			}

			msg += task.Title
			if task.Due != "" {
				time, _ := time.Parse(time.RFC3339, task.Due)
				msg += "(" + time.Format("2006-1-2") + ")"
			}
			msg += "\n"
		}

		msg = strings.TrimRight(msg, "\n")
		if len(msg) != 0 {
			notify.Push("gtodo", msg, "dialog-information", notificator.UR_CRITICAL)
		}
	}
	return nil
}
