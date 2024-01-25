package blocker

import (
	"bufio"
	"os"
	"strings"
)

const (
	StopToken = "~"
	HostsFile = "/etc/hosts"
)

type Blocker struct {
	HostsFile string
	Disable   bool
}

func NewBlocker(disable bool) Blocker {
	return Blocker{
		HostsFile: HostsFile,
		Disable:   disable,
	}
}

func (b *Blocker) Unblock() error {
	if !b.Disable {
		unblockFn := func(line string) string {
			if string(line[0]) == "#" {
				return line
			}
			return "# " + line
		}
		err := b.UpdateBlockList(unblockFn)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Blocker) BlockAndReset() error {
	if !b.Disable {
		blockFn := func(line string) string {
			if string(line[0]) == "#" {
				return strings.TrimSpace(line[1:])
			}
			return line
		}
		err := b.UpdateBlockList(blockFn)
		if err != nil {
			return err
		}

		err = ResetDNS()
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Blocker) UpdateBlockList(parseLine func(string) string) error {
	var data []byte

	file, err := os.Open(b.HostsFile)
	if err != nil {
		return err
	}
	defer file.Close()

	sc := bufio.NewScanner(file)

	var done = false
	for sc.Scan() {
		line := sc.Text()
		if !done && len(line) > 0 {
			if strings.Contains(line, StopToken) {
				done = true
			}
			line = parseLine(line)
		}
		strBytes := []byte(line + "\n")
		data = append(data, strBytes...)
	}

	if err = sc.Err(); err != nil {
		return err
	}

	err = truncateFile(data, b.HostsFile)
	if err != nil {
		return err
	}

	return nil
}

func truncateFile(content []byte, destinationPath string) error {
	file, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return err
	}
	return nil
}
