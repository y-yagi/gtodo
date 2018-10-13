# gtodo

`gotod` is a simple CLI tool for Google ToDo.

[![asciicast](https://asciinema.org/a/1PYigUpJoy1kW734r2EMOPTfE.png)](https://asciinema.org/a/1PYigUpJoy1kW734r2EMOPTfE)


## Installation

Use `go get` to install this package:

```bash
$ go get github.com/y-yagi/gtodo/cmd/gtodo
```

## Usage

### Setup

Need credentials file to use Google API. Please refer to [Tasks API](https://developers.google.com/tasks/quickstart/go) and downloads credentials file for Google Tasks API.

The credentials file path can specify via `CREDENTIALS` env. If not specified `CREDENTIALS` env, `gtodo` try to read `.credentials.json` under the home directory.

### Help

```
$ gtodo help
NAME:
   gtodo - CLI for Google ToDo

USAGE:
   gtodo [global options] command [command options] [arguments...]

VERSION:
   0.1.0

COMMANDS:
     add, a       add a new todo
     complete, c  complete a todo
     delete, d    delete a todo
     update, u    update a todo
     tasklist     action for tasklist
     help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version

$ gtodo tasklist help
NAME:
   gtodo tasklist - action for tasklist

USAGE:
   gtodo tasklist [global options] command [command options] [arguments...]

VERSION:
   0.1.0

COMMANDS:
     add     add a new tasklist
     delete  delete a tasklist

GLOBAL OPTIONS:
   --help, -h  show help
```

