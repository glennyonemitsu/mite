package main

type Node struct {
	tag string
	attrs map[string]string
	text string

	// dumb appended string while attribute assignment is still being determined
	// this is used if it turns out tokens are NOT part of an attribute assignment
	// and the string (with whitespace if any) gets used as the node's text
	attrString string
}

