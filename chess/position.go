package chess

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"unicode"
)

const PawnBit = byte(1)
const KnightBit = byte(2)
const BishopBit = byte(4)
const RookBit = byte(8)
const QueenBit = byte(16)
const KingBit = byte(32)
const isWhiteBit = byte(64)

var piecesToPromote = []byte{
	KnightBit,
	BishopBit,
	RookBit,
	QueenBit,
}

type Position struct {
	board     [8][8]uint8
	moveNum   int8
	whiteTurn bool

	evaluation  float32
	isCheckmate bool

	whiteShortCastleAllowed bool
	blackShortCastleAllowed bool
	whiteLongCastleAllowed  bool
	blackLongCastleAllowed  bool

	//used for en passant
	whitePawnDoubleStepCol uint8
	blackPawnDoubleStepCol uint8

	//optimizations
	whiteKingPosRow uint8
	whiteKingPosCol uint8
	blackKingPosRow uint8
	blackKingPosCol uint8
	availableMoves  []Move

	hash uint64
}

type Move struct {
	fromRow   uint8
	fromCol   uint8
	toRow     uint8
	toCol     uint8
	isWhite   bool
	isCapture bool

	pawnPromotePiece uint8
	isEnPassant      bool
}

func IsCheckMate(position Position) bool {
	return position.isCheckmate || position.evaluation == GetCheckmateEvaluation(position.whiteTurn)
}

func (m Move) String() string {
	sign := "-"
	if m.isCapture {
		sign = "x"
	}
	color := "w"
	if !m.isWhite {
		color = "b"
	}
	return fmt.Sprintf("(%s) %s%d%s%s%d (%d,%d->%d,%d)", color, string(rune('a'+m.fromCol)), m.fromRow+1, sign, string(rune('a'+m.toCol)), m.toRow+1, m.fromRow, m.fromCol, m.toRow, m.toCol)
}

func printMoves(moves []Move, prefix string) {
	if Debug {
		for i, m := range moves {
			println(i, ". ", prefix, " - ", m.String())
		}
	}
}

func (m Move) Equal(other Move) bool {
	return m.fromRow == other.fromRow &&
		m.fromCol == other.fromCol &&
		m.toRow == other.toRow &&
		m.toCol == other.toCol &&
		m.isWhite == other.isWhite
}

type PositionOperations interface {
	InitPosition(board *[8][8]string, moveNum int8, turnWhite bool) *Position
	PrintPosition()
	IsValidMove(move *Move) bool
	GetAllMoves() []Move
	Evaluate(prevPos *Position, move *Move, positionHashes map[uint64]bool) float32
}

func isWhitePiece(piece uint8) bool {
	return piece&isWhiteBit != 0
}

func isVacantSquare(board [8][8]uint8, row uint8, col uint8) bool {
	return board[row][col] == 0
}

func filterLongMoves(position *Position, moves []Move) []Move {
	var res []Move

	for _, move := range moves {
		piece, w := getPiece(move.toRow, move.toCol, position)
		if piece != 0 {
			if w != move.isWhite {
				res = append(res, move)
			}
			break
		}
		res = append(res, move)
	}

	return res
}

func getPiece(row uint8, col uint8, p *Position) (uint8, bool) {
	board := p.board
	piece := board[row][col]
	pieceBit := piece & ^isWhiteBit
	return pieceBit, piece&isWhiteBit != 0
}

func createPiece(piece uint8, isWhite bool) uint8 {
	if isWhite {
		return piece + 64
	} else {
		return piece
	}
}

func (p *Position) applyMoves(moves []Move) []PositionMove {
	res := make([]PositionMove, len(moves))
	for i, move := range moves {
		newPos := ApplyMove(*p, &move)
		res[i] = PositionMove{
			pos:  newPos,
			move: &move,
		}
	}
	return res
}

func absDiff(a, b uint8) uint8 {
	if a > b {
		return a - b
	} else {
		return b - a
	}
}

