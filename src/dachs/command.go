package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// Execute runs the command with the given filter functions.
func (c Command) Execute() error {
	V("execute command %q\n", c.Run)

	cmd := exec.Command("sh", "-c", c.Run)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return err
	}

	c.Compare(output)

	return nil
}

// State returns the filename for the state. It depends on the command
// executed.
func (c Command) State() string {
	hash := sha256.Sum256([]byte(c.Run))
	return hex.EncodeToString(hash[:])
}

// Compare compares the output to the old output by running `git diff`.
func (c Command) Compare(output []byte) (err error) {
	tempdir, err := ioutil.TempDir("", "dachs-compare-")
	defer func() {
		e := os.RemoveAll(tempdir)
		if err != nil {
			err = e
		}
	}()

	fileNew, err := os.Create(filepath.Join(tempdir, "new"))
	if err != nil {
		return err
	}

	_, err = fileNew.Write(output)
	if err != nil {
		return err
	}
	
	err = fileNew.Close()
	if err != nil {
		return err
	}

	fileOld, err := os.Create(filepath.Join(tempdir, "old"))

	err = fileOld.Close()
	if err != nil {
		return err
	}

	return nil
}
