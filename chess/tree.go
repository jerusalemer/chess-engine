package chess

import (
	"fmt"
	"golang.org/x/exp/slices"
)

type NodeFunc func(*Node)

type Node struct {
	parent   *Node
	children []*Node

	move           *Move
	bestChild      *Node
	treeNodesCount int
	treeEvaluation float32
	posEvaluation  float32
}

func (n *Node) String() string {
	return fmt.Sprintf("Move: %s, treeNodesCount:%d, treeEvaluation:%f, posEvaluation:%f", n.move.String(), n.treeNodesCount, n.treeEvaluation, n.posEvaluation)
}

func PrintNode(n *Node, rootPos *Position) {
	p := GetPosition(n, rootPos)
	p.PrintPosition()
	fmt.Printf("Node: %s\n", n.String())
}

func AppendNode(parent *Node, position *Position, move *Move) {
	node := &Node{
		parent:         parent,
		children:       []*Node{},
		move:           move,
		treeNodesCount: 0,
	}
	parent.children = append(parent.children, node)
}

func RemoveNodes(parent *Node, nodes []*Node) {
	var newChildren []*Node
	for _, node := range parent.children {
		if !slices.Contains(nodes, node) {
			newChildren = append(newChildren, node)
		}
	}
	parent.children = newChildren
}

func GetPosition(node *Node, position *Position) *Position {
	newPos := ClonePosition(position)
	parents := GetParentNodes(node)
	for i := len(parents) - 1; i >= 0; i-- {
		if parents[i].move != nil {
			ApplyMovePointers(newPos, parents[i].move)
		}
	}
	newPos.evaluation = node.treeEvaluation
	if newPos.evaluation == GetCheckmateEvaluation(newPos.whiteTurn) {
		newPos.isCheckmate = true
	}
	return newPos
}

func GetParentNodes(node *Node) []*Node {
	parents := make([]*Node, 0)
	for n := node; n != nil; n = n.parent {
		parents = append(parents, n)
	}
	return parents

}

func UpdateParentValue(parent *Node, nodeFunc NodeFunc) {
	if parent == nil {
		return
	}
	nodeFunc(parent)
	UpdateParentValue(parent.parent, nodeFunc)
}

func Print(root *Node) {
	if root == nil {
		return
	}

	queue := []*Node{root}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.move != nil {
			fmt.Printf("Move: %s, Evaluation: %.2f, whiteTurn=%t\n", current.move, current.treeEvaluation, current.move.isWhite)
		} else {
			fmt.Printf("Move: nil, Evaluation: %.2f\n", current.treeEvaluation)
		}

		// Enqueue the children of the current node
		for _, child := range current.children {
			queue = append(queue, child)
		}
	}
}
