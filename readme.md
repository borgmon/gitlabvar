# gitlabvar

A CLI tool export or apply your gitlab variables.

## Build

```zsh
go build .
```

## Install

### With Go

```zsh
go get github.com/borgmon/gitlabvar
gitlabvar
```

if you cannot find the command, add your `~/go/bin` to your `PATH`

```zsh
export PATH="$PATH:$HOME/go/bin"
```

### Download binary

```zsh
./gitlabvar
```

## Usage

```zsh
NAME:
   gitlabvar - Export and import your CI variable from gitlab

USAGE:
   gitlabvar [global options] command [command options] [arguments...]

DESCRIPTION:
   Example: gitlabvar --token {gitlab-token} -project {projectID} get

COMMANDS:
   apply, a  apply variable yaml to gitlab project
   get, g    get variable yaml from gitlab project
   env, e    export .env file
   init, i   get a sample yaml file

GLOBAL OPTIONS:
   --import value, -i value   import yaml file path (default: ".gitlab-ci-var.yaml")
   --export value, -o value   export yaml file path (default: ".gitlab-ci-var.yaml")
   --token value, -t value    gitlab token. scope required: api. get it from here: https://gitlab.com/-/profile/personal_access_tokens
   --project value, -p value  Project ID, get it from frontpage of the project
   --help, -h                 show help (default: false)
```
