package main

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/jessevdk/go-flags"
)

// Config holds all configuration parameters set in the config file.
type Config struct {
	Interval       int
	StateDirectory string `toml:"state_dir"`

	Commands []Command `toml:"command"`
}

// Command holds all data needed to start a command.
type Command struct {
	Run    string
	Filter []string

	state string
}

var opts = &struct {
	Verbose bool   `short:"v" long:"verbose"                   description:"be verbose"`
	Config  string `short:"c" long:"config" env:"DACHS_CONFIG" description:"use this config file"`
}{}

// V prints the message when verbose is active.
func V(format string, args ...interface{}) {
	if !opts.Verbose {
		return
	}

	fmt.Printf(format, args...)
}

// E prints an error to stderr.
func E(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

// Er prints the error err if it is set.
func Er(err error) {
	if err == nil {
		return
	}

	E("error: %v\n", err)
}

// Erx prints the error and exits with the given code, but only if the error is non-nil.
func Erx(err error, exitcode int) {
	if err == nil {
		return
	}

	Er(err)
	os.Exit(exitcode)
}

// findConfig returns the first config file that could be found, following the
// xdg basedir spec.
func findConfig() (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		user, err := user.Current()
		if err == nil {
			configHome = filepath.Join(user.HomeDir, ".config")
		}
	}

	var dirs []string
	if configHome != "" {
		dirs = []string{configHome}
	}

	if os.Getenv("XDG_CONFIG_DIRS") != "" {
		configDirs := strings.Split(os.Getenv("XDG_CONFIG_DIRS"), ":")
		dirs = append(dirs, configDirs...)
	} else {
		dirs = append(dirs, "/etc/xdg")
	}

	fmt.Printf("dirs: %v\n", dirs)

	for _, dir := range dirs {
		filename := filepath.Join(dir, "dachs.conf")
		_, err := os.Stat(filename)
		if err == nil {
			return filename, nil
		}

	}

	return "", errors.New("no config file found")
}

func main() {
	var parser = flags.NewParser(opts, flags.Default)

	_, err := parser.Parse()
	if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
		os.Exit(0)
	}

	if err != nil {
		os.Exit(1)
	}

	if opts.Config == "" {
		opts.Config, err = findConfig()
		Erx(err, 1)
	}

	var cfg Config
	_, err = toml.DecodeFile(opts.Config, &cfg)
	Erx(err, 1)

	for _, cmd := range cfg.Commands {
		cmd.state = cfg.StateDirectory
		Er(cmd.Execute())
	}
}
