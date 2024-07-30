package chess

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

const TreeDepth = 2
const Debug = false
const Moves = 20

// RandomSeed 0 = random
const RandomSeed = 0

type Game struct {
	initPosition     Position
	position         Position
	moves            []Move
	bestMoveSequence []*Move

	isFinished bool
	result     int

	treeDepth int
}

var g GameOperations = &Game{}

type GameOperations interface {
	InitGame(board *[8][8]string, moveWhite bool, treeDepth int)
	MakeMove()
}

func (g *Game) InitGame(board *[8][8]string, moveWhite bool, treeDepth int) {
	s := int64(RandomSeed)
	if RandomSeed == 0 {
		s = time.Now().UnixNano()
	}
	rand.Seed(s)
	log.Println("Random seed: ", s)
	InitZobrist()

	var positionStam PositionOperations = &Position{}
	position := positionStam.InitPosition(board, 1, moveWhite)
	g.initPosition = *position
	g.position = *position
	g.treeDepth = treeDepth
}

func (g *Game) GetLastMove() *Move {
	if len(g.moves) == 0 {
		return nil
	}
	return &g.moves[len(g.moves)-1]
}

func (g *Game) MakeMove() {
	moveSequence, eval := MakeMove(&g.position, g.treeDepth, g)
	if IsCheckmateEvaluation(eval) {
		g.isFinished = true
		g.result = 1 * int(ColorFactor(g.position.whiteTurn))
	}

	if moveSequence != nil && len(moveSequence) > 0 {
		newPosition := ApplyMove(g.position, *moveSequence[0])
		g.moves = append(g.moves, *moveSequence[0])
		g.position = *newPosition
		g.position.evaluation = eval
		g.bestMoveSequence = moveSequence
	} else if !g.isFinished {
		g.isFinished = true
		g.result = 0
	}

}

func StartGame() {
	newGame := NewGame()

	for i := 0; i < Moves; i++ {
		if newGame.isFinished {
			break
		}
		start := time.Now()
		newGame.MakeMove()
		move := newGame.GetLastMove()
		if move == nil {
			newGame.position.PrintPosition()
			break
		}
		fmt.Printf("%d. %s, took:%f secs\n", i+1, move, time.Since(start).Seconds())
		fmt.Print("Best Sequence: ")
		for _, bestMove := range newGame.bestMoveSequence {
			fmt.Printf("%s, ", bestMove)
		}
		fmt.Println("")

		newGame.position.PrintPosition()

	}

	println("Finished")

}

func NewGame() *Game {
	var newGame = new(Game)
	var initB = [8][8]string{
		{"r", "n", "b", "q", "k", "b", "n", "r"},
		{"p", "p", "p", "p", "p", "p", "p", "p"},
		{"", "", "", "", "", "", "", ""},
		{"", "", "", "", "", "", "", ""},
		{"", "", "", "", "", "", "", ""},
		{"", "", "", "", "", "", "", ""},
		{"P", "P", "P", "P", "P", "P", "P", "P"},
		{"R", "N", "B", "Q", "K", "B", "N", "R"},
	}

	//var initB = [8][8]string{
	//	{"", "", "", "", "", "", "", ""},
	//	{"", "", "", "", "", "", "", ""},
	//	{"", "", "", "", "", "", "", ""},
	//	{"", "", "", "", "", "", "", ""},
	//	{"", "", "", "", "", "b", "", ""},
	//	{"", "", "", "", "", "", "", ""},
	//	{"", "", "", "Q", "", "", "", ""},
	//	{"", "", "", "", "", "", "", ""},
	//}

	isWhite := true
	newGame.InitGame(&initB, isWhite, TreeDepth)
	return newGame
}
