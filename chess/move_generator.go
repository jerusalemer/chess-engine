package chess

func (p *Position) appendMoveIfValid(fromRow, fromCol, toRow, toCol, pawnPromotePiece uint8, isWhite, isCapture, checkVacancy bool, moves []Move) []Move {
	if isSquareOutsideTheBoard(toRow, toCol) {
		return moves
	}

	piece, color := getPiece(toRow, toCol, p)
	if !checkVacancy {
		if piece != 0 {
			isCapture = true
		}
		return append(moves, Move{fromRow, fromCol, toRow, toCol, isWhite, isCapture, pawnPromotePiece, false})
	}

	if piece != 0 && color == isWhite {
		return moves
	}

	if piece != 0 && color != isWhite {
		fromPiece, _ := getPiece(fromRow, fromCol, p)
		if !isCapture && fromPiece == PawnBit {
			return moves
		}

		isCapture = true
	}

	return append(moves, Move{fromRow, fromCol, toRow, toCol, isWhite, isCapture, pawnPromotePiece, false})
}

func isSquareOutsideTheBoard(toRow uint8, toCol uint8) bool {
	return toRow < 0 || toRow > 7 || toCol < 0 || toCol > 7
}

func getKingInitSquare(isWhite bool) (uint8, uint8) {
	if isWhite {
		return 0, 4
	} else {
		return 7, 4
	}
}

func (p *Position) addCastlingMoves(row, col uint8, isWhite bool, moves []Move) []Move {

	kRow, kCol := getKingInitSquare(isWhite)
	if row == kRow && kCol == col {
		if isWhite {
			piece, w := getPiece(0, 0, p)
			if piece == RookBit && w == isWhite && p.board[0][1] == 0 && p.board[0][2] == 0 && p.board[0][3] == 0 && p.whiteLongCastleAllowed {
				moves = p.appendMoveIfValid(row, col, row, 2, 0, isWhite, false, true, moves)
			}
			piece, w = getPiece(0, 7, p)
			if piece == RookBit && w == isWhite && p.board[0][6] == 0 && p.board[0][5] == 0 && p.whiteShortCastleAllowed {
				moves = p.appendMoveIfValid(row, col, row, 6, 0, isWhite, false, true, moves)
			}

		} else {
			piece, w := getPiece(7, 0, p)
			if piece == RookBit && w == isWhite && p.board[7][1] == 0 && p.board[7][2] == 0 && p.board[7][3] == 0 && p.blackLongCastleAllowed {
				moves = p.appendMoveIfValid(row, col, row, 2, 0, isWhite, false, true, moves)
			}
			piece, w = getPiece(7, 7, p)
			if piece == RookBit && w == isWhite && p.board[7][6] == 0 && p.board[7][5] == 0 && p.blackLongCastleAllowed {
				moves = p.appendMoveIfValid(row, col, row, 6, 0, isWhite, false, true, moves)
			}

		}

	}
	return moves
}

