package chess

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

var commandsSentToUCI []string

func StartUCI() {

	logFile, _ := os.OpenFile("/Users/artg/dev/chess-engine/build/uci_engine.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	log.SetOutput(logFile)
	defer func() {
		if r := recover(); r != nil {
			log.Printf("UNHANDLED PANIC: %v", r)
		}
	}()
	defer logFile.Close()

	reader := bufio.NewReader(os.Stdin)
	var game *Game

	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		log.Printf("Received: %s\n", text)

		var isFinished bool
		game, isFinished = HandleUciCommand(text, game)
		if isFinished {
			return
		}
	}
}

func HandleUciCommand(commandText string, game *Game) (*Game, bool) {
	switch {
	case commandText == "uci":
		handleUCI()
	case commandText == "isready":
		handleIsReady()
	case commandText == "ucinewgame":
		game = handleUCINewGame()
	case strings.HasPrefix(commandText, "position"):
		game = handlePosition(commandText)
	case strings.HasPrefix(commandText, "go"):
		handleGo(game)
	case commandText == "stop":
		handleStop(game)
		return game, true
	case commandText == "quit":
		handleQuit(game)
		return game, true
	default:
		// Handle other commands
	}
	return game, false
}

func handleUCI() {
	sendToUCI("id name SimpleButCuteChessEngine")
	sendToUCI("id author Art")
	sendToUCI("uciok")
}

func sendToUCI(cmd string) {
	log.Println("Output: ", cmd)
	fmt.Println(cmd)
	commandsSentToUCI = append(commandsSentToUCI, cmd)
}

func handleIsReady() {
	sendToUCI("readyok")
}

func handleUCINewGame() *Game {
	return NewGame()
}

func handlePosition(command string) *Game {
	parts := strings.Split(command, " ")
	if len(parts) < 2 {
		println("Invalid command: ", command)
		os.Exit(1)
	}

	var game = NewGame()

	if parts[1] == "startpos" {
	}

	if parts[1] == "fen" && len(parts) > 2 {
		fenString := strings.Join(parts[2:], " ")
		fenInfo := parseFEN(fenString)
		game.InitGame(&fenInfo.Board, fenInfo.WhiteTurn, TreeDepth)
	}

	// Apply moves if any
	if len(parts) > 2 {
		moveIndex := 2
		if parts[1] == "fen" {
			moveIndex = 7 // Skips the FEN parts
		}
		if parts[moveIndex] == "moves" {
			for i := moveIndex + 1; i < len(parts); i++ {
				move := parseMove(parts[i], &game.position)
				ApplyMovePointers(&game.position, &move)
			}
		}
	}
	game.position.PrintPosition()
	return game
}

type FENInfo struct {
	Board           [8][8]string
	WhiteTurn       bool
	CastlingRights  string
	EnPassantSquare string
	HalfmoveClock   int
	FullmoveNumber  int
}

func parseFEN(fen string) FENInfo {
	parts := strings.Split(fen, " ")
	board := fenToBoard(parts[0])
	activeColor := parts[1]
	whiteTurn := false
	if activeColor == "w" {
		whiteTurn = true
	}
	castlingRights := parts[2]
	enPassantSquare := parts[3]
	halfmoveClock := atoi(parts[4])
	fullmoveNumber := atoi(parts[5])

	return FENInfo{
		Board:           board,
		WhiteTurn:       whiteTurn,
		CastlingRights:  castlingRights,
		EnPassantSquare: enPassantSquare,
		HalfmoveClock:   halfmoveClock,
		FullmoveNumber:  fullmoveNumber,
	}
}

func fenToBoard(fen string) [8][8]string {
	var board [8][8]string
	ranks := strings.Split(fen, "/")
	for i, rank := range ranks {
		file := 0
		for _, char := range rank {
			if char >= '1' && char <= '8' {
				for k := 0; k < int(char-'0'); k++ {
					board[i][file] = ""
					file++
				}
			} else {
				board[i][file] = string(char)
				file++
			}
		}
	}
	return board
}

func atoi(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

func handleGo(game *Game) {
	game.MakeMove()
	move := *game.GetLastMove()
	uciMove := moveToUCI(move)

	str := "Best Sequence: "
	for _, bestMove := range game.bestMoveSequence {
		str += fmt.Sprintf("%s, ", bestMove)
	}
	log.Println(str)
	sendToUCI("bestmove " + uciMove + "\n")
}

func moveToUCI(m Move) string {
	uciMove := fmt.Sprintf("%s%d%s%d", string(rune('a'+m.fromCol)), m.fromRow+1, string(rune('a'+m.toCol)), m.toRow+1)
	if m.pawnPromotePiece != 0 {
		pieceToPromote := m.pawnPromotePiece & ^isWhiteBit
		uciMove += PieceToString(pieceToPromote)
	}
	return uciMove
}

func handleStop(game *Game) {
	game.isFinished = true
}

func handleQuit(game *Game) {
	game.isFinished = true
}

func parseMove(moveStr string, p *Position) Move {
	// Convert UCI move string to Move struct
	m := Move{
		fromRow: moveStr[1] - '1',
		fromCol: moveStr[0] - 'a',
		toRow:   moveStr[3] - '1',
		toCol:   moveStr[2] - 'a',
	}
	_, isWhite := getPiece(m.fromRow, m.fromCol, p)
	m.isWhite = isWhite
	piece, _ := getPiece(m.toRow, m.toCol, p)
	if piece != 0 {
		m.isCapture = true
	}
	if len(moveStr) == 5 {
		pawnPromotePiece := moveStr[4:5]
		m.pawnPromotePiece = PieceStrToPieceBit(pawnPromotePiece)
	}
	return m
}
