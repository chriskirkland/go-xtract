# go-xtract
<a href="https://travis-ci.com/chriskirkland/go-xtract.svg?branch=master" alt="build status">
  <img src="https://travis-ci.com/chriskirkland/go-xtract.svg?branch=master" /></a>

Library for extracting arbitrary strings from Go code.

## Usage

### Installing the CLI
```
go install github.com/chriskirkland/go-xtract/cmd/xtract
```

### Examples:
Help:
```
~$ xtract -h
Usage of xtract:
  -func string
     target func (default "fmt.Sprintf")
  -o string
      output file (default "<stdout>")
  -template string
      output template (default "{{range .Strings}}{{print .}}\n{{end}}")
  -v	enable debug output
```

Run for all Go file in the repo:
```
xtract **/*.go
```

Run over specific set of Go files:
```
xtract 'plugins/models/command_metadata.go' 'plugins/commands/cmdalb/*.go'
xtract 'pkg/*.go'
```

Write output to a file:
```
xtract -o plugins/i18n/resources/en_US.all.json
```
