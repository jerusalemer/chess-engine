package chess

import (
	"testing"
)

// TestUCIEngine simulates creating a game and making several moves
func TestUCIEngine(t *testing.T) {
	var game *Game
	var finished bool
	game, finished = HandleUciCommand("ucinewgame", game)
	cmd := "position startpos moves d2d4 d7d5 b1c3 g8f6 e2e3 c8f5 f1b5 c7c6 b5a4 b8d7 c1d2 e7e5 d4e5 d7e5 g2g3 d8d7 a1c1 e8c8 f2f4 e5c4 a4c6 d7c6 c3d5 f6d5 e3e4 f5e4 d1g4 c8b8 g1f3 c4d2 e1d2 d5f4 d2e1 f8b4 c2c3 f4d3 e1f1 d3c1 g4f4 b8c8 f4g4 c8b8 g4f4 b8c8 f4g4 f7f5 g4g7 d8d1 f1g2 e4f3 g2h3 f3g2\n"
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
