package visitor

import (
	"go/ast"

	"github.com/longyue0521/code-generator/annotation"
)

type SingleFileEntryVisitor struct {
	f *FileVisitor
}

func (s *SingleFileEntryVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch n := node.(type) {
	case *ast.File:
		s.f = &FileVisitor{ans: annotation.FromCommentGroup(n, n.Doc)}
		s.f.visited = true
	}
	return s.f
}

func (s *SingleFileEntryVisitor) Get() annotation.ASTFile {
	return s.f.Get()
}

type FileVisitor struct {
	ans     annotation.Group[*ast.File]
	types   []*typeVisitor
	visited bool
}

func (f *FileVisitor) Visit(node ast.Node) (w ast.Visitor) {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *ast.TypeSpec:
		visitor := &typeVisitor{ans: annotation.FromCommentGroup(n, n.Doc)}
		visitor.Visit(n)
		f.types = append(f.types, visitor)
		return visitor
	}
	return f
}

func (f *FileVisitor) Get() annotation.ASTFile {
	types := make([]annotation.ASTType, 0, len(f.types))
	for _, v := range f.types {
		types = append(types, v.Get())
	}
	return annotation.ASTFile{
		Group: f.ans,
		Types: types,
	}
}

type typeVisitor struct {
	ans    annotation.Group[*ast.TypeSpec]
	fields []annotation.ASTField
}

func (t *typeVisitor) Get() annotation.ASTType {
	return annotation.ASTType{
		Group:  t.ans,
		Fields: t.fields,
	}
}

func (t *typeVisitor) Visit(node ast.Node) (w ast.Visitor) {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *ast.Field:
		if n.Doc != nil {
			t.fields = append(t.fields, annotation.ASTField{Group: annotation.FromCommentGroup[*ast.Field](n, n.Doc)})
		}
	}
	return t
}
