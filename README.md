# block-cli

**block** reduces distractions from the command line. 

Features:
- 🙆 Pomodoro-like progress bar inidicator (right in your terminal!). 
  - 🙅Automatically block/unblock any site at the IP level during the duration of the program.
    - 📬 Alerted by a system notification when a session ends.
- ⛳ Automatically record your progress.
  - 📒 Answer 'what did I get done today' by running `block history`.
- 󰑊 Capture your progress by enabling the screen recorder with `-x` or `--screen-recorder`.
  - 🎥 Compile recordings into a time-lapse.


```
❯ block start 10 'draft emails' -x
Setting a timer for 10.0 minutes.
ESC or 'q' to exit. Press any key to pause.
 100% |███████████████████████████████████████████████████████████████████████████████████████████████| (90/90, 15 it/s) [5s]
```

# Mission

Spend less time on the computer and more time in the sun.

---

> One must be concious of the time he spends on the internet.
It should be an event. Coming onto the net. Endless information and oppportunity.
An event treated with respend and dowe with intention in mind. Very much like reading a book.
Careful and focused on the goal. Read the pdf, write the text, send the message and fuck off of there.
Back to the real world. Sitting at a cafe, chatting with people in your viciniy and reading a newspaper.
Or perhaps sun bathing. Who knows. You get my idea. Lindy.

---

# Usage
```
Usage:
  block [flags]
  block [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  delete      Deletes a task by given ID.
  help        Help about any command
  history     Show task history.
  start

Flags:
  -h, --help              help for block
  -d, --no-block          Do not block hosts file.
  -x, --screen-recorder   Enable screen recorder.
  -v, --verbose           Logs additional details.

Use "block [command] --help" for more information about a command.
```

# Install

> To install the program, please read the instructions below:

** important! **
- `linux / mac`
- requires `go`, *you can install go here: [go.dev](https://go.dev/)*
- requires `ffmpeg` 
- **screencapture** is only supported on `macOS`.


## Download
`git clone https://github.com/connorkuljis/block-cli.git && cd block-cli`

## Build
`make`

## Run
`make run`

### Move binary to your path (optional)
`make release`

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

```

- The program will uncomment the lines when you start the program, and add them back in when upon exit.
 - If you have content you dont want the program to manipulate, add the following line to the hosts file.

 `# ~ <-- lines below this will not be uncommented/commented by block-cli`

 in which the `~` acts as a delimiter.

# Screen Recording with Ffmpeg

If you have `ffmpeg` on you machine you can automatically capture your screen. It will be saved to `~/Downloads`.

To record your screen use the `-x` flag.

# Improvements:
1. `resetDNS` only flushes the DNS cache on macos.
2. Implement `.config` file for user settings (eg: screen capture directory, default task length...)


## Optional

Editing the /etc/hosts file typically requires administrative privileges, and for security reasons, it's not recommended to completely eliminate the password prompt when using sudo to modify system files. However, you can configure sudo to not prompt for a password for specific commands, as discussed earlier.

If you want to allow the mv command on the /etc/hosts file without entering a password,  with a specific configuration for the mv command. Open the sudoers file using visudo:

`sudo visudo`

Add a line at the end of the file to allow running the mv command on the /etc/hosts file without entering a password:

`your_username ALL=(ALL) NOPASSWD: /bin/mv /etc/hosts`

Replace your_username with your actual username.

Save and exit the editor.

Now, when you use the mv command to move the /etc/hosts file, you won't be prompted for a password.

