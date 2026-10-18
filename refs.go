package main

import (
	"text/template"
	"text/template/parse"
)

type templateRef struct {
	Root  string
	Field string
}

func collectRefs(t *template.Template) []templateRef {
	var refs []templateRef
	visit := func(ident []string) {
		r := templateRef{Root: ident[0]}
		if len(ident) > 1 {
			r.Field = ident[1]
		}
		refs = append(refs, r)
	}
	for _, tt := range t.Templates() {
		if tt.Tree != nil {
			walk(tt.Tree.Root, true, visit)
		}
	}
	return refs
}

func walk(n parse.Node, atRoot bool, visit func([]string)) {
	switch n := n.(type) {
	case *parse.ListNode:
		if n == nil {
			return
		}
		for _, c := range n.Nodes {
			walk(c, atRoot, visit)
		}
	case *parse.ActionNode:
		walk(n.Pipe, atRoot, visit)
	case *parse.PipeNode:
		if n == nil {
			return
		}
		for _, c := range n.Cmds {
			walk(c, atRoot, visit)
		}
	case *parse.CommandNode:
		for _, a := range n.Args {
			walk(a, atRoot, visit)
		}
	case *parse.ChainNode:
		walk(n.Node, atRoot, visit)
	case *parse.FieldNode:
		if atRoot {
			visit(n.Ident)
		}
	case *parse.VariableNode:
		if n.Ident[0] == "$" && len(n.Ident) > 1 {
			visit(n.Ident[1:])
		}
	case *parse.IfNode:
		walk(n.Pipe, atRoot, visit)
		walk(n.List, atRoot, visit)
		walk(n.ElseList, atRoot, visit)
	case *parse.RangeNode:
		walk(n.Pipe, atRoot, visit)
		walk(n.List, false, visit)
		walk(n.ElseList, atRoot, visit)
	case *parse.WithNode:
		walk(n.Pipe, atRoot, visit)
		walk(n.List, false, visit)
		walk(n.ElseList, atRoot, visit)
	case *parse.TemplateNode:
		walk(n.Pipe, atRoot, visit)
	}
}