func (p *Position) IsValidMove(move *Move) bool {
	if move.isWhite != p.whiteTurn {
		return false
	}

	if moveOutsideBoard(move) {
		return false
	}

	fromPieceWithColor := p.board[move.fromRow][move.fromCol]
	fromPiece, _ := getPiece(move.fromRow, move.fromCol, p)
	if fromPieceWithColor == 0 || (move.isWhite && !isWhitePiece(fromPieceWithColor)) ||
		(!move.isWhite && isWhitePiece(fromPieceWithColor)) {
		return false
	}

	toPieceWithColor := p.board[move.toRow][move.toCol]
	if toPieceWithColor != 0 && (move.isWhite == isWhitePiece(toPieceWithColor)) {
		return false
	}

	if move.isCapture && isVacantSquare(p.board, move.toRow, move.toCol) {
		return false
	}

	if !move.isCapture && !isVacantSquare(p.board, move.toRow, move.toCol) {
		return false
	}

	//if the move is a castle and a king is attacked, the move is invalid
	if fromPiece == KingBit && absDiff(move.fromCol, move.toCol) > 1 {
		if isKingAttacked(p, p.whiteTurn) {
			return false
		}
	}

	// if the pawn jumps by 2 rows, check that the square it jumps over is not occupied
	if fromPiece == PawnBit && absDiff(move.fromRow, move.toRow) > 1 {
		var somePiece uint8
		if move.isWhite {
			somePiece, _ = getPiece(move.fromRow+1, move.fromCol, p)
		} else {
			somePiece, _ = getPiece(move.fromRow-1, move.fromCol, p)
		}
		if somePiece != 0 {
			return false
		}

	}

	//todo this copies the entire board - should be done with pointers
	newPos := ApplyMove(*p, move)
	if isKingAttacked(newPos, move.isWhite) {
		return false
	}

	return true
}

func getFirstPiece(p *Position, row int8, col int8, dirX int8, dirY int8) [2]int8 {
	row = row + dirX
	col = col + dirY
	if !isInsideBoard(row, col) {
		return [2]int8{9, 9}
	}
	if p.board[row][col] != 0 {
		return [2]int8{row, col}
	}
	return getFirstPiece(p, row, col, dirX, dirY)
}

func isInsideBoard(row int8, col int8) bool {
	return row <= 7 && row >= 0 && col <= 7 && col >= 0
}

func isKingAttacked(p *Position, isWhite bool) bool {

	isCorrectPiece := func(p *Position, locations [][2]int8, pieces []uint8, isWhite bool) bool {
		for _, location := range locations {
			if isInsideBoard(location[0], location[1]) {
				actualP, w := getPiece(uint8(location[0]), uint8(location[1]), p)
				if slices.Contains(pieces, actualP) && w == isWhite {
					return true
				}
			}
		}
		return false
	}

	i := int8(p.whiteKingPosRow)
	j := int8(p.whiteKingPosCol)
	if !isWhite {
		i = int8(p.blackKingPosRow)
		j = int8(p.blackKingPosCol)
	}

	piece, white := getPiece(uint8(i), uint8(j), p)
	if white == isWhite && piece == KingBit {
		locs := [][2]int8{{i + 1, j + 1}, {i + 1, j - 1}, {i + 1, j},
			{i - 1, j + 1}, {i - 1, j}, {i - 1, j - 1},
			{i, j + 1}, {i, j - 1}}
		if isCorrectPiece(p, locs, []uint8{KingBit}, !isWhite) {
			return true
		}

		if isCorrectPiece(p, [][2]int8{{i + ColorFactorInt(isWhite), j + 1}, {i + ColorFactorInt(isWhite), j - 1}}, []uint8{PawnBit}, !isWhite) {
			return true
		}

		locs = [][2]int8{{i + 2, j + 1}, {i + 2, j - 1}, {i - 2, j + 1}, {i - 2, j - 1},
			{i + 1, j + 2}, {i + 1, j - 2}, {i - 1, j + 2}, {i - 1, j - 2}}
		if isCorrectPiece(p, locs, []uint8{KnightBit}, !isWhite) {
			return true
		}

		if isCorrectPiece(p, [][2]int8{getFirstPiece(p, i, j, 1, 1)}, []uint8{BishopBit, QueenBit}, !isWhite) ||
			isCorrectPiece(p, [][2]int8{getFirstPiece(p, i, j, 1, -1)}, []uint8{BishopBit, QueenBit}, !isWhite) ||
			isCorrectPiece(p, [][2]int8{getFirstPiece(p, i, j, -1, 1)}, []uint8{BishopBit, QueenBit}, !isWhite) ||
			isCorrectPiece(p, [][2]int8{getFirstPiece(p, i, j, -1, -1)}, []uint8{BishopBit, QueenBit}, !isWhite) {
			return true
		}

		if isCorrectPiece(p, [][2]int8{getFirstPiece(p, i, j, 1, 0)}, []uint8{RookBit, QueenBit}, !isWhite) ||
			isCorrectPiece(p, [][2]int8{getFirstPiece(p, i, j, 0, 1)}, []uint8{RookBit, QueenBit}, !isWhite) ||
			isCorrectPiece(p, [][2]int8{getFirstPiece(p, i, j, -1, 0)}, []uint8{RookBit, QueenBit}, !isWhite) ||
			isCorrectPiece(p, [][2]int8{getFirstPiece(p, i, j, 0, -1)}, []uint8{RookBit, QueenBit}, !isWhite) {
			return true
		}

		return false
	}
	return false
}

func moveOutsideBoard(move *Move) bool {
	return move.fromCol < 0 || move.fromRow < 0 || move.toCol < 0 || move.toRow < 0 || move.fromRow > 7 || move.fromCol > 7 || move.toRow > 7 || move.toCol > 7
}

