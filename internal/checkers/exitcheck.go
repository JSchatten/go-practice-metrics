// Package checkers содержит пользовательские анализаторы кода, применяемые в проекте.
//
// - exitcheck: анализатор, запрещающий прямой вызов os.Exit в функции main пакета main.
// Вместо этого следует возвращать код ошибки и завершать приложение ч��рез логику верхнего уровня.
package checkers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

const Doc = `exitcheck анализирует использование os.Exit в функции main пакета main`

var Analyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  Doc,
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// Проверяем, что это пакет main
		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			// Ищем функцию main
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" {
				return true
			}

			// Проверяем тело функции main
			for _, stmt := range fn.Body.List {
				visitStmt(stmt, pass)
			}
			return false
		})
	}
	return nil, nil
}

func visitStmt(stmt ast.Stmt, pass *analysis.Pass) {
	ast.Inspect(stmt, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Проверяем, что вызывается os.Exit
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		if ident.Name == "os" && sel.Sel.Name == "Exit" {
			pass.Reportf(call.Pos(), "запрещён прямой вызов os.Exit в функции main пакета main")
		}
		return true
	})
}
