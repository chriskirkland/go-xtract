package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type spec struct {
	Command string `yaml:"cmd"`
	Output string `yaml:"output"`
}

func TestE2E(t *testing.T) {
	matches, err := filepath.Glob("*/test.golden")
	require.NoError(t, err, "failed to find test specs")

	for _, match := range matches {
		testName := strings.Split(match, "/")[0]
		runner := newRunner(t, testName)
		runner.Run()
	}
}

type testRunner struct {
	t *testing.T
	testName string
	testSpec spec
}

func newRunner(t *testing.T, name string) *testRunner {
	return &testRunner{
		t: t,
		testName: name,
	}
}

func (r *testRunner) Run() {
	r.t.Run(r.testName, func(t *testing.T){
		r.t = t

		r.loadSpec()
		t.Logf("test spec: \n%+v\n", r.testSpec)

		output := r.runCommand()
		t.Logf("test output: \n%+v\n", output)

		r.verifyOutput(output)
	})
}


func (r *testRunner) loadSpec() {
	// read test spec
	specData, err := ioutil.ReadFile(fmt.Sprintf("%s/test.golden", r.testName))
	require.NoError(r.t, err, "failed to read test spec file")
	r.t.Logf("spec data: |\n%s", specData)

	var spec spec
	err = yaml.Unmarshal(specData, &spec)
	require.NoError(r.t, err, "failed to parse test spec")

	r.testSpec = spec
}

func (r *testRunner) runCommand() string {
	args := strings.Split(r.testSpec.Command, " ")
	first, rest := args[0], args[1:]

	cmd := exec.Command(first, rest...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Dir = r.testName

	err := cmd.Run()
	require.NoError(r.t, err, "command failed: '%s %s'", first, strings.Join(rest, " "))
	return output.String()
}

func (r *testRunner) verifyOutput(actual string) {
	assert.Equal(r.t, r.testSpec.Output, actual)
}


