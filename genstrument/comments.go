package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

const (
	commentPrefix = "+genstrument:"
)

func (l *loader) toFunctionConfig(cg *ast.CommentGroup) (cfg FunctionConfig, ok bool) {
	comments := extractDocComments(cg)
	if len(comments) == 0 {
		return
	}
	cfg.AttributeFunctions = make(map[string]*AttributeKeyFunc)
	for _, comment := range comments {
		if comment.Text == "wrap" {
			ok = true
			continue
		}
		if strings.HasPrefix(comment.Text, "external ") {
			var err error
			cfg.ExternalType, err = l.parseSelectorExpr(strings.TrimPrefix(comment.Text, "external "))
			if err != nil {
				l.recordError(comment.Pos, err)
			}
			continue
		}
		if strings.HasPrefix(comment.Text, "prefix ") {
			cfg.Prefix = strings.TrimPrefix(comment.Text, "prefix ")
			continue
		}
		if strings.HasPrefix(comment.Text, "attr ") {
			attr := strings.TrimPrefix(comment.Text, "attr ")
			keyargfun := strings.SplitN(attr, " ", 3)
			switch len(keyargfun) {
			case 2:
			case 3:
			default:
				l.recordError(comment.Pos, fmt.Errorf("attr: expected 2 or 3 arguments, got %d", len(keyargfun)))
				continue
			}
			keyname := strings.TrimSpace(keyargfun[0])
			argName := strings.TrimSpace(keyargfun[1])
			cfg.AttributeFunctions[argName] = &AttributeKeyFunc{
				Key: keyname,
			}
			if len(keyargfun) == 3 {
				funName := strings.TrimSpace(keyargfun[2])
				pkgType := strings.SplitN(strings.TrimSpace(funName), ".", 2)
				switch len(pkgType) {
				case 1:
					cfg.AttributeFunctions[argName].Func = &ast.Ident{Name: funName}
				case 2:
					cfg.AttributeFunctions[argName].Func = &ast.SelectorExpr{
						X:   &ast.Ident{Name: pkgType[0]},
						Sel: &ast.Ident{Name: pkgType[1]},
					}
				}
			}
			continue
		}
		if strings.HasPrefix(comment.Text, "op ") {
			cfg.OperationName = strings.TrimPrefix(comment.Text, "op ")
			continue
		}
		l.recordError(comment.Pos, fmt.Errorf("unknown function comment: %s", comment.Text))
	}
	return
}

func (l *loader) toInterfaceConfig(cg *ast.CommentGroup) (cfg InterfaceConfig, ok bool) {
	comments := extractDocComments(cg)
	if len(comments) == 0 {
		return
	}
	for _, comment := range comments {
		if comment.Text == "wrap" {
			ok = true
			continue
		}
		if strings.HasPrefix(comment.Text, "prefix ") {
			cfg.Prefix = strings.TrimPrefix(comment.Text, "prefix ")
			continue
		}
		if strings.HasPrefix(comment.Text, "external ") {
			var err error
			cfg.ExternalType, err = l.parseSelectorExpr(strings.TrimPrefix(comment.Text, "external "))
			if err != nil {
				l.recordError(comment.Pos, err)
			}
			continue
		}
		if strings.HasPrefix(comment.Text, "constructor ") {
			cfg.ConstructorPrefix = strings.TrimPrefix(comment.Text, "constructor ")
			continue
		}
		l.recordError(comment.Pos, fmt.Errorf("unknown interface comment: %s", comment.Text))
	}
	return
}

type Comment struct {
	Text string
	Pos  token.Pos
}

func (l *loader) parseSelectorExpr(expr string) (*ast.SelectorExpr, error) {
	pkgType := strings.SplitN(strings.TrimSpace(expr), ".", 2)
	if len(pkgType) != 2 {
		return nil, fmt.Errorf("invalid selector expression '%s': expected 'package.TypeName'", expr)
	}
	return &ast.SelectorExpr{
		X:   &ast.Ident{Name: pkgType[0]},
		Sel: &ast.Ident{Name: pkgType[1]},
	}, nil
}

func extractDocComments(cg *ast.CommentGroup) []Comment {
	if cg == nil {
		// only documented interfaces are considered
		return nil
	}
	var comments []Comment
	for _, comment := range cg.List {
		text := strings.TrimPrefix(comment.Text, "//")
		text = strings.TrimSpace(text)
		if !strings.HasPrefix(text, commentPrefix) {
			continue // not a doc comment
		}
		text = strings.TrimPrefix(text, commentPrefix)
		comments = append(comments, Comment{
			Text: text,
			Pos:  comment.Pos(),
		})
	}
	return comments
}
