package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/y-yagi/gtodo"
	tasks "google.golang.org/api/tasks/v1"
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	app := cli.NewApp()
	app.Name = "gtodo"
	app.Usage = "CLI for Google ToDo"
	app.Version = "0.1.0"
	app.Action = appRun
	app.Commands = commands()

	return msg(app.Run(args))
}

func commands() []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "add a new todo",
			Action:  cmdAdd,
		},
		cli.Command{
			Name:    "delete",
			Aliases: []string{"d"},
			Usage:   "delete a todo",
			Action:  cmdDelete,
		},
		cli.Command{
			Name:    "update",
			Aliases: []string{"u"},
			Usage:   "update a todo",
			Action:  cmdUpdate,
		},
	}
}

func msg(err error) int {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		return 1
	}
	return 0
}

func appRun(c *cli.Context) error {
	srv, err := gtodo.NewService()
	if err != nil {
		return err
	}

	tList, err := srv.Tasklists().List().MaxResults(10).Do()
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve task lists")
	}

	if len(tList.Items) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Title", "Due", "Note"})
		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")

		for _, i := range tList.Items {
			tasks, err := srv.Tasks().List(i.Id).MaxResults(50).Do()
			if err != nil {
				return errors.Wrap(err, "Unable to retrieve tasks")
			}

			showHeader(os.Stdout, i.Title)

			for _, task := range tasks.Items {
				var data []string
				var due string

				if task.Title == "" {
					continue
				}

				data = append(data, task.Title)
				if task.Due != "" {
					time, _ := time.Parse(time.RFC3339, task.Due)
					due = time.Format("2006-1-2")
				}
				data = append(data, due)
				data = append(data, task.Notes)
				table.Append(data)
			}
			table.Render()
			table.ClearRows()
			fmt.Fprintf(os.Stdout, "\n")
		}
	} else {
		fmt.Print("No task lists found.")
	}

	return nil
}

func showHeader(w io.Writer, header string) {
	fmt.Fprintf(w, "─────────────────────────────────────\n")
	fmt.Fprintf(w, "  %s\n", strings.TrimSpace(header))
	fmt.Fprintf(w, "─────────────────────────────────────\n")
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
