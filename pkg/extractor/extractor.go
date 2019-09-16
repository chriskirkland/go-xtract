package extractor

import (
	"go/ast"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chriskirkland/go-xtract/pkg/util"
	"github.com/pkg/errors"
)

/* TODO(cmkirkla
- cleanup and 'segregate' imports. maybe allow users to filter which debug information gets printed via CLI flag?
*/

// Extractor responsible for extracting strings from Go files
type Extractor interface {
	ast.Visitor
	Load(*ast.File, string)

	Strings() []string
}

// New creates a new Extractor
func New(targetFuncPackage, targetFuncName string) Extractor {
	t := newExtractor()
	t.tfPackage = targetFuncPackage
	t.tfName = targetFuncName
	return t
}

// NewFromFunction create a new Extractor for the provided function object
func NewFromFunction(targetFunc interface{}) Extractor {
	t := newExtractor()
	//TODO(cmkirkla): extract the package and function name from the provided symbol itself using reflection
	return t
}

func newExtractor() *extractor {
	return &extractor{
		strings:   make(map[string]bool),
		imports:   make(map[string]string),
		symbols:   make(map[string]string),
		tfPackage: "fmt",
		tfName:    "Sprintf",
	}
}

// implements the ast.Visitor interface
type extractor struct {
	// target function information
	tfPackage string
	tfName    string

	// internal file information
	currentFile string
	imports     map[string]string
	symbols     map[string]string

	// extracted artifacts
	strings map[string]bool
}

// Visit visit a node in the go file's AST
func (r *extractor) Visit(node ast.Node) ast.Visitor {
	switch node.(type) {
	case *ast.CallExpr: // function call
		call := node.(*ast.CallExpr)
		function, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			break // wrong type
		}
		if function == nil {
			break // nil function pkg.name. should never happen.
		}

		pkg, ok := function.X.(*ast.Ident)
		if !ok {
			break // wrong type
		}
		if pkg == nil {
			break // function defined in the same package
		}
		pkgName := pkg.Name
		funcName := function.Sel.Name

		var err error
		if funcName != r.tfName || r.imports[pkgName] != r.tfPackage {
			break // wrong function
		}

		if len(call.Args) == 0 {
			log.Printf("skipping niladic call to target function")
			break //skip
		}

		targetNode := call.Args[0]
		log.Printf("string key: (%T) %+v", targetNode, targetNode)

		var value string
		switch targetNode.(type) {
		case *ast.BasicLit:
			value, ok = r.extractStringLiteral(targetNode)
		case *ast.Ident:
			value, ok = r.extractLocalConstVar(targetNode)
		case *ast.SelectorExpr:
			value, ok = r.extractImportedConstVar(targetNode)
		}
		if !ok {
			break // failed to extract
		}

		// unquote the string literal
		value, err = strconv.Unquote(value)
		if err == nil && value != "" {
			if _, ok := r.strings[value]; !ok {
				log.Printf("recorded new string: '%s'", value)
				r.strings[value] = true
			}
		}

		// stop traversing this branch of the tree
		return nil
	}

	return r
}

func (r extractor) extractStringLiteral(node ast.Node) (value string, ok bool) {
	literal := node.(*ast.BasicLit)
	if literal == nil || literal.Kind != token.STRING {
		return "", false
	}
	return literal.Value, true
}

func (r extractor) extractLocalConstVar(node ast.Node) (value string, ok bool) {
	ident := node.(*ast.Ident)
	symbol := ident.Name

	value, ok = r.symbols[symbol]
	if ok {
		return value, true
	}

	// symbol not defined in current file. need to scan other files in the package.
	currentImportPath := strings.TrimPrefix(
		filepath.Dir(r.currentFile),
		filepath.Join(os.Getenv("GOPATH"), "src"),
	)
	value, err := r.resolveSymbol(currentImportPath, symbol)
	if err != nil {
		log.Printf("unable to resolve symbol %s: %s", symbol, err.Error())
		return "", false
	}
	log.Printf("successfully resolved local symbol %s.%s = %s", currentImportPath, symbol, value)
	return value, true
}

