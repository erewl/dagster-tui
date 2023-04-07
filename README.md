# DagsterTUI - Terminal UI for Dagster

## Configuration

In your root directory create a `~/.dagstertui` folder and in there you can create your `config.json` file

```
# config.json

{
    "environments": {
        "test": "https://your-url-to-your-dagster.environment",
        "acce": "https://another-url-to-your-dagster.environment",
    }
}
```

And then you can start the dagster-tui by specifying which environment you want to target: `/path/to/dagstertui -e test`


## Local Development
Currently using Go Version `1.16.7 darwin/amd64`

Build with:

```
go install
```

and running application with:

```
~/go/bin/dagsterui -e <dagster-environment>
```