package chess

import (
	"testing"
)

func setup() {
	Debug = true
	TreeDepth = 2
}

// TestUCIEngine simulates creating a game and making several moves
func TestUCIEngine(t *testing.T) {
	setup()
	RandomSeed = 1722630608576018000

	var game *Game
	var finished bool
	game, finished = HandleUciCommand("ucinewgame", game)
	//look there is a bug
	//cmd := "position startpos moves d2d4 e7e6 c2c4 d8h4 g2g3 f8b4 b1c3 h4e4 g1f3 b8c6 f1g2 d7d5 e1g1 b4c3 b2c3 e4f5 d1d3 f5d3 e2d3 d5c4 d3c4 g8f6 f1e1 e8g8 c1a3 f8d8 a1d1 h7h6 f3e5 c6e5 e1e5 f6g4 e5e1 f7f5 d4d5 e6d5 c4d5 g4f6 a3e7 d8e8 e7f6 e8e1 d1e1 g7f6 e1e7 c7c5 d5d6 g8f8 g2d5 h6h5 h2h4 a7a6 e7h7 f8e8 d5f7 e8d7 f7h5 d7d6 h5f3 a8a7 c3c4 c8e6 h7b7 a7b7 f3b7 e6c4 a2a3 d6c7 b7f3 c4f7 h4h5 c5c4 g1f1 c4c3 f1e2 f7h5 f3h5 f5f4 g3f4 a6a5 e2d3 f6f5 d3c3 c7b7 c3c4 b7a8 c4b5 a5a4 b5a4 a8b8 a4b5 b8a7 b5c5 a7a8 c5d5 a8b8 d5e5 b8b7 e5f5 b7c7 f5e6 c7c8 f4f5 c8b8 f5f6 b8c7 f6f7 c7c8 f7f8q\n"

	cmd := "position startpos moves e2e4 e7e5 b1c3 b8c6 f1c4 d7d5 e4d5 c6d4 c3e4 f8b4 c2c3 b4c3 d2c3 d4f5 c4b5 e8f8 c3c4 f5h4 g2g3 h4g2 e1f1 g2e3 c1e3 a7a5 e3c5 g8e7 c5e7 f8e7 f1e1 c8f5 d5d6 c7d6 e4c3 h8f8 c3d5 e7e6 g1f3 e5e4 f3d4 e6e5 f2f4\n"
	game, finished = HandleUciCommand(cmd, game)
	game, finished = HandleUciCommand("go infinite", game)

	lastCommand := commandsSentToUCI[len(commandsSentToUCI)-1]
	println(lastCommand)

	if finished {
		t.Errorf("The game shouldn't be finished")
	}

}

func TestValidKingMovesUnderCheck(t *testing.T) {
	setup()
	var game *Game
	game, _ = HandleUciCommand("ucinewgame", game)
	game, _ = HandleUciCommand("position startpos moves e2e4 e7e5 f1b5 c7c6 b5c4 d7d5 e4d5 c6d5 c4b5 e8e7 d1f3 f7f5 f3a3 e7e6 a3e3 e6f7 e3e5 c8d7 e5d5 f7f6 d5d4 f6f7 b5c4 d7e6 c4e6", game)

	moves := game.position.GetAllMoves()

	if len(moves) != 4 {
		t.Errorf("Expected 4 moves in this position, got %d", len(moves))
	}

}

func TestValidMovesForStartingPosition(t *testing.T) {
	setup()
	var game *Game
	game, _ = HandleUciCommand("ucinewgame", game)
	game, _ = HandleUciCommand("position startpos moves e2e4 e7e5", game)

	moves := game.position.GetAllMoves()

	if len(moves) != 29 {
		t.Errorf("Expected 29 moves in this position, got %d", len(moves))
	}
}

func TestMatInOne(t *testing.T) {
	setup()
	var game *Game
	game, _ = HandleUciCommand("ucinewgame", game)
	game, _ = HandleUciCommand("position startpos moves e2e4 e7e5 f1c4 f8c5 d1h5 g8f6", game)
	game, _ = HandleUciCommand("go infinite", game)

	lastCommand := commandsSentToUCI[len(commandsSentToUCI)-1]
	if lastCommand != "bestmove h5f7\n" {
		t.Errorf("The engine didn't find a mate in 1")
	}

}

