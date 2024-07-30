package chess

import (
	"fmt"
	"log"
	"time"
)

func MakeMove(p *Position, treeDepth int, game *Game) ([]*Move, float32) {
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

	AddMovesToTree(&parent, p, treeDepth, game)
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
	evalStr := fmt.Sprintf("%.2f", parent.treeEvaluation)
	log.Println("Eval:", evalStr, "Tree size:", parent.treeNodesCount, ", took: ", int(took), ", speed=", int(float64(parent.treeNodesCount)/(1000*took)), "Knodes/sec")
	return bestMoves, parent.treeEvaluation
}

func generateNextMovePositions(rootPosition *Position, parent *Node, positionHashes map[uint64]bool) []*Node {
	p := GetPosition(parent, rootPosition)
	moves := p.GetAllMoves()
	if len(moves) == 0 {
		return make([]*Node, 0)
	}

	p.availableMoves = moves

	positions := p.applyMoves(moves)

	var nodes = make([]*Node, len(positions))
	for i, pos := range positions {
		pos.Evaluate(p, &moves[i], positionHashes)
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

func AddMovesToTree(parent *Node, rootPosition *Position, movesToAdd int, game *Game) {
	if Abs(parent.posEvaluation) == GetCheckmateEvaluation(true) || Abs(parent.posEvaluation) == Abs(ThreeFoldRepetitionEvalution) {
		return
	}
	if movesToAdd == 0 {
		return
	}
	addNextMoveToTree(parent, rootPosition, game.positionHashes)
	for _, c := range parent.children {
		AddMovesToTree(c, rootPosition, movesToAdd-1, game)
	}
}

func addNextMoveToTree(parent *Node, rootPosition *Position, positionHashes map[uint64]bool) {
	nodes := generateNextMovePositions(rootPosition, parent, positionHashes)
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
