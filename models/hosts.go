package models

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type HostsFile struct {
	File string
}

func NewBlocker() HostsFile {
	file := "/etc/hosts"
	return HostsFile{File: file}
}

func resetDNS() error {
	cmd := exec.Command("sudo", "dscacheutil", "-flushcache")
	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("sudo", "killall", "-HUP", "mDNSResponder")
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (h *HostsFile) Unblock() (int, error) {
	insertCommentCharacterFromLine := func(line string) string {
		if string(line[0]) == "#" {
			return line
		}
		return "# " + line
	}

	n, err := h.UpdateBlockList(insertCommentCharacterFromLine)
	if err != nil {
		return n, err
	}

	err = resetDNS()
	if err != nil {
		return n, err
	}

	return n, nil
}

func (h *HostsFile) Block() (int, error) {
	removeCommentCharacterFromLine := func(line string) string {
		if string(line[0]) == "#" {
			return strings.TrimSpace(line[1:])
		}
		return line
	}

	n, err := h.UpdateBlockList(removeCommentCharacterFromLine)
	if err != nil {
		return n, err
	}

	err = resetDNS()
	if err != nil {
		return n, err
	}

	return n, nil
}

func (h *HostsFile) UpdateBlockList(edit func(string) string) (int, error) {
	totalBytes := 0

	file, err := os.Open(h.File)
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

	cmd := exec.Command("sudo", "mv", tmpFile.Name(), h.File)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return totalBytes, err
	}

	return totalBytes, nil
}
