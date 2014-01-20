package main

import (
	"fmt"
)

type NodeType int

const (
	NodeRoot NodeType = iota
	NodeTag
	NodeText
	NodeComment
	NodeDoctype
	NodeExpression
)

var NodeTypeString = map[NodeType]string {
	NodeRoot:		"Root",
	NodeTag:		"Tag",
	NodeText:		"Text",
	NodeComment:	"Comment",
	NodeDoctype:	"Doctype",
	NodeExpression:	"Expression",
}

func newNode(t NodeType) *Node {
	n := new(Node)
	n.attrs = make(map[string]string)
	n.Type = t
	return n
}

type Node struct {
	Type NodeType
	parent *Node
	children []*Node
	index int // index for Node.parent.children ([]*Node)

	// NodeTag
	tag string
	opened bool
	closed bool
	attrs map[string]string

	// NodeText
	text string

	// NodeExpression
	funcName string
}

func (n *Node) appendChild(c *Node) {
	c.index = len(n.children)
	n.children = append(n.children, c)
}

func (n *Node) prevSibling() *Node {
	if n.parent != nil && n.index > 0 {
		return n.parent.children[n.index-1]
	} else {
		return nil
	}
}

func (n *Node) nextSibling() *Node {
	if n.parent != nil && len(n.parent.children) > (n.index+1) {
		return n.parent.children[n.index+1]
	} else {
		return nil
	}
}

func (n *Node) TypeString() string {
	if nodeType, found := NodeTypeString[n.Type]; found {
		return nodeType
	} else {
		return "???"
	}
}

func (n *Node) OpenString() string {
	output := ""
	if n.Type == NodeTag && !n.opened {
		output += "<" + n.tag
		for k, v := range n.attrs {
			output += fmt.Sprintf(" %s='%s'", k, v)
		}
		output += ">"
		n.opened = true
	}
	if n.Type == NodeTag || n.Type == NodeText {
		output += n.text
		n.text = ""
	}
	return output
}

func (n *Node) CloseString() string {
	output := ""
	if n.Type == NodeTag && !n.closed {
		output += "</" + n.tag + ">"
		n.closed = true
	}
	return output
}

func (n *Node) Debug() string {
	output := ""
	output += fmt.Sprintf("[Type:%s]", n.TypeString())
	output += fmt.Sprintf("[opened:%t]", n.opened)
	output += fmt.Sprintf("[closed:%t]", n.closed)
	output += fmt.Sprintf("[index:%d]", n.index)
	output += fmt.Sprintf("[tag:%s]", n.tag)
	output += fmt.Sprintf("[attrs:%s]", n.attrs)
	output += fmt.Sprintf("[text(%d):%s]", len(n.text), n.text)
	return output
}

func (n *Node) DebugTree() {
	var edge *Node
	var log string
	indent := 0

	edge = n
	for edge != nil {
		log = ""
		for i := 0; i < indent; i += 1 {
			log += "\t"
		}
		log += n.Debug()
		log += "\n"
		println(log)
		// look for first child, then nextSibling, then parent
		if len(edge.children) > 0 {
			edge = edge.children[0]
		} else if n.nextSibling() != nil {
			edge = edge.nextSibling()
		} else {
			edge = edge.parent
		}
	}
}
