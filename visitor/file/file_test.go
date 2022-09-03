package visitor

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/longyue0521/code-generator/annotation"
	"github.com/stretchr/testify/assert"
)

func TestFileVisitor_Get(t *testing.T) {
	testCases := []struct {
		src  string
		want annotation.ASTFile
	}{
		{
			src: `
// annotation go through the source code and extra the annotation
// @author Deng Ming
/* @multiple first line
second line
*/
// @date 2022/04/02
package annotation

type (
	// FuncType is a type
	// @author Deng Ming
	/* @multiple first line
	   second line
	*/
	// @date 2022/04/02
	FuncType func()
)

type (
	// StructType is a test struct
	//
	// @author Deng Ming
	/* @multiple first line
	   second line
	*/
	// @date 2022/04/02
	StructType struct {
		// Public is a field
		// @type string
		Public string
	}

	// SecondType is a test struct
	//
	// @author Deng Ming
	/* @multiple first line
	   second line
	*/
	// @date 2022/04/03
	SecondType struct {
	}
)

type (
	// Interface is a test interface
	// @author Deng Ming
	/* @multiple first line
	   second line
	*/
	// @date 2022/04/04
	Interface interface {
		// MyFunc is a test func
		// @parameter arg1 int
		// @parameter arg2 int32
		// @return string
		MyFunc(arg1 int, arg2 int32) string

		// second is a test func
		// @return string
		second() string
	}
)
`,
			want: annotation.ASTFile{
				Group: annotation.Group[*ast.File]{
					List: []annotation.Single{
						{
							Key:   "author",
							Value: "Deng Ming",
						},
						{
							Key:   "multiple",
							Value: "first line\nsecond line\n",
						},
						{
							Key:   "date",
							Value: "2022/04/02",
						},
					},
				},
				Types: []annotation.ASTType{
					{
						Group: annotation.Group[*ast.TypeSpec]{
							List: []annotation.Single{
								{
									Key:   "author",
									Value: "Deng Ming",
								},
								{
									Key:   "multiple",
									Value: "first line\n\t   second line\n\t",
								},
								{
									Key:   "date",
									Value: "2022/04/02",
								},
							},
						},
					},
					{
						Group: annotation.Group[*ast.TypeSpec]{
							List: []annotation.Single{
								{
									Key:   "author",
									Value: "Deng Ming",
								},
								{
									Key:   "multiple",
									Value: "first line\n\t   second line\n\t",
								},
								{
									Key:   "date",
									Value: "2022/04/02",
								},
							},
						},
						Fields: []annotation.ASTField{
							{
								Group: annotation.Group[*ast.Field]{
									List: []annotation.Single{
										{
											Key:   "type",
											Value: "string",
										},
									},
								},
							},
						},
					},
					{
						Group: annotation.Group[*ast.TypeSpec]{
							List: []annotation.Single{
								{
									Key:   "author",
									Value: "Deng Ming",
								},
								{
									Key:   "multiple",
									Value: "first line\n\t   second line\n\t",
								},
								{
									Key:   "date",
									Value: "2022/04/03",
								},
							},
						},
					},
					{
						Group: annotation.Group[*ast.TypeSpec]{
							List: []annotation.Single{
								{
									Key:   "author",
									Value: "Deng Ming",
								},
								{
									Key:   "multiple",
									Value: "first line\n\t   second line\n\t",
								},
								{
									Key:   "date",
									Value: "2022/04/04",
								},
							},
						},
						Fields: []annotation.ASTField{
							{
								Group: annotation.Group[*ast.Field]{
									List: []annotation.Single{
										{
											Key:   "parameter",
											Value: "arg1 int",
										},
										{
											Key:   "parameter",
											Value: "arg2 int32",
										},
										{
											Key:   "return",
											Value: "string",
										},
									},
								},
							},
							{
								Group: annotation.Group[*ast.Field]{
									List: []annotation.Single{
										{
											Key:   "return",
											Value: "string",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		f, err := parser.ParseFile(token.NewFileSet(), "src.go", tc.src, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		tv := &SingleFileEntryVisitor{}
		ast.Walk(tv, f)
		file := tv.Get()
		assertAnnotations(t, tc.want.Group, file.Group)

		if len(tc.want.Types) != len(file.Types) {
			t.Fatal()
		}

		for i, typ := range file.Types {
			wantType := tc.want.Types[i]
			assertAnnotations(t, wantType.Group, typ.Group)
			if len(wantType.Fields) != len(typ.Fields) {
				t.Fatal()
			}
			for j, fd := range typ.Fields {
				wantFd := wantType.Fields[j]
				assertAnnotations(t, wantFd.Group, fd.Group)
			}
		}
	}
}

func assertAnnotations[N ast.Node](t *testing.T, wantAns annotation.Group[N], dst annotation.Group[N]) {
	want := wantAns.List
	if len(want) != len(dst.List) {
		t.Fatal()
	}
	for i, an := range want {
		val := dst.List[i]
		assert.Equal(t, an.Value, val.Value)
	}
}
