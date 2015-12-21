package elecciones

import (
	"fmt"
	"strings"
)

type Pais []string

func (p Pais) ParentID() string {
	return p[2]
}

func (p Pais) Name() string {
	return p[1]
}

func (p Pais) ID() string {
	return p[0]
}

type Paises []Pais

type Comunidad []string

func (p Comunidad) ParentID() string {
	return p[2]
}

func (p Comunidad) Name() string {
	return p[1]
}

func (p Comunidad) ID() string {
	return p[0]
}

type Comunidades []Comunidad

type Provincia []string

func (p Provincia) ParentID() string {
	return p[2]
}

func (p Provincia) Name() string {
	return p[1]
}

func (p Provincia) ID() string {
	return p[0]
}

type Provincias []Provincia

type Isla []string

func (p Isla) ParentID() string {
	return p[2]
}

func (p Isla) Name() string {
	return p[1]
}

func (p Isla) ID() string {
	return p[0]
}

type Islas []Isla

type Municipio []string

func (p Municipio) ParentID() string {
	return p[2]
}

func (p Municipio) Name() string {
	return p[1]
}

func (p Municipio) ID() string {
	return p[0]
}

type Municipios []Municipio

type Distrito []string

func (p Distrito) ParentID() string {
	return p[2]
}

func (p Distrito) Name() string {
	return p[1]
}

func (p Distrito) ID() string {
	return p[0]
}

type Distritos []Distrito

type RawNode interface {
	Name() string
	ID() string
	ParentID() string
}

type Node struct {
	data     RawNode
	parent   *Node
	children []*Node
}

func NewNode(node RawNode, parent *Node) (*Node, error) {
	return &Node{
		data:   node,
		parent: parent,
	}, nil
}

func (n *Node) AddChild(node *Node) {
	n.children = append(n.children, node)
}

func (n Node) Parent() *Node {
	return n.parent
}

func (n Node) Children() []*Node {
	return n.children
}

func (n Node) Path() string {
	items := []string{n.data.ID()}

	for s := n.Parent(); s != nil; s = s.Parent() {
		items = append(items, s.data.ID())
	}
	items = reverseStringSlice(items)

	return strings.Join(items, "/")
}

func (n Node) URL() string {
	return fmt.Sprintf(urlFormat, n.Path())
}