package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type MethodInfo struct {
	StructName string
	MethodBody []string
}

func main() {
	root := "."
	fileSets := token.NewFileSet()
	packageMethods := make(map[string]MethodInfo) // key: dir → info

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() || strings.HasSuffix(path, "_test") {
			return nil
		}

		// Ensure path is a directory
		if !info.IsDir() {
			return nil
		}

		astFile, err := parser.ParseDir(fileSets, path, nil, parser.ParseComments)
		if err != nil {
			return nil // skip invalid dirs
		}

		for _, pkg := range astFile {
			for _, file := range pkg.Files {
				for _, decl := range file.Decls {
					genDecl, ok := decl.(*ast.GenDecl)
					if !ok || genDecl.Tok != token.TYPE {
						continue
					}
					for _, spec := range genDecl.Specs {
						typeSpec, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}
						structType, ok := typeSpec.Type.(*ast.StructType)
						if !ok {
							continue
						}

						fmt.Println(structType)
						// Check for // generate:reset comment
						if genDecl.Doc == nil {
							continue
						}
						resetComment := false
						for _, comment := range genDecl.Doc.List {
							if strings.TrimSpace(comment.Text) == "// generate:reset" {
								resetComment = true
								break
							}
						}
						if !resetComment {
							continue
						}

						// Generate Reset method
						// method := generateResetMethod(typeSpec.Name.Name, structType)
						methodCode := generateResetMethod(structType)
						info, exists := packageMethods[path]
						if !exists {
							info = MethodInfo{
								StructName: typeSpec.Name.Name,
								MethodBody: []string{},
							}
							info.MethodBody = append(info.MethodBody, methodCode)
						}
						packageMethods[path] = info
						fmt.Println("NEXT STRUCT")
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking the path: %v\n", err)
		os.Exit(1)
	}

	// Write reset.gen.go files
	fmt.Println("=======\nSTART GENERATING")
	fmt.Println(packageMethods)
	fmt.Println("----------")
	for dir, info := range packageMethods {
		fmt.Println(dir, info)
		if len(info.MethodBody) == 0 {
			continue
		}
		outputFile := filepath.Join(dir, "reset.gen.go")
		var buf strings.Builder
		buf.WriteString("package " + filepath.Base(dir) + "\n\n")
		// Add Reset method
		buf.WriteString("// Reset resets the struct to its zero values.\n")
		buf.WriteString(fmt.Sprintf("func (s *%s) Reset() {\n", info.StructName))
		buf.WriteString("\tif s == nil {\n\t\treturn\n\t}\n")
		for _, line := range info.MethodBody {
			buf.WriteString(line + "\n")
		}

		buf.WriteString("}\n")
		if err := os.WriteFile(outputFile, []byte(buf.String()), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file %s: %v\n", outputFile, err)
			os.Exit(1)
		}
		fmt.Println("+++++++++++++")
	}
}

func generateResetMethod(structType *ast.StructType) string {
	var code strings.Builder
	fmt.Println(*structType)
	for _, field := range structType.Fields.List {
		fmt.Println(field.Names)
		fieldName := field.Names[0].Name
		fieldType := field.Type

		// Handle pointers
		isPointer := false
		if ptr, ok := fieldType.(*ast.StarExpr); ok {
			fieldType = ptr.X
			isPointer = true
		}

		fieldCode := fmt.Sprintf("\ts.%s", fieldName)

		// Check for slices
		if arrayType, ok := fieldType.(*ast.ArrayType); ok {
			if arrayType.Len != nil {
				// fixed array
				code.WriteString(fieldCode + " = " + fmt.Sprintf("%#v", reflect.Zero(reflect.TypeOf((*ast.ArrayType)(nil)).Elem()).Interface()) + "\n")
			} else {
				// slice
				code.WriteString(fieldCode + " = " + fieldCode + "[:0]\n")
			}
			// code.WriteString("\n")
			continue
		}

		// Check for maps
		if _, ok := fieldType.(*ast.MapType); ok {
			code.WriteString("\tif s." + fieldName + " != nil {\n")
			code.WriteString("\t\tclear(s." + fieldName + ")\n")
			code.WriteString("\t}\n")
			continue
		}

		// Handle primitive types and structs
		if ident, ok := fieldType.(*ast.Ident); ok {
			if isResettableType(ident.Name) {
				if isPointer {
					code.WriteString("\tif s." + fieldName + " != nil {\n")
					code.WriteString("\t\t*s." + fieldName + " = " + zeroValue(ident.Name) + "\n")
					code.WriteString("\t}\n")
				} else {
					code.WriteString(fieldCode + " = " + zeroValue(ident.Name) + "\n")
				}
			} else {
				// Assume it's a struct with Reset method
				if isPointer {
					code.WriteString("\tif s." + fieldName + " != nil {\n")
					code.WriteString("\t\ts." + fieldName + ".Reset()\n")
					code.WriteString("\t}\n")
				} else {
					code.WriteString(fieldCode + ".Reset()\n")
				}
			}
			continue
		}
	}
	return code.String()
}

func isResettableType(typeName string) bool {
	switch typeName {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "string", "bool":
		return true
	default:
		return false
	}
}

func zeroValue(typeName string) string {
	switch typeName {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return "0"
	case "float32", "float64":
		return "0.0"
	case "string":
		return "\"\""
	case "bool":
		return "false"
	default:
		return "nil"
	}
}