func (r extractor) extractImportedConstVar(node ast.Node) (value string, ok bool) {
	function := node.(*ast.SelectorExpr)
	if function == nil {
		return "", false
	}

	pkg, _ := function.X.(*ast.Ident)
	if pkg == nil {
		// exported function should have non-nil package name. parser should have handled this.
		return "", false
	}
	pkgName := pkg.Name
	symbol := function.Sel.Name

	value, err := r.resolveSymbol(r.imports[pkgName], symbol)
	if err != nil {
		log.Printf("unable to resolve symbol %s.%s: %s", pkgName, symbol, err.Error())
		return "", false
	}
	log.Printf("successfully resolved %s.%s = %s", pkgName, symbol, value)
	return value, true
}

// Load loads import, const, and variable declarations from the provide go file AST
func (r *extractor) Load(file *ast.File, filename string) {
	r.currentFile = filename

	log.Printf("parsing import declarations for file: %s", filename)
	r.imports = make(map[string]string, len(file.Imports))
	for _, importSpec := range file.Imports {
		path := strings.Trim(importSpec.Path.Value, "\"")
		pathParts := strings.Split(path, "/")
		name := pathParts[len(pathParts)-1]
		if importSpec.Name != nil {
			name = importSpec.Name.Name
		}

		log.Printf("recorded named import %s -> %s\n", name, path)
		r.imports[name] = path
	}

	log.Printf("parsing global declarations for file: %s", filename)
	r.symbols = make(map[string]string, len(file.Decls))
	for _, decl := range file.Decls {
		log.Printf("declaration: (%T) %+v", decl, decl)

		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue // skip
		}

		if gd.Tok != token.VAR && gd.Tok != token.CONST {
			// not a const or var declaration block
			continue // skip
		}

		for _, spec := range gd.Specs {
			log.Printf("  %s spec: (%T) %+v", gd.Tok, spec, spec)
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				// not a constant or var assignment
				continue
			}

			for ix := range valueSpec.Values {
				name := valueSpec.Names[ix].Name
				valueExpr := valueSpec.Values[ix]
				log.Printf("    %s spec: %s = (%T) %+v", gd.Tok, name, valueExpr, valueExpr)

				value, ok := valueExpr.(*ast.BasicLit)
				if !ok {
					// not a string literal
					continue // skip
				}

				if value.Value == "" {
					// empty string
					continue // skip
				}
				log.Printf("    recorded symbol %s.%s = '%s'\n", file.Name, name, value.Value)
				r.symbols[name] = value.Value
			}
		}

	}
}

func (r extractor) Strings() []string {
	results := make([]string, 0, len(r.strings))
	for s := range r.strings {
		results = append(results, s)
	}
	return results
}

// resolve value of the given symbol in the provided package. must be declared in a 'const' or 'var' block.
func (r *extractor) resolveSymbol(path, name string) (string, error) {
	goSrc := filepath.Join(os.Getenv("GOPATH"), "src")
	if !strings.HasPrefix(path, goSrc) {
		path = filepath.Join(os.Getenv("GOPATH"), "src", path)
	}

	log.Printf("attempting to resolve symbol %s in %s", name, path)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return "", errors.Wrap(err, "failed to read dir for imported package")
	}

	// create separate generator instance to avoid overwriting file/translation data
	gen := newExtractor()

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := filepath.Join(path, file.Name())
		if filename == r.currentFile {
			// should not load the same file twice
			continue
		}
		log.Printf("scanning file %s for symbol %s...", filename, name)

		astFile, err := util.ParseGoFile(filename)
		if err != nil {
			log.Printf("failed to parse Go file: %s", err.Error())
		}

		// load globals
		gen.Load(astFile, filename)

		// check globals against provided symbol
		for symbol, value := range gen.symbols {
			if symbol == name {
				return value, nil
			}
		}
	}

	return "", errors.New("desired const/variable declaration not found")
}
