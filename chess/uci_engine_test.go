package chess

import (
	"testing"
)

// TestUCIEngine simulates creating a game and making several moves
func TestUCIEngine(t *testing.T) {
	RandomSeed = 1722317207137502000

	var game *Game
	var finished bool
	game, finished = HandleUciCommand("ucinewgame", game)
	cmd := "position startpos moves e2e4 e7e5 f1b5 c7c6 b5c4 d7d5 e4d5 c6d5 c4b5 e8e7 d1h5 f7f5 h5h4 g7g5 h4g5 g8f6 g1f3 c8e6 f3e5 e7d6 g5e3 a7a5 e1g1 d6e7 e3a3 d8d6 a3d6 e7d6 d2d4 f6e4 f2f4 h8g8 c1e3 b7b6 f1e1 d6c7 e1c1 c7c8 c2c3 f8h6 c1c2 g8f8 b2b3 f8h8 c2c1 h8f8 c1f1 f8g8 g2g3 h6f8 f1f3 e4f6 e3f2 f8e7 b1d2 e7d6 e5d3 g8g7 f2e3 g7b7 a1c1 h7h6 c1a1 d6a3 f3f1 a8a7 d2f3 b7c7 d3e5 c7c3 b5d7 b8d7 e5d7 a7d7 f3e5 d7g7 e3d2 c3c7 f1e1 c7a7 e1b1 f6e4 d2e1 a5a4 b3a4 a7a4 b1b6 e6g8 b6c6 c8b8 a1b1 b8a8 c6c8 a8a7 e5c6 a7a6 c6b8 a6a7 b8c6 a7a6 c6b8 a6a7\n"
	game, finished = HandleUciCommand(cmd, game)
	game, finished = HandleUciCommand("go infinite", game)

	lastCommand := commandsSentToUCI[len(commandsSentToUCI)-1]
	println(lastCommand)

	if finished {
		t.Errorf("The game shouldn't be finished")
	}

}

func TestValidKingMovesUnderCheck(t *testing.T) {
	var game *Game
	game, _ = HandleUciCommand("ucinewgame", game)
	game, _ = HandleUciCommand("position startpos moves e2e4 e7e5 f1b5 c7c6 b5c4 d7d5 e4d5 c6d5 c4b5 e8e7 d1f3 f7f5 f3a3 e7e6 a3e3 e6f7 e3e5 c8d7 e5d5 f7f6 d5d4 f6f7 b5c4 d7e6 c4e6", game)

	moves := game.position.GetAllMoves()

	if len(moves) != 4 {
		t.Errorf("Expected 4 moves in this position, got %d", len(moves))
	}

}

func TestValidMovesForStartingPosition(t *testing.T) {
	var game *Game

	game, _ = HandleUciCommand("ucinewgame", game)
	game, _ = HandleUciCommand("position startpos moves e2e4 e7e5", game)

	moves := game.position.GetAllMoves()

	if len(moves) != 29 {
		t.Errorf("Expected 29 moves in this position, got %d", len(moves))
	}
}

func TestMatInOne(t *testing.T) {
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
	var game *Game
	game, _ = HandleUciCommand("ucinewgame", game)
	game, _ = HandleUciCommand("position startpos moves f2f4 e7e6 e2e4 d8h4", game)
	game, _ = HandleUciCommand("go infinite", game)

	lastCommand := commandsSentToUCI[len(commandsSentToUCI)-1]
	if lastCommand != "bestmove g2g3\n" {
		t.Errorf("The engine didn't escape a check")
	}
}

func TestFreePieceCapture(t *testing.T) {
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
	var game *Game
	game, _ = HandleUciCommand("ucinewgame", game)
	game, _ = HandleUciCommand("position startpos", game)
	positionHash := ComputeZobristHash(&game.position)
	moves := []string{"e2e4", "e7e5", "g1f3", "b8c6", "f3e5", "c6e5"}
	p := &game.position
	for _, moveStr := range moves {
		move := parseMove(moveStr, p)
		newHash := UpdateZobristHash(positionHash, &move, p)
		if newHash == positionHash {
			t.Errorf("Zobrist hash must have changed")
		}
		positionHash = newHash
		ApplyMovePointers(p, &move)
	}

	finalHash := ComputeZobristHash(p)
	if finalHash != positionHash {
		t.Errorf("Zobrist hash was not computed correctly")
	}

}

func TestThreeFoldRepetition(t *testing.T) {
	var game *Game
	game, _ = HandleUciCommand("ucinewgame", game)
	game, _ = HandleUciCommand("position startpos moves g1f3 b8c6 f3g1 c6b8 g1f3 b8c6 f3g1 c6b8", game)
	prevPosition := game.position

	game, _ = HandleUciCommand("position startpos moves g1f3 b8c6 f3g1 c6b8 g1f3 b8c6 f3g1 c6b8 g1f3", game)

	evaluation := game.position.Evaluate(&prevPosition, game.GetLastMove(), game.positionHashes)

	if evaluation != ColorFactor(game.GetLastMove().isWhite)*ThreeFoldRepetitionEvalution {
		t.Errorf("The evaluation should be a three fold repetition evaluation, %f", evaluation)
	}
}
