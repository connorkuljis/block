package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Blocker struct {
	File string
}

func StartBlockerWrapper(blocker Blocker) {
	blocker.Start()
	fmt.Printf("Blocker:\tstarted\n")
}

func StopBlockerWrapper(blocker Blocker) {
	blocker.Stop()
	fmt.Printf("Blocker:\tstopped\n")
}

func NewBlocker() Blocker {
	file := "/etc/hosts"
	return Blocker{File: file}
}

func (b *Blocker) Start() int {
	n, err := b.Block()
	if err != nil {
		log.Println(err)
	}
	if flags.Verbose {
		fmt.Printf("Removed comments from %s (%d) bytes written.\n", b.File, n)
	}
	return n
}

func (b *Blocker) Stop() int {
	n, err := b.Unblock()
	if err != nil {
		log.Println(err)
	}
	if flags.Verbose {
		fmt.Printf("Added comments from %s (%d) bytes written.\n", b.File, n)
	}
	return n
}

func (b *Blocker) Unblock() (int, error) {
	insertCommentCharacterFromLine := func(line string) string {
		if string(line[0]) == "#" {
			return line
		}
		return "# " + line
	}
	n, err := b.UpdateBlockListAndRestartDNS(insertCommentCharacterFromLine)
	if err != nil {
		return n, err
	}
	return n, nil
}

func (b *Blocker) Block() (int, error) {
	removeCommentCharacterFromLine := func(line string) string {
		if string(line[0]) == "#" {
			return strings.TrimSpace(line[1:])
		}
		return line
	}
	n, err := b.UpdateBlockListAndRestartDNS(removeCommentCharacterFromLine)
	if err != nil {
		return n, err
	}
	return n, nil
}

func (b *Blocker) UpdateBlockListAndRestartDNS(edit func(string) string) (int, error) {
	totalBytes := 0

	file, err := os.Open(b.File)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return totalBytes, err
	}
	defer file.Close()

	tmpFile, err := os.CreateTemp("", "temporary")
	if err != nil {
		return totalBytes, err
	}
	defer tmpFile.Close()

	scanner := bufio.NewScanner(file)

	stopToken := "~" // lines below this character should not be manipulated in the hosts file.
	isProcessingDone := false
	for scanner.Scan() {
		line := scanner.Text()
		if !isProcessingDone && len(line) != 0 {
			if strings.Contains(line, stopToken) {
				isProcessingDone = true
			} else {
				line = edit(line)
			}
		}
		n, err := tmpFile.WriteString(line + "\n")
		if err != nil {
			return totalBytes, err
		}
		totalBytes += n
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}

	cmd := exec.Command("sudo", "mv", tmpFile.Name(), b.File)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return totalBytes, err
	}

	err = resetDNS()
	if err != nil {
		log.Println(err)
	}

	return totalBytes, nil
}

func resetDNS() error {
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
