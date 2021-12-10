package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func main() {
	if err := walkDir("/repo"); err != nil {
		panic(err)
	}
}

func walkDir(path string) error {
	if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err == nil &&
			!info.IsDir() &&
			!strings.HasPrefix(info.Name(), ".") &&
			!strings.Contains(path, "vendor") &&
			strings.HasSuffix(info.Name(), ".go") {
			err = processFile(path, nil, os.Stdout)
		}
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "walking filepath")
	}
	return nil
}

func processFile(filename string, in io.Reader, out io.Writer) error {
	if in == nil {
		f, err := os.Open(filename)
		if err != nil {
			return errors.Wrap(err, "os.Open failed")
		}
		in = f
	}

	src, err := ioutil.ReadAll(in)
	if err != nil {
		return errors.Wrap(err, "ioutil.ReadAll failed")
	}

	res, err := formatFile(filename, src)
	if err != nil {
		return errors.Wrap(err, "sqlfmt.Process failed")
	}

	if !bytes.Equal(src, res) {
		if err = ioutil.WriteFile(filename, res, 0); err != nil {
			return errors.Wrap(err, "ioutil.WriteFile failed")
		}
	}
	return nil
}

func formatFile(filename string, src []byte) ([]byte, error) {
	fset := token.NewFileSet()
	parserMode := parser.ParseComments

	astFile, err := parser.ParseFile(fset, filename, src, parserMode)
	if err != nil {
		return nil, errors.Wrap(err, "parser.ParseFile failed")
	}

	replaceAst(astFile, fset)

	var buf bytes.Buffer

	if err = printer.Fprint(&buf, fset, astFile); err != nil {
		return nil, errors.Wrap(err, "printer.Fprint failed")
	}

	out, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, errors.Wrap(err, "format.Source failed")
	}
	return out, nil
}

// TODO get errors out of here properly
func replaceAst(f *ast.File, fset *token.FileSet) {
	ast.Inspect(f, func(n ast.Node) bool {
		if x, ok := n.(*ast.CallExpr); ok {
			if fun, ok := x.Fun.(*ast.SelectorExpr); ok {
				if isSQLFunc(fun.Sel.Name) {
					for _, xArg := range x.Args {
						if arg, ok := xArg.(*ast.BasicLit); ok {
							sqlStmt := arg.Value
							if !strings.HasPrefix(sqlStmt, "`") {
								return true
							}
							src := strings.Trim(sqlStmt, "`")
							res, err := formatQuery(src)
							if err != nil {
								log.Println(fmt.Sprintf("format failed at %s: %v", fset.Position(arg.Pos()), err))
								return true
							}

							funPos := fset.Position(fun.Pos())
							n, err := getLineIndentationLevel(funPos.Filename, funPos.Line)
							if err != nil {
								log.Println(fmt.Sprintf("getting indentation failed at %s: %v", fset.Position(arg.Pos()), err))
								return true
							}

							parts := strings.Split(res, "\n")
							for i, p := range parts {
								parts[i] = strings.Repeat("\t", n+1) + p
							}
							res = strings.Join(parts, "\n")
							// TODO more gooder
							arg.Value = "`\n" + res + "`"
						}
					}
				}
			}
		}
		return true
	})
}

func isSQLFunc(f string) bool {
	funcNames := []string{
		"Get",
		"Select",
		"Exec",
		"NamedExec",
		"NamedQuery",
		"Query",
		"Prepare",
		"GetContext",
		"SelectContext",
		"ExecContext",
		"NamedExecContext",
		"QueryContext",
		"PrepareContext",
		"PrepareNamedContext",
	}
	for _, fn := range funcNames {
		if f == fn {
			return true
		}
	}
	return false
}

func formatQuery(src string) (string, error) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "sql-")
	if err != nil {
		log.Fatal("Cannot create temporary file", err)
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	if _, err = tmpFile.WriteString(src); err != nil {
		return "", errors.Wrap(err, "writing query to tmpfile")
	}

	cmd := exec.Command("pg_format", "--inplace", "--comma-break", "--tabs", tmpFile.Name())
	if err := cmd.Run(); err != nil {
		return "", errors.Wrap(err, "running pg_format")
	}

	formattedQuery, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		return "", errors.Wrap(err, "reading formatted query from temp file")
	}

	return string(formattedQuery), nil
}

func getLineIndentationLevel(path string, lineNo int) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	line, err := readLine(file, lineNo)
	if err != nil {
		return 0, errors.Wrap(err, "reading line")
	}

	return len(line) - len(strings.TrimLeft(line, " \t\n")), nil
}

func readLine(r io.Reader, lineNo int) (string, error) {
	currLine := 1
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		if currLine == lineNo {
			return sc.Text(), sc.Err()
		}
		currLine++
	}
	return "", io.EOF
}
