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
	Interval int
	StateDir string `toml:"state_dir"`

	Commands []Command `toml:"command"`
}

var opts = &struct {
	Verbose  bool   `short:"v"   long:"verbose"                       description:"be verbose"`
	Config   string `short:"c"   long:"config"  env:"DACHS_CONFIG"    description:"use this config file"`
	StateDir string `short:"s"   long:"state"   env:"DACHS_STATE_DIR" description:"directory to use for saving states"`
	Force    bool   `short:"f"   long:"force"                         description:"force update of all commands"`
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

	for _, dir := range dirs {
		filename := filepath.Join(dir, "dachs.conf")
		_, err := os.Stat(filename)
		if err == nil {
			return filename, nil
		}

	}

	return "", errors.New("no config file found")
}

// cacheDir returns the cache directory according to the XDG specification.
func cacheDir() string {
	if d := os.Getenv("XDG_CACHE_DIR"); d != "" {
		return d
	}

	user, err := user.Current()
	if err == nil {
		return filepath.Join(user.HomeDir, ".cache")
	}

	return ""
}

func main() {
	var parser = flags.NewParser(opts, flags.Default)

	_, err := parser.Parse()
	if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
		os.Exit(0)
	}
	Erx(err, 1)

	if opts.Config == "" {
		opts.Config, err = findConfig()
		Erx(err, 1)
	}

	var cfg Config
	_, err = toml.DecodeFile(opts.Config, &cfg)
	Erx(err, 1)

	// update to defaults
	if cfg.Interval == 0 {
		cfg.Interval = 3600
	}

	if opts.Force {
		cfg.Interval = 0
	}

	var statedir = filepath.Join(cacheDir(), "dachs-state")
	if cfg.StateDir != "" {
		statedir = cfg.StateDir
	}

	if opts.StateDir != "" {
		statedir = opts.StateDir
	}

	V("using state directory %v\n", statedir)

	_, err = os.Stat(cfg.StateDir)
	if err != nil && os.IsNotExist(err) {
		V("creating state dir %v\n", cfg.StateDir)
		Erx(os.MkdirAll(cfg.StateDir, 0700), 1)
	}

	for _, cmd := range cfg.Commands {
		if cmd.Interval == 0 {
			cmd.Interval = cfg.Interval
		}

		diff, err := cmd.Execute(cfg.StateDir)
		if len(diff) > 0 {
			fmt.Printf("diff for command %v\n", cmd.Name)
			fmt.Printf("============================\n")
			fmt.Print(string(diff))
			fmt.Println()
		}

		if err != nil {
			E("error executing %v: %v\n", cmd.Run, err)
		}
	}
}
