package annotation

import (
	"go/ast"
	"strings"
)

type Single struct {
	Key   string
	Value string
}

type Group[N ast.Node] struct {
	Node N
	List []Single
}

type ASTField struct {
	Group[*ast.Field]
}

type ASTType struct {
	Group[*ast.TypeSpec]
	Fields []ASTField
}

type ASTFile struct {
	Group[*ast.File]
	Types []ASTType
}

func FromComment(c *ast.Comment) (Single, bool) {
	txt := c.Text
	if strings.HasPrefix(txt, "// ") {
		txt = txt[3:]
	} else if strings.HasPrefix(txt, "/* ") {
		txt = txt[3 : len(txt)-2]
	} else {
		return Single{}, false
	}

	// todo: multiple annotation case
	index := strings.Index(txt, "@")
	if index < 0 {
		return Single{}, false
	}
	txt = txt[index+1:]
	sep := strings.SplitN(txt, " ", 2)

	key, val := sep[0], ""
	if len(sep) > 1 {
		val = sep[1]
	}
	return Single{key, val}, true
}

func FromCommentGroup[N ast.Node](node N, cg *ast.CommentGroup) Group[N] {
	if cg == nil || len(cg.List) == 0 {
		return Group[N]{Node: node}
	}
	ans := make([]Single, 0, len(cg.List))
	for _, c := range cg.List {
		annotation, ok := FromComment(c)
		if !ok {
			continue
		}
		ans = append(ans, annotation)
	}
	return Group[N]{Node: node, List: ans}
}
