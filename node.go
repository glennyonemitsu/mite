package main

import (
	"fmt"
)

type NodeType int

const (
	NodeNil NodeType = iota
	NodeRoot
	NodeTag
	NodeText
	NodeComment
	NodeDoctype
	NodeFunc
)

var NodeTypeString = map[NodeType]string {
	NodeNil:		"Nil",
	NodeRoot:		"Root",
	NodeTag:		"Tag",
	NodeText:		"Text",
	NodeComment:	"Comment",
	NodeDoctype:	"Doctype",
	NodeFunc:	"Function",
}

type Node struct {
	Type NodeType

	// were tags printed out? important to know if NodeText follows and needs
	// to append more text
	opened bool
	closed bool

	tag string
	attrs map[string]string
	text string

	// dumb appended string while attribute assignment is still being determined
	// this is used if it turns out tokens are NOT part of an attribute assignment
	// and the string (with whitespace if any) gets used as the node's text
	attrString string
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
		output += n.attrString
		output += n.text
		n.attrString = ""
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
	output += fmt.Sprintf("[tag:%s]", n.tag)
	output += fmt.Sprintf("[attrs:%s]", n.attrs)
	output += fmt.Sprintf("[text(%d):%s]", len(n.text), n.text)
	output += fmt.Sprintf("[attrString:%s]", n.attrString)
	return output
}
