package util

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"gopkg.in/godo.v2/glob"
)

func init() {
	// discard logger ouput by default
	log.SetOutput(ioutil.Discard)
}

const parserMode = parser.Mode(0) // default parsing

var fileSet = token.NewFileSet()


// NewJSONEncoder json encoder with two-space indentation and no HTML escape characters
func NewJSONEncoder(w io.Writer) *json.Encoder {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder
}

// ParseGoFile parses a Go AST from the provided file
func ParseGoFile(filename string) (*ast.File, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	src, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return parser.ParseFile(fileSet, filename, src, parserMode)
}

// FilesFromPatterns generates a list of Go file matching the glob-style wildcard patterns. Both unit tests and
// vendored files are omitted.
func FilesFromPatterns(patterns ...string) ([]string, error) {
	files := make(map[string]bool)
	assets, _, err := glob.Glob(patterns)
	if err != nil {
		return nil, err
	}

	for _, asset := range assets {
		if strings.Contains(asset.Path, "/vendor/") {
			// skip vendor directory
			continue
		}
		if strings.HasSuffix(asset.Path, "_test.go") {
			// skip unit tests
			continue
		}

		log.Printf("found file: %s", asset.Path)
		files[asset.Path] = true
	}

	uniqs := make([]string, 0, len(files))
	for file := range files {
		uniqs = append(uniqs, file)
	}
	return uniqs, nil
}
