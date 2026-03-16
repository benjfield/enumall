package main

import (
	"flag"
	"go/ast"
	"go/token"
	"log"
	"strings"
	"text/template"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/tools/go/packages"
)

var typeNames = flag.String("type", "", "comma-separated list of type names; must be set")
var tmpl = template.New("all")

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// inspect traverses AST node and stores all const names of given type name.
func inspect(node ast.Node, gen *generator) bool {
	decl, ok := node.(*ast.GenDecl)
	if !ok || decl.Tok != token.CONST {
		return true
	}

	// The name of the type of the constants we are declaring.
	// Can change if this is a multi-element declaration.
	typ := ""

	for _, spec := range decl.Specs {
		// Guaranteed to succeed as this is CONST.
		vspec := spec.(*ast.ValueSpec)

		if vspec.Type == nil && len(vspec.Values) > 0 {
			// "X = 1" with no type but a value.
			typ = ""

			ce, ok := vspec.Values[0].(*ast.CallExpr)
			if !ok {
				continue
			}
			id, ok := ce.Fun.(*ast.Ident)
			if !ok {
				continue
			}
			typ = id.Name
		}
		if vspec.Type != nil {
			// "X T". Type is defined.
			ident, ok := vspec.Type.(*ast.Ident)
			if !ok {
				continue
			}
			typ = ident.Name
		}
		if typ != gen.TypeName {
			// This is not the type we're looking for.
			continue
		}

		for _, name := range vspec.Names {
			if name.Name == "_" {
				continue
			}

			// Add the value name to the list.
			gen.Values = append(gen.Values, name.Name)
		}
	}

	return false
}

// loadPackage loads the package from go:generate file.
func loadPackage(patterns []string) *packages.Package {
	cfg := &packages.Config{
		Mode:  packages.NeedSyntax | packages.NeedName,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, patterns...)
	handleError(err)
	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages found", len(pkgs))
	}

	return pkgs[0]
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	log.SetPrefix("enumall: ")
	types := strings.Split(*typeNames, ",")

	args := flag.Args()
	if len(args) == 0 {
		// Default: process whole package in current directory.
		args = []string{"."}
	}

	pkg := loadPackage(args)
	spew.Dump(pkg)

	for _, s := range pkg.Syntax {
		for _, lookupTypeName := range types {
			gen := &generator{
				PackageName: pkg.Name,
				TypeName:    lookupTypeName,
			}
			ast.Inspect(s, func(n ast.Node) bool { return inspect(n, gen) })
			gen.generate()
		}
	}
}
