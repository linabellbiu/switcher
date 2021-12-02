package main

import (
	_ "embed"
	"fmt"
	cmd "github.com/wangxudong123/easy-command"
	"github.com/wangxudong123/switcher/model"
	"github.com/wangxudong123/switcher/parse"
	"log"
	"os"
	"path/filepath"
)

var (
	fileBlacklist    []string
	fileWhitelist    []string
	postfixBlacklist []string
	postfixWhitelist = []string{".go", ".proto"}
)

var FindFile = make(chan string)

var (
	//go:embed cmd.yaml
	yamlBody []byte
	source   string
)

var (
	run   = make(map[string]func(cmd.FlagValueMap))
	_make = func(s cmd.FlagValueMap) {
		if source = s["path"].GetValueString(); source == "" {
			source, _ = os.Getwd()
			fmt.Println(source)
		}
	}
)

func main() {
	run["make"] = _make
	cmd.LoadCmd(run, yamlBody)
	filepath.Walk(source, Walkfunc)
}

func Walkfunc(path string, info os.FileInfo, err error) error {
	//过滤目录
	if info.IsDir() {
		return nil
	}
	ext := filepath.Ext(path)
	//黑白名单
	if !parse.In(ext, postfixWhitelist) || parse.In(ext, postfixBlacklist) {
		return nil
	}

	g := new(generator)
	pkg := new(model.Package)
	filePath := filepath.Dir(path) + "/" + filepath.Base(path)
	switch ext {
	case ".proto":
		b, err := parse.Proto(pkg, filePath)
		if err != nil {
			return nil
		}
		if err = g.Generate(pkg, b.OutPath()); err != nil {
			log.Println(err)
		}

	default:
		return nil
	}

	return nil
}
