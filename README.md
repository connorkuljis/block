# Block-cli

**Block** is a simple, cross-platform command line interface to eliminate digital distractions.

The projects web site is [https://try-block.fly.dev/](https://try-block.fly.dev/)

Get the latest release at [https://github.com/connorkuljis/block/releases](https://github.com/connorkuljis/block/releases)

To checkout the code use `git clone https://github.com/connorkuljis/block-cli.git`
 
![demo](.github/demo.gif)

# Mission

Spend less time on the computer and more time in the sun.

# Building from source

`make`, then run `make install` to install it.

- note: ensure a Golang compiler is present on the machine.

# Documentation

## Block Sites (Guide)

1. Open your `/etc/hosts` file

`sudo vi /etc/hosts`

2. Paste an example blocklist

```
# --- social media
# 0.0.0.0 twitter.com
# 0.0.0.0 www.youtube.com
# 0.0.0.0 www.instagram.com
# 0.0.0.0 www.reddit.com
# 0.0.0.0 reddit.com
# 0.0.0.0 www.old.reddit.com
# 0.0.0.0 old.reddit.com
# 0.0.0.0 www.facebook.com
# ~ <-- important! lines below the (~) character mark the end of the blocklist
```

# Usage
- To see the list of commands available, run `block --help`

# Faq
# Troubleshooting Screen Recording with Ffmpeg
- run `ffmpeg -v` and ensure the installation is not corrupted or missing.
- ensure system permissions are enabled to record your screen.
- a valid input device is configured in `config.yaml`
- restart the terminal application

## Configuration

- open `.config/block-cli/config.yaml` 

Example:

```
# config.yaml
ffmpegRecordingsPath: /Volumes/WD_2TB/Screen-Recordings
avfoundationDevice: "1:0"

```
