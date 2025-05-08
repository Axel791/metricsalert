// Package exitinmain содержит анализатор, запрещающий прямые вызовы
// os.Exit в функции main пакета main.
package exitinmain

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Doc = `запрещает прямой вызов os.Exit в функции main пакета main`

var Analyzer = &analysis.Analyzer{
	Name: "exitinmain",
	Doc:  Doc,
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	var mainDecl *ast.FuncDecl
	for _, f := range pass.Files {
		for _, d := range f.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
				mainDecl = fn
				break
			}
		}
	}
	if mainDecl == nil {
		return nil, nil
	}

	ast.Inspect(mainDecl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok &&
				ident.Name == "os" && sel.Sel.Name == "Exit" {
				pass.Reportf(call.Pos(), "запрещён прямой вызов os.Exit в main")
			}
		}
		return true
	})

	return nil, nil
}