func (p *Position) GetAllMoves() []Move {
	var validMoves []Move

	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			moves := p.GetPossibleMoves(uint8(row), uint8(col), p.whiteTurn)
			for _, move := range moves {
				if p.IsValidMove(&move) {
					validMoves = append(validMoves, move)
				}
			}
		}
	}
	validMoves = p.addEnPassantMoves(validMoves)

	//printMoves(validMoves, "ValidMoves")
	return validMoves
}

func (p *Position) addEnPassantMoves(validMoves []Move) []Move {
	if p.whiteTurn {
		validMoves = addElPassantMoveIfPossible(validMoves, p, p.blackPawnDoubleStepCol, p.whiteTurn)
	} else {
		validMoves = addElPassantMoveIfPossible(validMoves, p, p.whitePawnDoubleStepCol, p.whiteTurn)
	}
	return validMoves
}

func (p *Position) PrintPosition() {
	strPos := "\n"
	for i := 7; i >= 0; i-- {
		strPos += fmt.Sprintf("%d ", i+1)

		for j := 0; j < 8; j++ {
			s := p.board[i][j] & ^(uint8(1) << 6)
			str := PieceToString(s)
			if s != p.board[i][j] {
				str = strings.ToUpper(str)
			}
			strPos += str
			strPos += " "
		}

		strPos += "\n"
	}
	strPos += "  a b c d e f g h\n"

	suffix := ""
	if p.isCheckmate {
		suffix = "Checkmate! "
	}

	strPos += fmt.Sprintf("Move: %d, turn white: %t, eval: %f, %s\n", p.moveNum, p.whiteTurn, p.evaluation, suffix)
	strPos += "***************"
	log.Println(strPos)
}

func PieceToString(pieceBit uint8) string {
	str := ""
	switch pieceBit {
	case 0:
		str = "-"
	case PawnBit:
		str = "p"
	case KnightBit:
		str = "n"
	case BishopBit:
		str = "b"
	case RookBit:
		str = "r"
	case QueenBit:
		str = "q"
	case KingBit:
		str = "k"
	}
	return str
}

func cloneBoard(original [8][8]uint8) [8][8]uint8 {
	var clone [8][8]uint8
	for i := range original {
		for j := range original[i] {
			clone[i][j] = original[i][j]
		}
	}
	return clone
}

func (p *Position) UpdateCastingAllowance(lastMove *Move) {
	piece, _ := getPiece(lastMove.toRow, lastMove.toCol, p)

	if piece == KingBit {
		if lastMove.isWhite {
			p.whiteShortCastleAllowed = false
			p.whiteLongCastleAllowed = false
		} else {
			p.blackShortCastleAllowed = false
			p.blackLongCastleAllowed = false
		}
	}

	if piece == RookBit {
		if lastMove.isWhite {
			if lastMove.fromRow == 0 && lastMove.fromCol == 0 {
				p.whiteLongCastleAllowed = false
			}

			if lastMove.fromRow == 0 && lastMove.fromCol == 7 {
				p.whiteShortCastleAllowed = false
			}

		} else {
			if lastMove.fromRow == 7 && lastMove.fromCol == 0 {
				p.blackLongCastleAllowed = false
			}
			if lastMove.fromRow == 7 && lastMove.fromCol == 7 {
				p.blackShortCastleAllowed = false
			}
		}

	}

}

func convertBoard(board *[8][8]string) [8][8]uint8 {
	res := [8][8]uint8{}
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			c := board[i][j]
			p := PieceStrToPieceBit(c)
			if len(c) > 0 && unicode.IsUpper([]rune(c)[0]) {
				p += 64
			}
			res[7-i][j] = p
		}
	}
	return res
}

func PieceStrToPieceBit(c string) uint8 {
	p := uint8(0)
	switch strings.ToLower(c) {
	case "p":
		p = PawnBit
	case "n":
		p = KnightBit
	case "b":
		p = BishopBit
	case "r":
		p = RookBit
	case "q":
		p = QueenBit
	case "k":
		p = KingBit
	}
	return p
}

func (p *Position) InitPosition(board *[8][8]string, moveNum int8, turnWhite bool) *Position {
	newBoard := convertBoard(board)

	pos := Position{
		board:                   newBoard,
		moveNum:                 moveNum,
		whiteTurn:               turnWhite,
		whiteLongCastleAllowed:  true,
		whiteShortCastleAllowed: true,
		blackLongCastleAllowed:  true,
		blackShortCastleAllowed: true,
	}

	for i := uint8(0); i < 8; i++ {
		for j := uint8(0); j < 8; j++ {
			piece, isWhite := getPiece(i, j, &pos)
			if piece == KingBit {
				if isWhite {
					pos.whiteKingPosRow = i
					pos.whiteKingPosCol = j
				} else {
					pos.blackKingPosRow = i
					pos.blackKingPosCol = j
				}
			}
		}
	}

	return &pos
}

