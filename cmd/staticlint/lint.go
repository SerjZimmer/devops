// Package staticlint ...
package staticlint

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "report calls to os.Exit in main function of main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		fmt.Println(file.Name)
		if pass.Pkg.Name() == "main" {
			ast.Inspect(file, func(n ast.Node) bool {
				// find main function
				if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
					// find calls to os.Exit
					ast.Inspect(fn.Body, func(n ast.Node) bool {
						if call, ok := n.(*ast.CallExpr); ok {
							if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
								if id, ok := sel.X.(*ast.Ident); ok && id.Name == "os" && sel.Sel.Name == "Exit" {
									// report diagnostic
									pass.Reportf(call.Pos(), "do not use os.Exit in main function")
								}
							}
						}
						return true
					})
				}
				return true
			})
		}
	}
	return nil, nil
}
