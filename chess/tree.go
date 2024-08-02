package chess

import (
	"fmt"
)

type NodeFunc func(*Node)

type Node struct {
	parent   *Node
	children []*Node

	move           *Move
	bestChild      *Node
	treeNodesCount int
	treeEvaluation float32
}

func (n *Node) String() string {
	return fmt.Sprintf("Move: %s, treeNodesCount:%d, treeEvaluation:%f", n.move.String(), n.treeNodesCount, n.treeEvaluation)
}

func PrintNode(n *Node, rootPos *Position) {
	p := GetPosition(n, rootPos)
	p.PrintPosition()
	fmt.Printf("Node: %s\n", n.String())
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

func ToStringWithParents(node *Node) string {
	str := ""
	currNode := node
	for {
		if currNode.move == nil {
			return str
		}
		str = currNode.move.String() + ", " + str
		currNode = currNode.parent
	}
}