func TestCheckEscape(t *testing.T) {
	setup()
	var game *Game
	game, _ = HandleUciCommand("ucinewgame", game)
	game, _ = HandleUciCommand("position startpos moves f2f4 e7e6 c2c4 d8h4", game)
	game, _ = HandleUciCommand("go infinite", game)

	lastCommand := commandsSentToUCI[len(commandsSentToUCI)-1]
	if lastCommand != "bestmove g2g3\n" {
		t.Errorf("The engine didn't escape a check")
	}
}

func TestFreePieceCapture(t *testing.T) {
	setup()
	var game *Game
	game, _ = HandleUciCommand("ucinewgame", game)
	game, _ = HandleUciCommand("position startpos moves f2f4 e7e6 e2e4 d8g5", game)
	game, _ = HandleUciCommand("go infinite", game)

	lastCommand := commandsSentToUCI[len(commandsSentToUCI)-1]
	if lastCommand != "bestmove f4g5\n" {
		t.Errorf("The engine didn't capture a free piece")
	}
}

func TestZobristHashing(t *testing.T) {
	setup()
	var game *Game
	game, _ = HandleUciCommand("ucinewgame", game)
	game, _ = HandleUciCommand("position startpos", game)
	positionHash := ComputeZobristHash(game.position)
	moves := []string{"e2e4", "e7e5", "g1f3", "b8c6", "f3e5", "c6e5"}
	p := &game.position
	for _, moveStr := range moves {
		move := parseMove(moveStr, *p)
		newHash := UpdateZobristHash(positionHash, &move, *p)
		if newHash == positionHash {
			t.Errorf("Zobrist hash must have changed")
		}
		positionHash = newHash
		ApplyMovePointers(*p, &move)
	}

	finalHash := ComputeZobristHash(*p)
	if finalHash != positionHash {
		t.Errorf("Zobrist hash was not computed correctly")
	}

}

func TestThreeFoldRepetition(t *testing.T) {
	setup()
	var game *Game
	game, _ = HandleUciCommand("ucinewgame", game)
	game, _ = HandleUciCommand("position startpos moves g1f3 b8c6 f3g1 c6b8 g1f3 b8c6 f3g1 c6b8", game)
	prevPosition := game.position

	game, _ = HandleUciCommand("position startpos moves g1f3 b8c6 f3g1 c6b8 g1f3 b8c6 f3g1 c6b8 g1f3", game)

	evaluation := game.position.Evaluate(prevPosition, game.GetLastMove(), game.positionHashes)

	if evaluation != ColorFactor(game.GetLastMove().isWhite)*ThreeFoldRepetitionEvalution {
		t.Errorf("The evaluation should be a three fold repetition evaluation, %f", evaluation)
	}
}

func TestEnPassant(t *testing.T) {
	setup()
	var game *Game
	game, _ = HandleUciCommand("ucinewgame", game)
	game, _ = HandleUciCommand("position startpos moves f2f4 h7h6 f4f5 f7f6 g2g4 g7g5", game)
	game, _ = HandleUciCommand("go infinite", game)

	lastCommand := commandsSentToUCI[len(commandsSentToUCI)-1]
	if lastCommand != "bestmove f5g6\n" {
		t.Errorf("The engine didn't capture a free piece")
	}
}

func TestEnPassantBlack(t *testing.T) {
	setup()
	var game *Game
	game, _ = HandleUciCommand("ucinewgame", game)
	game, _ = HandleUciCommand("position startpos moves h2h4 f7f5 f2f3 f5f4 g2g4", game)
	game, _ = HandleUciCommand("go infinite", game)

	lastCommand := commandsSentToUCI[len(commandsSentToUCI)-1]
	if lastCommand != "bestmove f4g3\n" {
		t.Errorf("The engine didn't capture a free piece")
	}
}
