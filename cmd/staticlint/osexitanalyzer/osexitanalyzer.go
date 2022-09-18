package osexitanalyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "os.Exit presence for the main function in the main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if pass.Pkg.Name() != "main" {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if c, ok := node.(*ast.FuncDecl); ok {
				if c.Name.Name == "main" {
					checkOsExitCall(pass, node)
				}
			}
			return true
		})
	}
	return nil, nil
}

func checkOsExitCall(pass *analysis.Pass, node ast.Node) {
	ast.Inspect(node, func(node ast.Node) bool {
		c, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		s, ok := c.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		i, ok := s.X.(*ast.Ident)
		if !ok {
			return true
		}

		if i.Name == "os" && s.Sel.Name == "Exit" {
			pass.Reportf(c.Pos(), "os.exit detected")
		}

		return true
	})
}
