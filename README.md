# block-cli

**block** reduces distractions from the command line. 

The program immediately blocks and unblocks websites when a task is started or completes.


```
❯ ./block 10 --task "draft email"
Setting a timer for 10.0 minutes.
ESC or 'q' to exit. Press any key to pause.
Blocker:        started
 100% |███████████████████████████████████████████████████████████████████████████████████████████████| (90/90, 15 it/s) [5s]
Blocker:        stopped
Start time:     7:25:44am
End time:       7:35:51am
Duration:       0 hours, 10 minutes and 0 seconds.
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
Block saves you time by blocking websites at IP level.
Progress bar is displayed directly in the terminal. 
Automatically unblock sites when the task is complete.

Usage:
  block [flags]

Flags:
  -h, --help              help for block
      --no-block          Disables internet blocker.
  -x, --screen-recorder   Enables screen recorder.
  -t, --task string       Record optional task name.
  -v, --verbose           Logs additional details.
```

# Install

*note: requires `go`, you can install go here: [go.dev](https://go.dev/)*

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

