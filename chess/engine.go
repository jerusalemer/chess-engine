package chess

import (
	"fmt"
	"log"
	"math"
	"sort"
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
		treeEvaluation: GetWorstEvaluation(p.whiteTurn),
	}

	game.MinimaxTree(&parent, p, treeDepth, -math.MaxFloat32, math.MaxFloat32)
	if parent.bestChild == nil {
		panic("Didn't find best move")
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

	positionMoves := p.applyMoves(moves)
	sortPositions(positionMoves)

	var nodes = make([]*Node, len(positionMoves))
	var positions = make([]Position, len(positionMoves))

	for i, posMove := range positionMoves {
		if Debug {
			posMove.pos.PrintPosition()
			fmt.Printf("Debug: %s\n", posMove.move.String())
			posMove.pos.PrintPosition()
		}
		eval := GetWorstEvaluation(posMove.pos.whiteTurn)
		nodes[i] = &Node{
			parent:         parent,
			children:       nil,
			move:           posMove.move,
			treeNodesCount: 1,
			treeEvaluation: eval,
		}
		positions[i] = *posMove.pos
	}
	return nodes, positions
}

func GetWorstEvaluation(whiteTurn bool) float32 {
	eval := float32(math.MaxFloat32)
	if whiteTurn {
		eval = -math.MaxFloat32
	}
	return eval
}

type PositionMove struct {
	pos  *Position
	move *Move
}

func (m Move) toInt() int {
	captureInt := 0
	if m.isCapture {
		captureInt = 1
	}
	return int(m.fromRow) + int(m.fromCol)*8 + int(m.toRow)*8*8 + int(m.toCol)*8*8*8 + captureInt*8*8*8*8 + int(m.pawnPromotePiece)*8*8*8*8*8
}

func sortPositions(positionMoves []PositionMove) {
	sort.SliceStable(positionMoves, func(i, j int) bool {
		moveI := positionMoves[i].move
		moveJ := positionMoves[j].move
		if moveI.toInt() > moveJ.toInt() {
			return true
		}
		return false
	})

}

func (g *Game) MinimaxTree(currNode *Node, currPosition *Position, depth int, lowerBoundEval, upperBoundEval float32) {

	if Debug {
		log.Println("Current Position")
		currPosition.PrintPosition()
	}

	if depth == 0 {
		eval := currPosition.Evaluate(currPosition, currNode.move, g.positionHashes)
		currNode.treeEvaluation = eval
		return
	}

	nodes, positions := generateNextMovePositions(currPosition, currNode)
	if len(nodes) == 0 {
		eval := currPosition.Evaluate(currPosition, currNode.move, g.positionHashes)
		currNode.treeEvaluation = eval
		return
	}

	//Node evaluation is the evaluation of its best child (max for white and min for black)
	//dfs
	for i, childNode := range nodes {
		move := childNode.move

		childPosition := &positions[i]

		childPosition.hash = UpdateZobristHash(currPosition.hash, move, currPosition)
		if isThreeFoldRepetition(childPosition, g.positionHashes) {
			childNode.treeEvaluation = ColorFactor(move.isWhite) * ThreeFoldRepetitionEvalution
		} else {
			g.MinimaxTree(childNode, childPosition, depth-1, lowerBoundEval, upperBoundEval)
		}

		if Debug {
			evalStr := fmt.Sprintf("%.2f", childNode.treeEvaluation)
			log.Println("MinimaxTree: Move: ", ToStringWithParents(childNode), ", eval: ", evalStr, ", alpha: ", fmt.Sprintf("%.2f", lowerBoundEval), ", beta: ", fmt.Sprintf("%.2f", upperBoundEval))
		}

		currNode.children = append(currNode.children, childNode)
		eval := childNode.treeEvaluation
		parentEval := currNode.treeEvaluation

		if move.isWhite && eval > parentEval {
			currNode.bestChild = childNode
			currNode.treeEvaluation = eval
			//todo make it more efficient
			//updateParentEvaluations(currNode)

			if eval > lowerBoundEval {
				lowerBoundEval = eval
			}
			if eval >= upperBoundEval {
				break
			}
		}

		if !move.isWhite && eval <= parentEval {
			currNode.bestChild = childNode
			currNode.treeEvaluation = eval
			//todo make it more efficient
			//updateParentEvaluations(currNode)

			if eval < upperBoundEval {
				upperBoundEval = eval
			}
			if eval <= lowerBoundEval {
				break
			}
		}

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
