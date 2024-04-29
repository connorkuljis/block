package blocker

import (
	"bufio"
	"bytes"
	"log/slog"
	"os"
)

const (
	StopToken = '~'
)

type Blocker struct {
	hostsFile string
}

func NewBlocker() Blocker {
	hostsFile := "/etc/hosts"

	return Blocker{
		hostsFile: hostsFile,
	}
}

func (b *Blocker) Start() (int, error) {
	shouldBlock := true
	n, err := updateBlockList(b.hostsFile, shouldBlock)
	if err != nil {
		return n, err
	}
	return n, nil
}

func (b *Blocker) Stop() (int, error) {
	shouldBlock := false
	n, err := updateBlockList(b.hostsFile, shouldBlock)
	if err != nil {
		return n, err
	}
	return n, nil
}

func addComment(line []byte) []byte {
	// true if '#' byte exists in slice
	isComment := bytes.IndexByte(line, '#') == 0

	if isComment {
		return line
	}
	return append([]byte("# "), line...)
}

func stripComment(line []byte) []byte {
	isComment := bytes.IndexByte(line, '#') == 0

	if isComment {
		return bytes.TrimSpace(line[1:])
	}
	return line
}

func updateBlockList(target string, shouldBlock bool) (int, error) {
	// open the special hosts file, (requires root password)
	var n int
	file, err := os.Open(target)
	if err != nil {
		return n, err
	}
	defer file.Close()

	sc := bufio.NewScanner(file)

	var data []byte
	found := false
	// iterate over the line until stop token found
	for sc.Scan() {
		line := sc.Bytes()

		lineHasTok := bytes.IndexByte(line, StopToken) >= 0

		if lineHasTok {
			found = true
		}

		if !found {
			if shouldBlock {
				line = stripComment(line)
			} else {
				line = addComment(line)
			}
		}

		line = append(line, '\n')
		data = append(data, line...)
	}

	if err = sc.Err(); err != nil {
		return n, err
	}

	slog.Debug(string(data))

	n, err = overwriteFile(target, data)
	if err != nil {
		return n, err
	}

	return n, nil
}

// takes a filename and overwrites it with data
func overwriteFile(filename string, data []byte) (int, error) {
	var n int
	file, err := os.Create(filename)
	if err != nil {
		return n, err
	}
	defer file.Close()

	n, err = file.Write(data)
	if err != nil {
		return n, err
	}
	return n, nil
}
