[![Build Status](https://travis-ci.org/fd0/dachs.svg?branch=master)](https://travis-ci.org/fd0/dachs)

# dachs

Watch for changes in outputs of commands and prints a report.

# Installation

Dachs requires Go version 1.4 or newer to compile. To build dachs, run the
following command:

```shell
$ go run build.go
```

Afterwards please find a binary of dachs in the current directory.

# Configuration

Have a look at the sample config file [`dachs.conf`](dachs.conf).



# Compatibility

Dachs follows [Semantic Versioning](http://semver.org) to clearly define which
versions are compatible. The configuration file and command-line parameters and
user-interface are considered the "Public API" in the sense of Semantic
Versioning.
