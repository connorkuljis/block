package blocker

import (
	"bufio"
	"bytes"
	"os"
)

const (
	StopToken = '~'
	HostsFile = "/etc/hosts"
)

type Blocker struct {
	HostsFile string
	IsEnabled bool
}

func NewBlocker() Blocker {
	return Blocker{
		HostsFile: HostsFile,
	}
}

// enables the blocker
func (b *Blocker) Enable() error {
	b.IsEnabled = true
	err := updateBlockList(b)
	if err != nil {
		return err
	}
	return nil
}

// disables the blocker
func (b *Blocker) Disable() error {
	b.IsEnabled = false
	err := updateBlockList(b)
	if err != nil {
		return err
	}
	return nil
}

func prependComment(line []byte) []byte {
	// true if '#' byte exists in slice
	isComment := bytes.IndexByte(line, '#') == 0

	if isComment {
		return line
	}
	return append([]byte("# "), line...)
}

func removeComment(line []byte) []byte {
	isComment := bytes.IndexByte(line, '#') == 0

	if isComment {
		return bytes.TrimSpace(line[1:])
	}
	return line
}

func updateBlockList(b *Blocker) error {
	// open the special hosts file, (requires root password)
	file, err := os.Open(b.HostsFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// scan the file
	sc := bufio.NewScanner(file)

	var data []byte
	found := false
	// iterate over the line until stop token found
	for sc.Scan() && !found {
		line := sc.Bytes()

		foundStopToken := bytes.IndexByte(line, StopToken) >= 0

		if foundStopToken {
			found = true
		} else {
			// parse the line depending if blocker is enabled.
			if b.IsEnabled {
				line = removeComment(line)
			} else {
				line = prependComment(line)
			}
		}

		line = append(line, '\n')
		data = append(data, line...)
	}

	// scan remainder of file
	for sc.Scan() {
		line := sc.Bytes()
		line = append(line, '\n')
		data = append(data, line...)
	}

	if err = sc.Err(); err != nil {
		return err
	}

	err = overwriteFile(b.HostsFile, data)
	if err != nil {
		return err
	}

	return nil
}

// takes a filename and overwrites it with data
func overwriteFile(filename string, data []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}
