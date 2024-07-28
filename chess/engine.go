package chess

import (
	"fmt"
	"time"
)

func MakeMove(p *Position, treeDepth int) ([]*Move, float32) {
	/**
	Returns sequence of best moves
	*/
	start := time.Now()
	parent := Node{
		parent:         nil,
		children:       nil,
		move:           nil,
		bestChild:      nil,
		treeNodesCount: 1,
		treeEvaluation: p.evaluation,
		posEvaluation:  p.evaluation,
	}

	AddMovesToTree(&parent, p, treeDepth)
	if parent.bestChild == nil {
		return nil, 0
	}

	if Debug {
		fmt.Println(" --- Printing Best Child --- ")
		PrintNode(parent.bestChild, p)
	}
	bestMoves := make([]*Move, 0)
	for currChild := &parent; currChild != nil; currChild = currChild.bestChild {
		if currChild.move != nil {
			bestMoves = append(bestMoves, currChild.move)
		}
	}
	took := time.Since(start).Seconds()
	println("Tree size:", parent.treeNodesCount, ", took: ", int(took), ", speed=", int(float64(parent.treeNodesCount)/(1000*took)), "Knodes/sec")
	return bestMoves, parent.treeEvaluation
}

func generateNextMovePositions(rootPosition *Position, parent *Node) []*Node {
	p := GetPosition(parent, rootPosition)
	moves := p.GetAllMoves()
	if len(moves) == 0 {
		return make([]*Node, 0)
	}

	//for _, move := range moves {
	//	if strings.Contains(move.String(), "(w) d4xd5 ") {
	//		p.PrintPosition()
	//		println("ok", move.String())
	//	}
	//}

	p.availableMoves = moves

	positions := p.applyMoves(moves)

	var nodes = make([]*Node, len(positions))
	for i, pos := range positions {
		pos.Evaluate(p, moves[i])
		if Debug {
			fmt.Printf("Debug: %s, %f\n", moves[i].String(), pos.evaluation)
		}
		nodes[i] = &Node{
			parent:         parent,
			children:       nil,
			move:           &moves[i],
			treeNodesCount: 1,
			treeEvaluation: pos.evaluation,
			posEvaluation:  pos.evaluation,
		}
	}
	return nodes
}

func AddMovesToTree(parent *Node, rootPosition *Position, movesToAdd int) {
	if movesToAdd == 0 {
		return
	}
	addNextMoveToTree(parent, rootPosition)
	for _, c := range parent.children {
		AddMovesToTree(c, rootPosition, movesToAdd-1)
	}
}

func addNextMoveToTree(parent *Node, rootPosition *Position) {
	nodes := generateNextMovePositions(rootPosition, parent)
	parent.children = nodes
	UpdateParentValue(parent, func(node *Node) {
		s := 0
		for _, c := range node.children {
			s += c.treeNodesCount
		}
		node.treeNodesCount = s
	})

	UpdateParentValue(parent, func(node *Node) {

		if len(node.children) == 0 {
			return
		}

		node.treeEvaluation = node.children[0].treeEvaluation

		for _, c := range node.children {
			if c.move.isWhite {
				if c.treeEvaluation >= node.treeEvaluation {
					node.treeEvaluation = c.treeEvaluation
					node.bestChild = c
				}
			} else {
				if c.treeEvaluation <= node.treeEvaluation {

					node.treeEvaluation = c.treeEvaluation
					node.bestChild = c
				}
			}
		}
	})

}
