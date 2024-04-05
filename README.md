# block-cli

**block** eliminates distractions and saves you time from the command line. 

![demo](.github/demo.gif)

# Mission

Spend less time on the computer and more time in the sun.

# Install

## Download latest binary release
> To install the program, please read the instructions below:

- [downloaded latest binary release tarball](https://github.com/connorkuljis/block/releases)
- extract the binary from the tarball
- add the binary to your system $PATH


## Build from source
** important! **
- requires `go`, *you can install go here: [go.dev](https://go.dev/)*
- requires `ffmpeg` 

`git clone https://github.com/connorkuljis/block-cli.git && cd block-cli`

`make`

`make install`

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
# ~ <-- lines below this character will be untouched by block-cli

# Added by Docker Desktop
192.168.0.13    host.docker.internal
192.168.0.13    gateway.docker.internal
# To allow the same kube context to work on the host and the container:
127.0.0.1   kubernetes.docker.internal
# End of section

```

- The program will uncomment the lines when you start the program, and add them back in when upon exit.
- If you have content you dont want the program to manipulate, add the following line to the hosts file.


Features:
- 🙆 Pomodoro-like progress bar inidicator (right in your terminal!). 
  - 🙅Automatically block/unblock any site at the IP level during the duration of the program.
    - 📬 Alerted by a system notification when a session ends.
- ⛳ Automatically record your progress.
  - 📒 Answer 'what did I get done today' by running `block history`.
- 󰑊 Capture your progress by enabling the screen recorder with `-x` or `--screen-recorder`.
  - 🎥 Compile recordings into a time-lapse.
- Cross-platform integration for `mac` and `linux` (`windows` soon).
  - If you are having issues -> [https://github.com/connorkuljis/block-cli/issues](https://github.com/connorkuljis/block-cli/issues)
- `YAML` file at `~/.config/block-cli/config.yaml`


```
❯ block start 10 'draft emails' -x
Setting a timer for 10.0 minutes.
ESC or 'q' to exit. Press any key to pause.
 100% |███████████████████████████████████████████████████████████████████████████████████████████████| (90/90, 15 it/s) [5s]
```

# Usage
- To see the list of commands available, run `block --help`


# Faq
# Troubleshooting Screen Recording with Ffmpeg
If you have `ffmpeg` on you machine you can automatically capture your screen. 

To record your screen use the `-c` flag.

If you are having issues recording your screen, follow the checklist:

- ensure system permissions are enabled to record your screen.
- run `$ ffmpeg -v` to ensure your installation is not corrupted or missing.
- a valid input device is configured in `config.yaml`
- you have restarted your application 


## Configuration

- open `.config/block-cli/config.yaml` 

Example:

```
# config.yaml
ffmpegRecordingsPath: /Volumes/WD_2TB/Screen-Recordings
avfoundationDevice: "1:0"

```
