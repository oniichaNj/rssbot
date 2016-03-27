# rssbot
IRC bot who throws RSS stuff into channels

## Installation

First, install `go` using your package manager.

Secondly, you need to set a `$GOPATH`. I suggest using `~/.go`. Do this by adding

```
set -x GOPATH $HOME/.go/
set -gx PATH  $HOME/.go/bin/ $PATH
```
to your `~/.config/config.fish` or equivilant to your shell rc file.

This also adds the ~/.go/bin/ folder to your $PATH which is useful for some.

After that, the package is installable using the `go get github.com/oniichaNj/rssbot` command. This will download the sources into the workspace at `~/.go/src/github.com/oniichaNj/rssbot` directory.

In this directory, `go build` will create a binary called `rssbot`. This is the executable for the bot.

If you wish to add this to your $PATH, run `go install`, which moves it into `~/.go/bin/`.

Bare in mind rssbot uses a relative path for config.json, just move a copy into wherever. 
