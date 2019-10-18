package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

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
			Name:   "notify",
			Usage:  "notify todos",
			Action: notifyTask,
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

	tList, err := srv.TaskLists()
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve task lists")
	}

	logger := log.New(os.Stdout, "", 0)
	var wg sync.WaitGroup

	if len(tList.Items) > 0 {
		for _, i := range tList.Items {
			wg.Add(1)

			go func(item *tasks.TaskList) {
				defer wg.Done()

				buf := &bytes.Buffer{}
				table := tablewriter.NewWriter(buf)
				table.SetAutoWrapText(false)
				table.SetHeader([]string{"Title", "Due", "Note"})
				table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
				table.SetCenterSeparator("|")

				tasks, err := srv.Tasks(item.Id)
				if err != nil {
					logger.Printf("Unable to retrieve tasks: %v\n", err)
					return
				}

				showHeader(buf, item.Title)

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
				logger.Print(buf.String() + "\n")
			}(i)
		}
	} else {
		logger.Print("No task lists found.")
	}
	wg.Wait()

	return nil
}

func showHeader(w io.Writer, header string) {
	fmt.Fprintf(w, "─────────────────────────────────────\n")
	fmt.Fprintf(w, "  %s\n", strings.TrimSpace(header))
	fmt.Fprintf(w, "─────────────────────────────────────\n")
}