func ApplyMove(p Position, move *Move) *Position {
	ApplyMovePointers(&p, move)
	return &p
}

func ClonePosition(p *Position) *Position {
	return &Position{
		board:                   cloneBoard(p.board),
		moveNum:                 p.moveNum,
		whiteTurn:               p.whiteTurn,
		evaluation:              p.evaluation,
		isCheckmate:             p.isCheckmate,
		availableMoves:          p.availableMoves,
		whiteKingPosRow:         p.whiteKingPosRow,
		blackKingPosRow:         p.blackKingPosRow,
		whiteKingPosCol:         p.whiteKingPosCol,
		blackKingPosCol:         p.blackKingPosCol,
		hash:                    p.hash,
		whiteShortCastleAllowed: p.whiteShortCastleAllowed,
		blackShortCastleAllowed: p.blackShortCastleAllowed,
		whiteLongCastleAllowed:  p.whiteLongCastleAllowed,
		blackLongCastleAllowed:  p.blackLongCastleAllowed,
		whitePawnDoubleStepCol:  p.whitePawnDoubleStepCol,
		blackPawnDoubleStepCol:  p.blackPawnDoubleStepCol,
	}
}

func ApplyMovePointers(p *Position, move *Move) {
	piece := p.board[move.fromRow][move.fromCol]
	origPiece, _ := getPiece(move.fromRow, move.fromCol, p)

	// regular move
	if move.pawnPromotePiece != 0 {
		piece = move.pawnPromotePiece
	}
	p.board[move.toRow][move.toCol] = piece
	p.board[move.fromRow][move.fromCol] = 0

	if move.isEnPassant {
		p.board[move.fromRow][move.toCol] = 0
	}

	if origPiece == KingBit {
		if move.isWhite {
			p.whiteKingPosRow = move.toRow
			p.whiteKingPosCol = move.toCol
		} else {
			p.blackKingPosRow = move.toRow
			p.blackKingPosCol = move.toCol
		}

		if move.fromCol > 1+move.toCol || move.toCol > 1+move.fromCol {
			if move.fromCol > move.toCol {
				p.board[move.fromRow][3] = p.board[move.fromRow][0]
				p.board[move.fromRow][0] = 0
			} else {
				p.board[move.fromRow][5] = p.board[move.fromRow][7]
				p.board[move.fromRow][7] = 0
			}
		}
	}

	if origPiece == PawnBit && absDiff(move.fromRow, move.toRow) == 2 {
		if move.isWhite {
			p.whitePawnDoubleStepCol = move.toCol
		} else {
			p.blackPawnDoubleStepCol = move.toCol
		}
	} else {
		if move.isWhite {
			p.whitePawnDoubleStepCol = 0
		} else {
			p.blackPawnDoubleStepCol = 0
		}
	}

	p.UpdateCastingAllowance(move)

	p.whiteTurn = !p.whiteTurn
	p.moveNum += 1
}

// positionToFEN converts a Position struct to a FEN string.
func (p *Position) positionToFEN() string {
	var sb strings.Builder

	// Board layout
	for row := 7; row >= 0; row-- {
		emptyCount := 0
		for col := 0; col < 8; col++ {
			piece, isWhite := getPiece(uint8(row), uint8(col), p)
			if piece == 0 {
				emptyCount++
			} else {
				if emptyCount > 0 {
					sb.WriteString(fmt.Sprintf("%d", emptyCount))
					emptyCount = 0
				}
				pieceStr := PieceToString(piece)
				if isWhite {
					pieceStr = strings.ToUpper(pieceStr)
				}
				sb.WriteString(pieceStr)
			}
		}
		if emptyCount > 0 {
			sb.WriteString(fmt.Sprintf("%d", emptyCount))
		}
		if row > 0 {
			sb.WriteString("/")
		}
	}

	// Active color
	if p.whiteTurn {
		sb.WriteString(" w ")
	} else {
		sb.WriteString(" b ")
	}

	// Castling availability
	castling := ""
	if p.whiteShortCastleAllowed {
		castling += "K"
	}
	if p.whiteLongCastleAllowed {
		castling += "Q"
	}
	if p.blackShortCastleAllowed {
		castling += "k"
	}
	if p.blackLongCastleAllowed {
		castling += "q"
	}
	if castling == "" {
		castling = "-"
	}
	sb.WriteString(castling + " ")

	// En passant target square (not implemented, so default to "-")
	sb.WriteString("- ")

	// Halfmove clock (not in struct, default to 0)
	sb.WriteString("0 ")

	// Fullmove number
	sb.WriteString(fmt.Sprintf("%d", p.moveNum))

	return sb.String()
}
