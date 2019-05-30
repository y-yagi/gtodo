package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/y-yagi/gocui"
	"github.com/y-yagi/gtodo"
	tasks "google.golang.org/api/tasks/v1"
)

var tasksPerList = map[string][]*tasks.Task{}

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

	tList, err := srv.Tasklists().List().MaxResults(10).Do()
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve task lists")
	}

	logger := log.New(os.Stdout, "", 0)
	var wg sync.WaitGroup

	if len(tList.Items) > 0 {
		for _, item := range tList.Items {
			wg.Add(1)

			go func(item *tasks.TaskList) {
				defer wg.Done()
				tasks, err := srv.Tasks().List(item.Id).MaxResults(50).Do()

				if err != nil {
					logger.Printf("Unable to retrieve tasks: %v\n", err)
				}

				tasksPerList[item.Title] = tasks.Items
			}(item)
		}
	} else {
		return errors.Wrap(err, "No task lists found")
	}
	wg.Wait()

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return errors.Wrap(err, "gui create error")
	}
	defer g.Close()

	g.Cursor = true
	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		return errors.Wrap(err, "Key bindings error")
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		return errors.Wrap(err, "Unexpected error")
	}

	return nil
}
