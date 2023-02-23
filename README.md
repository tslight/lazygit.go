 ![CI Result](https://github.com/tslight/lazygit.go/actions/workflows/build.yml/badge.svg?event=push) [![Go Report Card](https://goreportcard.com/badge/github.com/tslight/lazygit.go)](https://goreportcard.com/report/github.com/tslight/lazygit.go) [![Go Reference](https://pkg.go.dev/badge/github.com/tslight/lazygit.go.svg)](https://pkg.go.dev/github.com/tslight/lazygit.go)
# GitHub & GitLab API Clients

*Clone or pull all your projects/repos in one fell swoop.*

*For GitLab you can limit this to only certain groups.*

## Installation

``` shell
go install github.com/tslight/lazygit.go/cmd/gitlab@latest
go install github.com/tslight/lazygit.go/cmd/github@latest
```

Alternatively, download a suitable pre-compiled binary for your architecture
and operating system from the
[Releases](https://github.com/tslight/lazygit.go/releases) page and move it to
somewhere in your `$PATH`.

## GitHub CLI Usage

``` text
Usage: github

With no arguments will clone or pull all projects that can be accessed with
your API token to a specified directory.

The API token & directory belong in a JSON configuration file, which will live
in one of the following locations depending on your OS:

macOS:     $HOME/Library/Application Support/lazygit
Linux/BSD: $XDG_CONFIG_HOME/lazygit (usually $HOME/.config)
Windows:   %APPDATA%\lazygit (usually C:\Users\%USER%\AppData\Roaming)
Fallback:  $HOME/.config

If a JSON configuration file doesn't exist you will be prompted to enter an API
token and a directory. Those choices will be saved to a JSON file the
aforementioned directory.

  -v    print version info
```

## GitLab CLI Usage

``` text
Usage: gitlab [GROUP...]

With no arguments will clone or pull all projects that can be accessed with
your API token to a specified directory.

The API token & directory belong in a JSON configuration file, which will live
in one of the following locations depending on your OS:

macOS:     $HOME/Library/Application Support/lazygit
Linux/BSD: $XDG_CONFIG_HOME/lazygit (usually $HOME/.config)
Windows:   %APPDATA%\lazygit (usually C:\Users\%USER%\AppData\Roaming)
Fallback:  $HOME/.config

If a JSON configuration file doesn't exist you will be prompted to enter an API
token and a directory. Those choices will be saved to a JSON file the
aforementioned directory.

Optional [GROUP...] arguments will only clone or pull the projects found in
those groups.

  -v    print version info
```
