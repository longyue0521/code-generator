package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/longyue0521/code-generator/annotation"
	"github.com/longyue0521/code-generator/template"
	visitor "github.com/longyue0521/code-generator/visitor/file"
)

var (
	errParamInvalid  = errors.New("gen: 方法必须接收两个参数，其中第一个参数是 context.Context，第二个参数请求")
	errResultInvalid = errors.New("gen: 方法必须返回两个参数，其中第一个返回值是响应，第二个返回值是error")
)

// 实际上 main 函数这里要考虑接收参数
// src 源目标
// dst 目标目录
// type src 里面可能有很多类型，那么用户可能需要指定具体的类型
// 这里我们简化操作，只读取当前目录下的数据，并且扫描下面的所有源文件，然后生成代码
// 在当前目录下运行 go install 就将 main 安装成功了，
// 可以在命令行中运行 gen
// 在 testdata 里面运行 gen，则会生成能够通过所有测试的代码
func main() {
	err := gen(".")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("success")
}

func gen(src string) error {
	// 第一步找出符合条件的文件
	srcFiles, err := scanFiles(src)
	if err != nil {
		return err
	}
	// 第二步，AST 解析源代码文件，拿到 service definition 定义
	defs, err := parseFiles(srcFiles)
	if err != nil {
		return err
	}
	// 生成代码
	return genFiles(src, defs)
}

// 根据 defs 来生成代码
// src 是源代码所在目录，在测试里面它是 ./testdata
func genFiles(dir string, defs []template.ServiceDefinition) error {
	for _, def := range defs {
		fileName := dir + "/" + underscoreName(def.Name) + "_gen.go"
		f, err := os.Create(fileName)
		if err != nil {
			return err
		}

		if err := template.Gen(f, def); err != nil {
			return err
		}
	}
	return nil
}

func parseFiles(srcFiles []string) ([]template.ServiceDefinition, error) {
	defs := make([]template.ServiceDefinition, 0, 20)
	for _, src := range srcFiles {
		f, err := parser.ParseFile(token.NewFileSet(), src, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		v := &visitor.SingleFileEntryVisitor{}
		ast.Walk(v, f)

		// 你需要利用 annotation 里面的东西来扫描 src，然后生成 file
		var file annotation.ASTFile

		file = v.Get()
		for _, typ := range file.Types {
			_, ok := typ.Group.Get("HttpClient")
			if !ok {
				continue
			}
			def, err := parseServiceDefinition(file.Node.Name.Name, typ)
			if err != nil {
				return nil, err
			}
			defs = append(defs, def)
		}
	}
	return defs, nil
}

// 你需要利用 typ 来构造一个 template.ServiceDefinition
// 注意你可能需要检测用户的定义是否符合你的预期
func parseServiceDefinition(pkg string, typ annotation.ASTType) (template.ServiceDefinition, error) {
	defs := template.ServiceDefinition{Package: pkg, Name: typ.Node.Name.Name}

	name, ok := typ.Get("ServiceName")
	if ok {
		defs.Name = name.Value
	}

	for _, field := range typ.Fields {

		name := field.Node.Names[0].Name

		method := template.ServiceMethod{
			Name:         name,
			Path:         "/" + name,
			ReqTypeName:  "",
			RespTypeName: "",
		}

		path, ok := field.Get("Path")
		if ok {
			method.Path = path.Value
		}

		fn, ok := field.Node.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		if len(fn.Params.List) < 2 {
			return defs, errParamInvalid
		}

		for _, param := range fn.Params.List {
			pointerType, ok := param.Type.(*ast.StarExpr)
			if !ok {
				continue

			}
			name, ok := pointerType.X.(*ast.Ident)
			if !ok {
				continue
			}
			method.ReqTypeName = name.String()
		}

		if len(fn.Results.List) < 2 {
			return defs, errResultInvalid
		}

		for _, result := range fn.Results.List {
			pointerType, ok := result.Type.(*ast.StarExpr)
			if !ok {
				continue
			}
			name, ok := pointerType.X.(*ast.Ident)
			if !ok {
				continue
			}
			method.RespTypeName = name.String()
		}

		defs.Methods = append(defs.Methods, method)
	}

	return defs, nil
}

// 返回符合条件的 Go 源代码文件，也就是你要用 AST 来分析这些文件的代码
func scanFiles(src string) ([]string, error) {
	var fileNames []string
	err := filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			abs, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			fileNames = append(fileNames, abs)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return fileNames, nil
}

// underscoreName 驼峰转字符串命名，在决定生成的文件名的时候需要这个方法
// 可以用正则表达式，然而我写不出来，我是正则渣
func underscoreName(name string) string {
	var buf []byte
	for i, v := range name {
		if unicode.IsUpper(v) {
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			buf = append(buf, byte(v))
		}

	}
	return string(buf)
}
