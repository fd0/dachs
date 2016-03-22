package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// Command holds all data needed to start a command.
type Command struct {
	Name     string
	Run      string
	Filter   []string
	Interval int

	statefile string
	start     time.Time
}

// Execute runs the command with the given filter functions.
func (c Command) Execute(statedir string) ([]byte, error) {
	V("execute command %q\n", c.Run)

	c.statefile = filepath.Join(statedir, c.stateFilename())

	last := time.Now().Sub(c.lastRun())
	interval := time.Duration(c.Interval) * time.Second
	if last < interval {
		V("command has been run only %v ago, which is less than %v\n", last, interval)
		return nil, nil
	}

	cmd := exec.Command("sh", "-c", c.Run)
	cmd.Stderr = os.Stderr

	state, err := c.loadOldState()
	if err != nil {
		E("unable to read old state file: %v\n", err)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	diff, err := Compare(state, output)
	if err != nil {
		V("return error %#v\n", err)
	}

	return diff, c.saveNewState(output)
}

func (c Command) loadOldState() (state []byte, err error) {
	f, err := os.Open(c.statefile)
	if os.IsNotExist(err) {
		return nil, nil
	}
	defer func() {
		e := f.Close()
		if err == nil && e != nil {
			err = e
		}
	}()

	return ioutil.ReadAll(f)
}

func (c Command) saveNewState(state []byte) (err error) {
	f, err := os.Create(c.statefile)
	if err != nil {
		return err
	}
	defer func() {
		e := f.Close()
		if err == nil {
			err = e
		}
	}()

	_, err = f.Write(state)
	return err
}

func (c Command) stateFilename() string {
	hash := sha256.Sum256([]byte(c.Run))
	return hex.EncodeToString(hash[:])
}

func (c Command) lastRun() time.Time {
	fi, err := os.Stat(c.statefile)
	if os.IsNotExist(err) {
		return time.Unix(0, 0)
	}

	return fi.ModTime()
}

func writeToFile(filename string, content []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	_, err = file.Write(content)
	if err != nil {
		return err
	}

	return file.Close()
}

func diff(dir, file1, file2 string) ([]byte, error) {
	V("compare %v and %v\n", file1, file2)
	cmd := exec.Command("git", "diff", "--", file1, file2)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	buf, err := cmd.Output()

	if e, ok := err.(*exec.ExitError); ok {
		s := e.ProcessState.Sys().(syscall.WaitStatus)

		// ignore exit code 1, this just means there were differences
		if s.ExitStatus() == 1 {
			return buf, nil
		}
	}

	return buf, err
}

// Compare compares the output to the old output by running `git diff`.
func Compare(oldOutput, newOutput []byte) (output []byte, err error) {
	V("compare output to old output\n")

	tempdir, err := ioutil.TempDir("", "dachs-compare-")
	defer func() {
		e := os.RemoveAll(tempdir)
		if err != nil {
			err = e
		}
	}()

	err = writeToFile(filepath.Join(tempdir, "new"), newOutput)
	if err != nil {
		return nil, err
	}

	err = writeToFile(filepath.Join(tempdir, "old"), oldOutput)
	if err != nil {
		return nil, err
	}

	return diff(tempdir, "old", "new")
}
