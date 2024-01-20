package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
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

func (b *Blocker) UnblockAndReset() error {
	if !b.Disable {
		parseLineUnblock := func(line string) string {
			if string(line[0]) == "#" {
				return line
			}
			return "# " + line
		}
		err := b.UpdateBlockList(parseLineUnblock)
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

func (b *Blocker) BlockAndReset() error {
	if !b.Disable {
		parseLineBlock := func(line string) string {
			if string(line[0]) == "#" {
				return strings.TrimSpace(line[1:])
			}
			return line
		}
		err := b.UpdateBlockList(parseLineBlock)
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

	if flags.Verbose {
		log.Println(string(data))
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

	if flags.Verbose {
		fmt.Println("File content overwritten successfully!")
	}

	return nil
}

func ResetDNS() error {
	if runtime.GOOS == "darwin" {
		if flags.Verbose {
			fmt.Println("Flushing dscacheutil.")
		}
		cmd := exec.Command("sudo", "dscacheutil", "-flushcache")
		err := cmd.Run()
		if err != nil {
			return err
		}

		if flags.Verbose {
			fmt.Println("Terminating mDNSResponder. ")
		}
		cmd = exec.Command("sudo", "killall", "-HUP", "mDNSResponder")
		err = cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}
