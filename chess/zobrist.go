package chess

import (
	"log"
	"math/rand"
)

const (
	boardSize = 8
	numPieces = 12 // Number of different pieces (6 for white, 6 for black)
	whiteTurn = 1
	blackTurn = 0
)

// Zobrist table
var zobristTable [boardSize][boardSize][numPieces]uint64
var zobristTurn [2]uint64
var zobristUp = false

// InitZobrist Initialize Zobrist table
func InitZobrist() {
	if zobristUp {
		return
	}
	for i := 0; i < boardSize; i++ {
		for j := 0; j < boardSize; j++ {
			for k := 0; k < numPieces; k++ {
				zobristTable[i][j][k] = rand.Uint64()
			}
		}
	}
	zobristTurn[whiteTurn] = rand.Uint64()
	zobristTurn[blackTurn] = rand.Uint64()
	zobristUp = true
}

func ComputeZobristHash(pos *Position) uint64 {
	var hash uint64

	// Hash the board pieces
	for i := uint8(0); i < boardSize; i++ {
		for j := uint8(0); j < boardSize; j++ {
			piece, _ := getPiece(i, j, pos)
			if piece != 0 {
				hash ^= zobristTable[i][j][allPieceIndexes[piece]]
			}
		}
	}

	// Hash the turn
	if pos.whiteTurn {
		hash ^= zobristTurn[whiteTurn]
	} else {
		hash ^= zobristTurn[blackTurn]
	}

	return hash
}

// UpdateZobristHash updates the Zobrist hash based on a given move
func UpdateZobristHash(hash uint64, move *Move, pos *Position) uint64 {
	if move == nil {
		log.Println("Art debug")
	}
	fromPiece, _ := getPiece(move.fromRow, move.fromCol, pos)
	toPiece, _ := getPiece(move.toRow, move.toCol, pos)

	// Remove the piece from the source square
	hash ^= zobristTable[move.fromRow][move.fromCol][allPieceIndexes[fromPiece]]

	// If the move is a capture, remove the captured piece from the target square
	if move.isCapture {
		hash ^= zobristTable[move.toRow][move.toCol][allPieceIndexes[toPiece]]
	}

	// Add the piece to the target square
	if move.pawnPromotePiece != 0 {
		// If it's a pawn promotion, add the promoted piece to the target square
		hash ^= zobristTable[move.toRow][move.toCol][allPieceIndexes[move.pawnPromotePiece]]
	} else {
		// Otherwise, add the moved piece to the target square
		hash ^= zobristTable[move.toRow][move.toCol][allPieceIndexes[fromPiece]]
	}

	// Switch the turn
	if move.isWhite {
		hash ^= zobristTurn[whiteTurn]
		hash ^= zobristTurn[blackTurn]
	} else {
		hash ^= zobristTurn[blackTurn]
		hash ^= zobristTurn[whiteTurn]
	}

	return hash
}
