package model

type Package struct {
	Name string // 包名
	// PkgPath string   //生产包的路径
	Imports []string // 要导入包的路径
	Struct  map[string]*Struct
	Func    []*Func
}

type Parameter struct {
	Name string
	Type string
}

type Struct struct {
	Name string
	// Methods map[string]*Method
	Field   []Field
	OldName []string
}

type Field struct {
	Name string
	Type string
}

type Func struct {
	In, Out []*Parameter
	Content string
}
