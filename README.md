# Gongfig
Import and export [Kong](https://getkong.org/) configuration tool written in Go

Current version supports only config export for and services and routes resources.

## Install
`go install "github.com/romanovskyj/gongfig"`

## Usage
`gongfig [global options] command [command options] [arguments...]`

#### Commands
```
export - Obtain services and routes, write it to the config file
help, h - Shows a list of commands or help for one command
```

#### Global options
```
--help, -h show help
--version, -v print the version
```

#### Example
```
gongfig export --url=http://localhost:8001 --file /tmp/config.json
```