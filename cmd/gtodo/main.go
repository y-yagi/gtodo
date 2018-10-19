package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

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
			Action:  addTask,
		},
		cli.Command{
			Name:    "complete",
			Aliases: []string{"c"},
			Usage:   "complete a todo",
			Action:  completeTask,
		},
		cli.Command{
			Name:    "delete",
			Aliases: []string{"d"},
			Usage:   "delete a todo",
			Action:  deleteTask,
		},
		cli.Command{
			Name:    "update",
			Aliases: []string{"u"},
			Usage:   "update a todo",
			Action:  updateTask,
		},
		cli.Command{
			Name:  "tasklist",
			Usage: "action for tasklist",
			Subcommands: []cli.Command{
				{
					Name:   "add",
					Usage:  "add a new tasklist",
					Action: addTasklist,
				},
				{
					Name:   "delete",
					Usage:  "delete a tasklist",
					Action: deleteTasklist,
				},
			},
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
	if c.NArg() != 0 {
		cli.ShowAppHelp(c)
		return nil
	}

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
