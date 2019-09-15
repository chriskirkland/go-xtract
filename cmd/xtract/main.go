package main

import (
	"flag"
	"html/template"
	"io"
	"log"
	"os"
	"strings"

	"github.com/chriskirkland/go-xtract/pkg/extractor"
	"github.com/chriskirkland/go-xtract/pkg/util"
)

const stdoutSentinel = "<stdout>"

var (
	targetFunc = flag.String("func", "fmt.Sprintf", "taget func")
	//TODO(cmkirkla): fix character escaping in default template
	outputTemplate = flag.String("template", "{{range .Strings}}{{print .}}\n{{end}}", "output template")
	outputFile     = flag.String("o", stdoutSentinel, "output file")
	debug          = flag.Bool("v", false, "enable debug output")
)

func main() {
	flag.Parse()

	if *debug {
		log.SetOutput(os.Stdout)
	}

	tf := strings.Split(*targetFunc, ".")
	if len(tf) != 2 {
		log.Fatalf("'-func' must be a valid qualified function name but found '%s'", *targetFunc)
	}
	tfPackage, tfName := tf[0], tf[1]

	if flag.NArg() == 0 {
		log.Fatalf("one or more file patterns must be provided")
	}
	globs := flag.Args()

	files, err := util.FilesFromPatterns(globs...)
	if err != nil {
		log.Fatalf("error resolving one more provide file pattern: %s", err.Error())
	}

	ext := extractor.New(tfPackage, tfName)
	extractor.ProcessFiles(ext, files...)

	var writer io.Writer = os.Stdout
	if *outputFile != stdoutSentinel {
		f, err := os.Open(*outputFile)
		if err != nil {
			log.Fatalf("output file '%s' not found: %s", *outputFile, err)
		}
		defer f.Close()

		writer = f
	}

	// generate user output
	t, err := template.New("output").Parse(*outputTemplate)
	if err != nil {
		log.Fatalf("failed to parse provided output template: %s", err)
	}

	log.Println("writing extracted strings")
	t.Execute(writer, struct {
		Strings []string
	}{
		Strings: ext.Strings(),
	})
}