func (p *Position) GetPossibleMoves(row uint8, col uint8, isWhite bool) []Move {
	var moves []Move

	pieceBit, isWhitePiece := getPiece(row, col, p)
	if pieceBit == 0 || isWhitePiece != isWhite {
		return nil
	}

	if pieceBit == PawnBit {
		if isWhite {
			if row < 6 {
				moves = p.appendMoveIfValid(row, col, row+1, col, 0, isWhite, false, true, moves)
				moves = p.appendMoveIfValid(row, col, row+1, col+1, 0, isWhite, true, true, moves)
				moves = p.appendMoveIfValid(row, col, row+1, col-1, 0, isWhite, true, true, moves)
			} else {
				// Handle promotion
				for _, pieceToPromote := range piecesToPromote {
					pieceWithWhite := createPiece(pieceToPromote, isWhite)
					moves = p.appendMoveIfValid(row, col, row+1, col, pieceWithWhite, isWhite, false, true, moves)
					moves = p.appendMoveIfValid(row, col, row+1, col+1, pieceWithWhite, isWhite, true, true, moves)
					moves = p.appendMoveIfValid(row, col, row+1, col-1, pieceWithWhite, isWhite, true, true, moves)
				}
			}
			if row == 1 {
				moves = p.appendMoveIfValid(row, col, row+2, col, 0, isWhite, false, true, moves)
			}
		} else {
			if row > 1 {
				moves = p.appendMoveIfValid(row, col, row-1, col, 0, isWhite, false, true, moves)
				moves = p.appendMoveIfValid(row, col, row-1, col-1, 0, isWhite, true, true, moves)
				moves = p.appendMoveIfValid(row, col, row-1, col+1, 0, isWhite, true, true, moves)
			} else {
				for _, pieceToPromote := range piecesToPromote {
					pieceWithWhite := createPiece(pieceToPromote, isWhite)
					moves = p.appendMoveIfValid(row, col, row-1, col, pieceWithWhite, isWhite, false, true, moves)
					moves = p.appendMoveIfValid(row, col, row-1, col-1, pieceWithWhite, isWhite, true, true, moves)
					moves = p.appendMoveIfValid(row, col, row-1, col+1, pieceWithWhite, isWhite, true, true, moves)
				}
			}
			if row == 6 {
				moves = p.appendMoveIfValid(row, col, row-2, col, 0, isWhite, false, true, moves)
			}
		}

	}

	if pieceBit == KnightBit {
		moves = p.appendMoveIfValid(row, col, row+1, col+2, 0, isWhite, false, true, moves)
		moves = p.appendMoveIfValid(row, col, row-1, col+2, 0, isWhite, false, true, moves)
		moves = p.appendMoveIfValid(row, col, row+1, col-2, 0, isWhite, false, true, moves)
		moves = p.appendMoveIfValid(row, col, row-1, col-2, 0, isWhite, false, true, moves)
		moves = p.appendMoveIfValid(row, col, row+2, col+1, 0, isWhite, false, true, moves)
		moves = p.appendMoveIfValid(row, col, row+2, col-1, 0, isWhite, false, true, moves)
		moves = p.appendMoveIfValid(row, col, row-2, col+1, 0, isWhite, false, true, moves)
		moves = p.appendMoveIfValid(row, col, row-2, col-1, 0, isWhite, false, true, moves)

	}

	if pieceBit == BishopBit || pieceBit == QueenBit {
		var candidates []Move
		var candidates2 []Move
		var candidates3 []Move
		var candidates4 []Move
		for i := uint8(1); i < 8; i++ {
			candidates = p.appendMoveIfValid(row, col, row+i, col+i, 0, isWhite, false, false, candidates)
			candidates2 = p.appendMoveIfValid(row, col, row+i, col-i, 0, isWhite, false, false, candidates2)
			candidates3 = p.appendMoveIfValid(row, col, row-i, col+i, 0, isWhite, false, false, candidates3)
			candidates4 = p.appendMoveIfValid(row, col, row-i, col-i, 0, isWhite, false, false, candidates4)
		}
		moves = append(moves, filterLongMoves(p, candidates)...)
		moves = append(moves, filterLongMoves(p, candidates2)...)
		moves = append(moves, filterLongMoves(p, candidates3)...)
		moves = append(moves, filterLongMoves(p, candidates4)...)
	}

	if pieceBit == RookBit || pieceBit == QueenBit {
		var candidates []Move
		var candidates2 []Move
		var candidates3 []Move
		var candidates4 []Move
		for i := uint8(1); i < 8; i++ {
			candidates = p.appendMoveIfValid(row, col, row, col+i, 0, isWhite, false, false, candidates)
			candidates2 = p.appendMoveIfValid(row, col, row, col-i, 0, isWhite, false, false, candidates2)
			candidates3 = p.appendMoveIfValid(row, col, row-i, col, 0, isWhite, false, false, candidates3)
			candidates4 = p.appendMoveIfValid(row, col, row+i, col, 0, isWhite, false, false, candidates4)
		}
		moves = append(moves, filterLongMoves(p, candidates)...)
		moves = append(moves, filterLongMoves(p, candidates2)...)
		moves = append(moves, filterLongMoves(p, candidates3)...)
		moves = append(moves, filterLongMoves(p, candidates4)...)
	}

	if pieceBit == KingBit {
		moves = p.appendMoveIfValid(row, col, row+1, col, 0, isWhite, false, true, moves)
		moves = p.appendMoveIfValid(row, col, row+1, col+1, 0, isWhite, false, true, moves)
		moves = p.appendMoveIfValid(row, col, row+1, col-1, 0, isWhite, false, true, moves)
		moves = p.appendMoveIfValid(row, col, row-1, col, 0, isWhite, false, true, moves)
		moves = p.appendMoveIfValid(row, col, row-1, col+1, 0, isWhite, false, true, moves)
		moves = p.appendMoveIfValid(row, col, row-1, col-1, 0, isWhite, false, true, moves)
		moves = p.appendMoveIfValid(row, col, row, col+1, 0, isWhite, false, true, moves)
		moves = p.appendMoveIfValid(row, col, row, col-1, 0, isWhite, false, true, moves)

		moves = p.addCastlingMoves(row, col, isWhite, moves)
	}

	//printMoves(moves, "GetPossibleMoves")

	return moves
}

func createEnPassantMove(fromRow, fromCol, toRow, toCol uint8, isWhite bool, p *Position) (*Move, bool) {
	if isSquareOutsideTheBoard(fromRow, fromCol) || isSquareOutsideTheBoard(toRow, toCol) {
		return nil, false
	}
	piece, w := getPiece(fromRow, fromCol, p)
	if piece != PawnBit || isWhite != w {
		return nil, false
	}

	if toPiece, _ := getPiece(toRow, toCol, p); toPiece != 0 {
		return nil, false
	}
	return &Move{
		fromRow:     fromRow,
		fromCol:     fromCol,
		toRow:       toRow,
		toCol:       toCol,
		isWhite:     isWhite,
		isCapture:   true,
		isEnPassant: true,
	}, true
}

func addElPassantMoveIfPossible(moves []Move, p *Position, jumpingPawnCol uint8, white bool) []Move {
	if jumpingPawnCol == 0 {
		return moves
	}
	if white {
		if m, valid := createEnPassantMove(4, jumpingPawnCol+1, 5, jumpingPawnCol, white, p); valid {
			moves = append(moves, *m)
		}
		if m, valid := createEnPassantMove(4, jumpingPawnCol-1, 5, jumpingPawnCol, white, p); valid {
			moves = append(moves, *m)
		}
	} else {
		if m, valid := createEnPassantMove(3, jumpingPawnCol+1, 2, jumpingPawnCol, white, p); valid {
			moves = append(moves, *m)
		}
		if m, valid := createEnPassantMove(3, jumpingPawnCol-1, 2, jumpingPawnCol, white, p); valid {
			moves = append(moves, *m)
		}
	}
	return moves
}
