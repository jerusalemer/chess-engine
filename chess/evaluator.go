package chess

import (
	"math"
	"math/rand"
)

const HighestPositionScore = math.MaxFloat32
const LowestPositionScore = -math.MaxFloat32

const ThreeFoldRepetitionEvalution = -1.900128

const AvailableMovesFactor = float32(0.01)
const AttackingMovesFactor = float32(0.02)

var pieceCost = map[uint8]float32{
	PawnBit:   1,
	KnightBit: 3,
	BishopBit: 3,
	RookBit:   5,
	QueenBit:  9,
	KingBit:   0,
}

var allPieceIndexes = map[uint8]uint8{
	PawnBit:   1,
	KnightBit: 2,
	BishopBit: 3,
	RookBit:   4,
	QueenBit:  5,
	KingBit:   0,
}

var pieceSquares = map[byte][8][8]float32{
	PawnBit: {
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{-0.1, -0.1, 0, 0.15, 0.15, 0, -0.05, -0.05},
		{0, 0, 0, 0.15, 0.15, 0, 0, 0},
		{0.1, 0.1, 0.15, 0.15, 0.15, 0.15, 0.1, 0.1},
		{0.2, 0.2, 0.3, 0.3, 0.3, 0.3, 0.2, 0.2},
		{0, 0, 0, 0, 0, 0, 0, 0},
	},
	KnightBit: {
		{-0.05, 0, 0, 0, 0, 0, 0, -0.05},
		{-0.05, 0, 0, 0, 0, 0, 0, -0.05},
		{-0.03, 0, 0.1, 0.1, 0.1, 0.1, 0, -0.03},
		{0, 0, 0, 0.13, 0.13, 0, 0, 0},
		{0, 0, 0, 0.13, 0.13, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	},
	BishopBit: {
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05, 0},
		{0.05, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05},
		{0, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05, 0},
		{0, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	},
	RookBit: {
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	},
	QueenBit: {
		{0, 0, 0, 0.05, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	},
	KingBit: {
		{0, 0.5, 0.5, -0.2, 0, -0.2, 0.5, 0},
		{0, 0, -0.5, -0.5, -0.5, -0.5, 0, 0},
		{-0.5, -0.5, -0.5, -0.5, -0.5, -0.5, -0.5, -0.5},
		{-0.5, -0.5, -0.5, -0.5, -0.5, -0.5, -0.5, -0.5},
		{-0.5, -0.5, -0.5, -0.5, -0.5, -0.5, -0.5, -0.5},
		{-0.5, -0.5, -0.5, -0.5, -0.5, -0.5, -0.5, -0.5},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	},
}

func getAttackingMoves(p *Position, possibleMoves []Move) int {
	cnt := 0
	for _, m := range possibleMoves {
		if m.isCapture {
			cnt += 1
		}
	}
	return cnt
}

func (p *Position) Evaluate(prevPos *Position, move *Move, positionHashes map[uint64]bool) float32 {

	p.hash = UpdateZobristHash(prevPos.hash, move, prevPos)

	if isThreeFoldRepetition(p, positionHashes) {
		p.evaluation = ColorFactor(move.isWhite) * ThreeFoldRepetitionEvalution
		return p.evaluation
	}

	possibleMoves := p.GetAllMoves()
	possibleAttackingMoves := getAttackingMoves(p, possibleMoves)

	if len(possibleMoves) == 0 {
		if isKingAttacked(p, p.whiteTurn) {
			p.isCheckmate = true
			p.evaluation = GetCheckmateEvaluation(p.whiteTurn)

			return p.evaluation
		} else {
			p.evaluation = 0
			return 0
		}
	}

	eval := countMaterial(p)

	eval += ColorFactor(p.whiteTurn) * AvailableMovesFactor * float32(len(possibleMoves)-len(prevPos.availableMoves))
	eval += ColorFactor(p.whiteTurn) * AttackingMovesFactor * float32(possibleAttackingMoves-getAttackingMoves(prevPos, prevPos.availableMoves))

	//add a random value to evaluation to make the game less predictable, otherwise the same games keep occurring
	eval += ColorFactor(p.whiteTurn) * rand.Float32() * 0.2

	p.evaluation = eval

	return eval
}

// checks whether the hash of the position already exists in position hashes
func isThreeFoldRepetition(pos *Position, positionHashes map[uint64]bool) bool {
	_, exists := positionHashes[pos.hash]

	return exists
}

func countMaterial(p *Position) float32 {
	res := float32(0.0)
	for i := uint8(0); i < 8; i++ {
		for j := uint8(0); j < 8; j++ {
			piece, isWhite := getPiece(i, j, p)

			res += pieceCost[piece] * ColorFactor(isWhite)
			if isWhite {
				res += ColorFactor(isWhite) * pieceSquares[piece][i][j]
			} else {
				res += ColorFactor(isWhite) * pieceSquares[piece][7-i][j]
			}
		}
	}
	return res
}

func ColorFactor(isWhite bool) float32 {
	return float32(ColorFactorInt(isWhite))
}

func ColorFactorInt(isWhite bool) int8 {
	color := 1
	if !isWhite {
		color = -1
	}
	return int8(color)
}

func Abs(num float32) float32 {
	if num >= 0 {
		return num
	} else {
		return -num
	}
}

func GetCheckmateEvaluation(whiteTurn bool) float32 {
	return HighestPositionScore * ColorFactor(whiteTurn) * (-1)
}

func IsCheckmateEvaluation(evaluation float32) bool {
	return HighestPositionScore == math.Abs(float64(evaluation))
}
