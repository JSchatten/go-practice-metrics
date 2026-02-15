package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"

	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type MethodInfo struct {
	StructName string
	MethodBody string
}

type FilesInfo struct {
	FilePath string
	Methods  []MethodInfo
}

const root = "."

func main() {
	fileSets := token.NewFileSet()

	filesInfos := make(map[string]FilesInfo)

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

						// fmt.Println(structType)
						// Check for  "generate:reset" comment
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
						methodCode := generateResetMethod(typeSpec.Name.Name, structType)
						info, exists := filesInfos[path]
						if !exists {
							info = FilesInfo{
								FilePath: path,
								Methods:  []MethodInfo{},
							}
						}
						info.Methods = append(info.Methods, MethodInfo{
							StructName: typeSpec.Name.Name,
							MethodBody: methodCode,
						})
						filesInfos[path] = info
					}
				}

			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error walking the path: %v\n", err)
	}

	// Write reset.gen.go files
	for _, fileInfo := range filesInfos {
		if len(fileInfo.Methods) == 0 {
			continue
		}
		outputFile := filepath.Join(fileInfo.FilePath, "reset.gen.go")
		var buf strings.Builder
		buf.WriteString("package " + filepath.Base(fileInfo.FilePath) + "\n")
		// Add Reset methods

		for _, methodInfo := range fileInfo.Methods {
			firstLetter := strings.ToLower(methodInfo.StructName[:1])
			buf.WriteString("\n// Reset resets the struct to its zero values.\n")
			buf.WriteString(fmt.Sprintf("func ("+firstLetter+" *%s) Reset() {\n", methodInfo.StructName))
			buf.WriteString("\tif " + firstLetter + " == nil {\n\t\treturn\n\t}\n")
			buf.WriteString(methodInfo.MethodBody)
			buf.WriteString("\n}\n")
		}
		if err := os.WriteFile(outputFile, []byte(buf.String()), 0644); err != nil {
			log.Fatalf("Error writing file %s: %v\n", outputFile, err)
		}
	}
}

func generateResetMethod(structName string, structType *ast.StructType) string {
	var code strings.Builder
	// Получаем первую букву названия структуры в нижнем регистре
	firstLetter := strings.ToLower(structName[:1])
	for _, field := range structType.Fields.List {
		fieldName := field.Names[0].Name
		fieldType := field.Type

		// Handle pointers
		isPointer := false
		if ptr, ok := fieldType.(*ast.StarExpr); ok {
			fieldType = ptr.X
			isPointer = true
		}

		fieldCode := fmt.Sprintf("\t"+firstLetter+".%s", fieldName)

		// Check for slices
		if arrayType, ok := fieldType.(*ast.ArrayType); ok {
			if arrayType.Len != nil {
				// fixed array
				code.WriteString(fieldCode + " = " + fmt.Sprintf("%#v", reflect.Zero(reflect.TypeOf((*ast.ArrayType)(nil)).Elem()).Interface()))
			} else {
				// slice
				code.WriteString(fieldCode + " = " + fieldCode + "[:0]")
			}
			code.WriteString("\n")
			continue
		}

		// Check for maps
		if _, ok := fieldType.(*ast.MapType); ok {
			code.WriteString("\tif " + firstLetter + "." + fieldName + " != nil {\n")
			code.WriteString("\t\tclear(" + firstLetter + "." + fieldName + ")\n")
			code.WriteString("\t}\n")
			continue
		}

		// Handle primitive types and structs
		if ident, ok := fieldType.(*ast.Ident); ok {
			if isResettableType(ident.Name) {
				if isPointer {
					code.WriteString("\n\tif " + firstLetter + "." + fieldName + " != nil {")
					code.WriteString("\n\t\t*" + firstLetter + "." + fieldName + " = " + zeroValue(ident.Name))
					code.WriteString("\n\t}")
				} else {
					code.WriteString("\n" + fieldCode + " = " + zeroValue(ident.Name))
				}
			} else {
				// Assume it's a struct with Reset method
				if isPointer {
					code.WriteString("\n\tif " + firstLetter + "." + fieldName + " != nil {")
					code.WriteString("\n\t\t" + firstLetter + "." + fieldName + ".Reset()")
					code.WriteString("\n\t}")
				} else {
					code.WriteString("\n" + fieldCode + ".Reset()")
				}
			}
			// code.WriteString("\n")
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
