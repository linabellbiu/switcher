package parse

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/wangxudong123/switcher/model"
	"os"
	"regexp"
	"strings"
)

type proto struct {
	pkg        *model.Package
	structType *structType
	indent     string
	outPath    string // 输出的文件路径
	comment    string // 注解信息
	head       bool
}

type structType struct {
	structName    string // 要转换结构体名字
	outStructName string // 输出的结构体名字
	ok            bool
}

func Proto(pkg *model.Package, path string) (*proto, error) {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// 每行读取
	scanner := bufio.NewScanner(f)
	b := new(proto)
	pkg.Struct = make(map[string]*model.Struct)
	b.pkg = pkg
	var packageName string
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "//@switcher") {
			b.comment = scanner.Text()
			b.analyzeComment()
		}
		if packageName == "" {
			// 抓取当前proto文件的package
			reg := regexp.MustCompile("^(package)\\s+[a-zA-Z0-9_]+;$")
			result := reg.FindString(scanner.Text())
			if result == "" {
				continue
			}
			packageName = strings.TrimRight(strings.Trim(strings.ReplaceAll(result, "package", ""), " "), ";")
			b.head = true
		}

		if b.head == false || b.structType == nil {
			continue
		}

		// 抓取/api所在的行
		if strings.Contains(scanner.Text(), "//@switcher") {
			pkg.Name = packageName
			b.comment = scanner.Text()
			for scanner.Scan() {
				text := scanner.Text()
				if text == "" {
					continue
				}

				reg := regexp.MustCompile("^//")
				result := reg.FindString(delSpace(text))
				if result != "" {
					continue
				}

				if !b.structType.ok {
					if err := b.AddStruct(text); err != nil {
						// TODO error.warp
						panic(fmt.Sprintf(err.Error()+";in: %s", path))
					}
					b.structType.ok = true
					continue
				}
				if strings.Trim(delExtraSpace(scanner.Text()), " ") == "}" {
					break
				}
				b.AddField(text)
			}
		}
	}
	err = scanner.Err()
	if err != nil {
		panic(err)
	}
	if pkg.Name == "" || len(pkg.Struct) == 0 {
		return nil, errors.New(fmt.Sprintf("not struct in :%s", path))
	}

	return b, nil
}

func (b *proto) AddStruct(text string) error {
	var s []string

	if s = strings.Split(delExtraSpace(text), " "); len(s) != 3 {
		return errors.New(fmt.Sprintf("not found struct name in: %s", text))
	}

	if strings.Trim(s[0], " ") != "message" || strings.Trim(s[2], " ") != "{" {
		return errors.New(fmt.Sprintf("error:%s", text))
	}

	b.structType.structName = strings.Trim(s[1], " ")

	if _, ok := b.pkg.Struct[b.structType.outStructName]; !ok {
		b.pkg.Struct[b.structType.outStructName] = &model.Struct{
			Name:    b.structType.outStructName,
			Field:   nil,
			OldName: []string{b.structType.structName},
		}
	} else {
		b.pkg.Struct[b.structType.outStructName].OldName = append(b.pkg.Struct[b.structType.outStructName].OldName, b.structType.structName)
	}
	return nil
}

func (b *proto) AddField(text string) {
	var s []string

	if s = strings.Split(delExtraSpace(text), "="); len(s) != 2 {
		panic(errors.New(fmt.Sprintf("field `=` ? :%s", text)))
	}
	// 提取标量类型,转换go类型
	goType := b.fieldType(delExtraSpace(s[0]))
	var (
		_struct *model.Struct
		ok      bool
	)

	if _struct, ok = b.pkg.Struct[b.structType.outStructName]; !ok {
		panic(errors.New(fmt.Sprintf("struct %s not found", b.structType.outStructName)))
	}

	if ss := strings.Split(delExtraSpace(s[0]), " "); len(ss) == 2 {

		newFieldName := marshal(delExtraSpace(ss[1]))
		for _, field := range _struct.Field {
			// 重复的字段
			if newFieldName == field.Name {
				return
			}
		}

		b.pkg.Struct[b.structType.outStructName].Field = append(
			b.pkg.Struct[b.structType.outStructName].Field, model.Field{
				Name: newFieldName,
				Type: goType,
			},
		)
	} else if len(ss) == 3 {
		newFieldName := marshal(delExtraSpace(ss[2]))
		for _, field := range _struct.Field {
			// 重复的字段
			if newFieldName == field.Name {
				return
			}
		}
		b.pkg.Struct[b.structType.outStructName].Field = append(
			b.pkg.Struct[b.structType.outStructName].Field, model.Field{
				Name: newFieldName,
				Type: goType,
			},
		)
	} else {
		panic(errors.New(fmt.Sprintf("field error: %s", text)))
	}
}

// 拆解注解
func (b *proto) analyzeComment() {
	comment := strings.TrimLeft(b.comment, "//@switcher")
	comment = strings.Trim(delExtraSpace(comment), " ")
	// comments := strings.Split(comment, ">")
	if err := b.arg(comment); err != nil {
		panic(errors.New(fmt.Sprintf("注解错误,%s", b.comment)))
	}
	// b.pkg.PkgPath = b.outPath
}

func (b *proto) imports(imports []string) {
	b.pkg.Imports = imports
}

func (b *proto) OutPath() string {
	return b.outPath
}

func (b *proto) arg(text string) error {
	arg := strings.Split(text, " ")
	if len(arg) <= 1 {
		return errors.New("not found arg!")
	}
	switch arg[0] {
	case "protoGoSrc":
		reg := regexp.MustCompile("^\\[[a-zA-Z0-9_/\\-\\.]+\\]$")
		result := reg.FindString(arg[1])
		result = strings.TrimRight(result, "]")
		b.imports([]string{strings.TrimLeft(result, "[")})
	case "out":
		b.outPath = delSpace(arg[1])
	case "struct":
		b.structType = &structType{
			outStructName: delSpace(arg[1]),
		}
	default:
	}
	return nil
}

func (b *proto) fieldType(text string) string {
	s := strings.Split(text, " ")
	var goType string
	switch s[0] {
	case "repeated":
		if len(s) == 3 {
			if _goType, ok := t[s[1]]; ok {
				goType = "[]" + _goType
			} else {
				// 是否存在导入的包
				if strings.Index(s[1], ".") == -1 {
					goType = "[]*" + b.pkg.Name + "." + s[1]
				} else {
					goType = "[]*" + s[1]
				}
			}
		}
	case "reserved":
		panic("暂时不支持 `reserved`")
	case "enum":
		panic("暂时不支持 `enum`")
	default:
		if len(s) == 2 {
			if _goType, ok := t[s[0]]; ok {
				goType = _goType
			} else {
				goType = "*" + b.pkg.Name + "." + s[0]
			}
		} else {
			panic(errors.New(fmt.Sprintf("field error: %s", text)))
		}
	}
	return goType
}

var t = map[string]string{
	"double":   "float64",
	"float":    "float32",
	"int32":    "int32",
	"sint32":   "int32",
	"int64":    "int64",
	"sint64":   "int64",
	"uint32":   "uint32",
	"uint64":   "uint64",
	"bool":     "bool",
	"string":   "string",
	"fixed32":  "unit32",
	"fixed64":  "unit64",
	"sfixed32": "unit32",
	"sfixed64": "unit64",
	"byte":     "[]byte",
}
