package chess

import (
	"fmt"
	"log"
	"strings"
	"time"
)

func MakeMove(treeDepth int, game *Game) ([]*Move, float32) {
	/**
	Returns sequence of best moves
	*/
	p := game.position
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

	game.MinimaxTree(&parent, p, treeDepth, 0, 0)
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

func generateNextMovePositions(p *Position, parent *Node) ([]*Node, []Position) {
	moves := p.GetAllMoves()
	if len(moves) == 0 {
		return make([]*Node, 0), make([]Position, 0)
	}

	p.availableMoves = moves

	positions := p.applyMoves(moves)

	var nodes = make([]*Node, len(positions))
	for i, pos := range positions {
		if Debug {
			fmt.Printf("Debug: %s, %f\n", moves[i].String(), pos.evaluation)
		}
		nodes[i] = &Node{
			parent:         parent,
			children:       nil,
			move:           &moves[i],
			treeNodesCount: 1,
			treeEvaluation: 0,
			posEvaluation:  0,
		}
	}
	return nodes, positions
}

func (g *Game) MinimaxTree(currNode *Node, currPosition *Position, depth int, alpha, beta float32) {

	if Debug {
		log.Println("Current Position")
		currPosition.PrintPosition()
	}

	if depth == 0 {
		return
	}

	nodes, positions := generateNextMovePositions(currPosition, currNode)
	if len(nodes) == 0 {
		return
	}

	//Node evaluation is the evaluation of its best child (max for white and min for black)
	//dfs
	for i, childNode := range nodes {
		move := childNode.move

		childPosition := &positions[i]

		if strings.Contains(move.String(), "e1-e2") {
			log.Println("Debug xxx")
			childPosition.PrintPosition()
		}

		eval := childPosition.Evaluate(currPosition, move, g.positionHashes)
		//todo remove posEvaluation attribute
		childNode.posEvaluation = eval
		childNode.treeEvaluation = eval
		currNode.children = append(currNode.children, childNode)

		if i == 0 || (move.isWhite && childNode.treeEvaluation >= currNode.treeEvaluation) ||
			(!move.isWhite && childNode.treeEvaluation <= currNode.treeEvaluation) {
			currNode.bestChild = childNode
			currNode.treeEvaluation = childNode.treeEvaluation

			//todo make it more efficient
			updateParentEvaluations(currNode)
		}

		if Debug {
			evalStr := fmt.Sprintf("%.2f", childNode.treeEvaluation)
			log.Println("Move: ", ToStringWithParents(childNode), ", eval: ", evalStr)
		}

		// if the game is finished, or reached max depth no need to check the children
		if Abs(eval) == Abs(GetCheckmateEvaluation(true)) || Abs(eval) == Abs(ThreeFoldRepetitionEvalution) {
			continue
		}

		g.MinimaxTree(childNode, childPosition, depth-1, alpha, beta)
	}
	currNode.treeNodesCount = len(nodes)

	//todo fix or remove to support pruning
	UpdateParentValue(currNode, func(node *Node) {
		s := 0
		for _, c := range node.children {
			s += c.treeNodesCount
		}
		node.treeNodesCount = s
	})

}

func updateParentEvaluations(node *Node) {
	UpdateParentValue(node, func(node *Node) {

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
