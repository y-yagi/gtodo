package main

import (
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/y-yagi/gtodo"
	tasks "google.golang.org/api/tasks/v1"
)

func addTasklist(c *cli.Context) error {
	srv, err := gtodo.NewService()
	if err != nil {
		return err
	}

	var tasklist tasks.TaskList
	if err := buildTasklist(&tasklist); err != nil {
		return err
	}

	_, err = srv.Tasklists().Insert(&tasklist).Do()
	if err != nil {
		return errors.Wrap(err, "Tasklist insert failed")
	}
	return nil
}

func deleteTasklist(c *cli.Context) error {
	srv, err := gtodo.NewService()
	if err != nil {
		return err
	}

	tasklist, err := selectTasklist(srv)
	if err != nil {
		return err
	}

	label := "Will delete " + "'" + tasklist.Title + "'" + ". Are you sure"
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
	}

	if _, err = prompt.Run(); err != nil {
		// NOTE: confirm canceled
		return nil
	}

	if err = srv.Tasklists().Delete(tasklist.Id).Do(); err != nil {
		return errors.Wrap(err, "Tasklist delete failed")
	}

	return nil
}

func buildTasklist(tasklist *tasks.TaskList) error {
	var err error

	validate := func(input string) error {
		if len(input) == 0 {
			return errors.New("Title can not be empty")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:     "Title",
		Default:   tasklist.Title,
		Validate:  validate,
		AllowEdit: true,
	}

	tasklist.Title, err = prompt.Run()
	if err != nil {
		return errors.Wrap(err, "Prompt canceled")
	}
	return nil
}

func selectTasklist(srv *gtodo.Service) (tasks.TaskList, error) {
	var tasklist tasks.TaskList

	tList, err := srv.Tasklists().List().MaxResults(10).Do()
	if err != nil {
		return tasklist, errors.Wrap(err, "Unable to retrieve task lists")
	}

	if len(tList.Items) == 0 {
		return tasklist, errors.New("No task lists found")
	}

	if len(tList.Items) == 1 {
		tasklist = *tList.Items[0]
	} else {
		var selectItems []string
		titleList := map[string]*tasks.TaskList{}

		for _, i := range tList.Items {
			selectItems = append(selectItems, i.Title)
			titleList[i.Title] = i
		}

		pSelect := promptui.Select{
			Label: "Select Task List",
			Items: selectItems,
		}
		_, result, err := pSelect.Run()

		if err != nil {
			return tasklist, errors.Wrap(err, "Prompt canceled")
		}
		tasklist = *titleList[result]
	}

	return tasklist, nil
}
