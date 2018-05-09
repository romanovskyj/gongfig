# Gongfig

[![Build Status](https://travis-ci.org/romanovskyj/gongfig.svg?branch=master)](https://travis-ci.org/romanovskyj/gongfig)

Import and export [Kong](https://getkong.org/) configuration tool written in Go

Current version supports only config export for services and routes resources.

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

## Docker
As usually Kong admin api is not reachable externally, you can deploy docker container with gongfig and use it as sidecar application.
The image name is `eromanovskyj/gongfig`