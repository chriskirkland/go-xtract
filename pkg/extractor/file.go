package extractor

import (
	"go/ast"
	"log"

	"github.com/chriskirkland/go-xtract/pkg/util"
)

// ProcessFiles process each file with the provided Extractor
func ProcessFiles(extractor Extractor, files ...string) {
	for _, filename := range files {
		log.Printf("processing file %s", filename)

		file, err := util.ParseGoFile(filename)
		if err != nil {
			log.Fatalf("failed to parse file as AST: %s", err.Error())
		}

		// load in file imports & variable declarations
		extractor.Load(file, filename)

		// walk the target file for translation texts
		ast.Walk(extractor, file)
	}
}
