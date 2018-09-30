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
					due = time.Format("2006/1/2")
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
