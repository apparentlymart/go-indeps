package indeps

import (
	"fmt"
)

type Graph struct {
	nodes    map[Node]existence
	edgesOut map[Node]map[Node]existence
	edgesIn  map[Node]map[Node]existence
}

type Edge struct {
	From Node
	To   Node
}

func NewGraph() *Graph {
	return &Graph{
		nodes:    map[Node]existence{},
		edgesOut: map[Node]map[Node]existence{},
		edgesIn:  map[Node]map[Node]existence{},
	}
}

type existence struct{}

var exists = existence{}

// AddNode adds a node to the graph with no vertices
func (g *Graph) AddNode(n Node) {
	if _, exists := g.nodes[n]; exists {
		return
	}
	g.nodes[n] = exists
	g.edgesOut[n] = map[Node]existence{}
	g.edgesIn[n] = map[Node]existence{}
}

// RemoveNode removes a node from the graph, along with all of its vertices
func (g *Graph) RemoveNode(n Node) {
	delete(g.nodes, n)
	delete(g.edgesIn, n)
	delete(g.edgesOut, n)
}

// AddEdge add a directed edge between the two given nodes, and also adds
// the nodes to the graph implicitly if they aren't already present.
func (g *Graph) AddEdge(from, to Node) {
	// Self edges are not interesting for this application
	if from == to {
		return
	}

	g.AddNode(from)
	g.AddNode(to)
	g.edgesOut[from][to] = exists
	g.edgesIn[to][from] = exists
}

// RemoveEdge removes a directed edge between the two given nodes, if one
// exists. If it doesn't, this is a no-op.
func (g *Graph) RemoveEdge(from, to Node) {
	delete(g.edgesOut[from], to)
	delete(g.edgesIn[to], from)
}

// RemoveDisconnectedNodes removes any node that doesn't have any in or out
// edges.
func (g *Graph) RemoveDisconnectedNodes() {
	for n := range g.nodes {
		_, hasOut := g.edgesOut[n]
		_, hasIn := g.edgesIn[n]
		if !(hasIn || hasOut) {
			g.RemoveNode(n)
		}
	}
}

// HasNode returns true if the given node is in the graph, or false if it is
// not in the graph.
func (g *Graph) HasNode(n Node) bool {
	_, ret := g.nodes[n]
	return ret
}

// Nodes returns all of the nodes from the graph, in no particular order.
func (g *Graph) Nodes() []Node {
	ret := make([]Node, 0, len(g.nodes))
	for node := range g.nodes {
		ret = append(ret, node)
	}
	return ret
}

// Edges returns all of the edges from the graph, in no particular order.
func (g *Graph) Edges() []Edge {
	ret := make([]Edge, 0)
	for fromNode, targets := range g.edgesOut {
		for toNode := range targets {
			ret = append(ret, Edge{
				From: fromNode,
				To:   toNode,
			})
		}
	}
	return ret
}

// Node is the abstract type of nodes in a Graph. The concrete implementations
// of this type are Function, Type, Constant, and Variable.
type Node interface {
	private() private
}

type private int

type Function string

type Type string

type Constant string

type Variable string

// Node impl
func (n Function) private() private {
	return private(0)
}

func (n Function) String() string {
	return fmt.Sprintf("func %s", string(n))
}

// Node impl
func (n Type) private() private {
	return private(0)
}

func (n Type) String() string {
	return fmt.Sprintf("type %s", string(n))
}

// Node impl
func (n Constant) private() private {
	return private(0)
}

func (n Constant) String() string {
	return fmt.Sprintf("const %s", string(n))
}

// Node impl
func (n Variable) private() private {
	return private(0)
}

func (n Variable) String() string {
	return fmt.Sprintf("var %s", string(n))
}
