# DagsterTUI - Terminal UI for Dagster

The TUI wrapper for your dagster environments. Work with dagster from the comfort of your terminal

## Installation

- go to the releases tab and download the file package that matches the architecture of your machine

## Configuration

In your root directory create a `~/.dagstertui` folder and in there you can create your `config.json` file

```
# config.json

{
    "environments": {
        "default": "test", # fallback value for when no -e argument is given
        "test": "https://your-url-to-your-dagster.environment",
        "acce": "https://another-url-to-your-dagster.environment",
    }
}
```

And then you can start the dagster-tui by specifying which environment you want to target: `/path/to/dagstertui -e test`


**Pressing 'x' will open up the the different Keybindings to navigate through the TUI**

## Local Development
Currently using Go Version `1.19.1 darwin/amd64`

Install from source with:

```
go install cmd/dagstertui.go
```

and running application with:

```
~/go/bin/dagsterui -e <dagster-environment>
```

Run tests with:
```
go test ./test/
```
