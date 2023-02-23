package common

var ConfigUsage string = `

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
aforementioned directory. `

type Config struct {
	Token string `json:"token"`
	Path  string `json:"path"`
}
