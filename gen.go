package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/wangxudong123/switcher/model"
	"go/format"
	"golang.org/x/mod/modfile"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type generator struct {
	buf         bytes.Buffer
	indent      string
	outFilePath string
	pkg         string
	mockType    string
}

func (g *generator) Generate(pkg *model.Package, outFilePath string) error {
	g.pkg = pkg.Name
	g.outFilePath = outFilePath
	g.generateHead(g.pkg, outFilePath, pkg.Imports)
	for _, outStruct := range pkg.Struct {
		g.structName(outStruct.Name, outStruct.Field)
		g.funcName(outStruct.Name, g.mockType, outStruct.OldName, outStruct.Field)
	}

	_, err := g.Output()
	if err != nil {
		return err
	}
	log.Println(fmt.Sprintf("[info] generate package:%s in %s", g.pkg, outFilePath))
	return nil
}

func (g *generator) generateHead(outputPkgName string, outputPackagePath string, imports []string) {
	p, err := parsePackageImport(outputPackagePath)
	if err != nil {
		panic(err)
	}
	ps := strings.Split(p, "/")
	if len(ps) > 1 {
		outputPkgName = ps[len(ps)-1]
	} else {
		outputPkgName = ps[0]
	}

	g.p("package %v", outputPkgName)
	g.p("")

	if len(imports) > 0 {
		g.p("import (")
		g.in()
		g.mockType = g.pkg

		for _, pkgPath := range imports {
			if pkgPath == p {
				g.mockType = ""
				continue
			}
			g.p("%q", pkgPath)
		}

		g.out()
		g.p(")")
	}
}

func (g *generator) structName(name string, fields []model.Field) {
	g.p("type %v struct {", name)
	g.in()
	for _, field := range fields {
		g.p("%v %v", field.Name, field.Type)
	}
	g.out()
	g.p("}")
	g.p("")
}

func (g *generator) funcName(newStructName, pkgName string, oldStructNames []string, fields []model.Field) {
	if pkgName != "" {
		pkgName = pkgName + "."
	}

	for _, oldStructName := range oldStructNames {
		g.p("func New%vTo%v(data *%v%v) *%v {", oldStructName, newStructName, pkgName, oldStructName, newStructName)
		g.p("return &%v{", newStructName)
		g.in()
		for _, field := range fields {
			g.p("%v : data.%v,", field.Name, field.Name)
		}
		g.out()
		g.in()
		g.p("}")
		g.out()
		g.p("}")
		g.p("")

		//相反的
		g.p("func New%vTo%v(data *%v) *%v%v {", newStructName, oldStructName, newStructName, pkgName, oldStructName)
		g.p("return &%v%v{", pkgName, oldStructName)
		g.in()
		for _, field := range fields {
			g.p("%v : data.%v,", field.Name, field.Name)
		}
		g.out()
		g.in()
		g.p("}")
		g.out()
		g.p("}")
		g.p("")
	}
}

func (g *generator) p(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, g.indent+format+"\n", args...)
}

func (g *generator) in() {
	g.indent += "\t"
}

func (g *generator) out() {
	if len(g.indent) > 0 {
		g.indent = g.indent[0 : len(g.indent)-1]
	}
}

func (g *generator) Output() (n int, err error) {
	if g.outFilePath == "" {
		g.outFilePath = "./switcher/" + g.pkg + "/" + g.pkg + ".go"
	}

	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		log.Fatalf("Failed to format generated source code: %s\n%s", err, g.buf.String())
	}

	dst := os.Stdout
	//if len(g.dstFileName) > 0 {
	if err := os.MkdirAll(filepath.Dir(g.outFilePath), os.ModePerm); err != nil {
		log.Fatalf("Unable to create directory: %v", err)
	}
	var f *os.File
	//var err error
	//if g.head {
	f, err = os.Create(g.outFilePath)
	//} else {
	f, err = os.OpenFile(g.outFilePath, os.O_RDWR|os.O_APPEND, 0666)
	//}

	if err != nil {
		log.Fatalf("Failed opening destination file: %v", err)
	}
	defer dst.Close()
	dst = f
	//}

	return dst.Write(src)
}

func (g *generator) packName() {
	s := strings.Split(g.outFilePath, ".")
	if len(s) > 1 {
		if s[len(s)-1] == "go" {
			s = strings.Split(g.outFilePath, "/")
			if len(s) == 1 || (len(s) == 2 && s[0] == ".") {
				g.pkg = "main"
			}
		}
	}
}

var errOutsideGoPath = errors.New("Source directory is outside GOPATH")

func parsePackageImport(srcDir string) (string, error) {

	re, _ := regexp.Compile("[_a-zA-Za-zA-z]+\\.[a-zA-Z]+")
	srcDir = strings.Trim(re.ReplaceAllString(srcDir, ""), " ")

	moduleMode := os.Getenv("GO111MODULE")
	// trying to find the module
	if moduleMode != "off" {
		currentDir := srcDir
		for {
			dat, err := ioutil.ReadFile(filepath.Join(currentDir, "go.mod"))
			if os.IsNotExist(err) {
				if currentDir == filepath.Dir(currentDir) {
					// at the root
					break
				}
				currentDir = filepath.Dir(currentDir)
				continue
			} else if err != nil {
				return "", err
			}
			modulePath := modfile.ModulePath(dat)
			return filepath.ToSlash(filepath.Join(modulePath, strings.TrimPrefix(srcDir, currentDir))), nil
		}
	}
	// fall back to GOPATH mode
	goPaths := os.Getenv("GOPATH")
	if goPaths == "" {
		return "", fmt.Errorf("GOPATH is not set")
	}
	goPathList := strings.Split(goPaths, string(os.PathListSeparator))
	for _, goPath := range goPathList {
		sourceRoot := filepath.Join(goPath, "src") + string(os.PathSeparator)
		if strings.HasPrefix(srcDir, sourceRoot) {
			return filepath.ToSlash(strings.TrimPrefix(srcDir, sourceRoot)), nil
		}
	}
	return "", errOutsideGoPath
}
