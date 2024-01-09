package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	StopToken = "~"
)

type Blocker struct {
	File string
}

func NewBlocker() Blocker {
	file := "/etc/hosts"
	return Blocker{File: file}
}

// func (b *Blocker) Start() error {
// 	err := b.Block()
// 	if err != nil {
// 		return err
// 	}

// 	err = resetDNS()
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (b *Blocker) Stop() error {
// 	err := b.Unblock()
// 	if err != nil {
// 		return err
// 	}

// 	err = resetDNS()
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func (b *Blocker) Unblock() error {
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
	return nil
}

func (b *Blocker) Block() error {
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
	return nil
}

func (b *Blocker) UpdateBlockList(parseLine func(string) string) error {
	hostsFile := b.File
	var content []byte

	hf, err := os.Open(hostsFile)
	if err != nil {
		return err
	}

	sc := bufio.NewScanner(hf)

	done := false

	for sc.Scan() {
		line := sc.Text()
		if !done && len(line) > 0 {
			if strings.Contains(line, StopToken) {
				done = true
				break
			}
			line = parseLine(line) + "\n"
		}
		strBytes := []byte(line)
		content = append(content, strBytes...)
	}

	err = sc.Err()
	if err != nil {
		hf.Close()
		return err
	}

	if flags.Verbose {
		log.Println(string(content))
	}

	hf.Close()

	err = truncateFile(content, hostsFile)
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

	// Change the file permissions to allow writing without elevated privileges
	err = os.Chmod(destinationPath, 0666)
	if err != nil {
		return err
	}

	return nil
}

func ResetDNS() error {
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

	return nil
}
