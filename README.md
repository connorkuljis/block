# Task Tracker CLI

## Optional

Editing the /etc/hosts file typically requires administrative privileges, and for security reasons, it's not recommended to completely eliminate the password prompt when using sudo to modify system files. However, you can configure sudo to not prompt for a password for specific commands, as discussed earlier.

If you want to allow the mv command on the /etc/hosts file without entering a password,  with a specific configuration for the mv command. Open the sudoers file using visudo:

`sudo visudo`

Add a line at the end of the file to allow running the mv command on the /etc/hosts file without entering a password:

`your_username ALL=(ALL) NOPASSWD: /bin/mv /etc/hosts`

Replace your_username with your actual username.

Save and exit the editor.

Now, when you use the mv command to move the /etc/hosts file, you won't be prompted for a password.

