package othellogame

import (
	"fmt"
	"log"

	"github.com/ArminGh02/othello-bot/pkg/consts"
	"github.com/ArminGh02/othello-bot/pkg/othellogame/cell"
	"github.com/ArminGh02/othello-bot/pkg/othellogame/color"
	"github.com/ArminGh02/othello-bot/pkg/othellogame/direction"
	"github.com/ArminGh02/othello-bot/pkg/othellogame/turn"
	"github.com/ArminGh02/othello-bot/pkg/util"
	"github.com/ArminGh02/othello-bot/pkg/util/coord"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/xid"
)

var offset = [direction.Count]coord.Coord{
	{-1, -1},
	{0, -1},
	{1, -1},
	{-1, 0},
	{1, 0},
	{-1, 1},
	{0, 1},
	{1, 1},
}

type Game struct {
	id              string
	users           [2]tgbotapi.User
	disksCount      [2]int
	board           [boardSize][boardSize]cell.Cell
	turn            turn.Turn
	placeableCoords util.CoordSet
	ended           bool
	whiteStarted    bool
	movesSequence   []coord.Coord
}

func New(user1, user2 *tgbotapi.User) *Game {
	g := &Game{
		id:              xid.New().String(),
		users:           [2]tgbotapi.User{*user1, *user2},
		disksCount:      [2]int{2, 2},
		turn:            turn.Random(),
		placeableCoords: util.NewCoordSet(),
		movesSequence:   make([]coord.Coord, 0, boardSize*boardSize-4),
	}

	g.whiteStarted = g.turn == turn.White

	mid := len(g.board)/2 - 1
	g.board[mid][mid] = cell.White
	g.board[mid][mid+1] = cell.Black
	g.board[mid+1][mid] = cell.Black
	g.board[mid+1][mid+1] = cell.White

	g.updatePlaceableCoords()

	return g
}

func (g *Game) String() string {
	return fmt.Sprintf(
		"Game between %s and %s",
		util.UsernameElseName(&g.users[0]),
		util.UsernameElseName(&g.users[1]),
	)
}

func (g *Game) ID() string {
	return g.id
}

func (g *Game) Board() [][]cell.Cell {
	res := make([][]cell.Cell, len(g.board))
	for i := range g.board {
		res[i] = g.board[i][:]
	}
	return res
}

func (g *Game) ActiveColor() string {
	return g.turn.Cell().Emoji()
}

func (g *Game) ActiveUser() *tgbotapi.User {
	return &g.users[g.turn.Int()]
}

func (g *Game) WhiteUser() *tgbotapi.User {
	return &g.users[color.White]
}

func (g *Game) BlackUser() *tgbotapi.User {
	return &g.users[color.Black]
}

func (g *Game) WhiteDisks() int {
	return g.disksCount[color.White]
}

func (g *Game) BlackDisks() int {
	return g.disksCount[color.Black]
}

func (g *Game) IsEnded() bool {
	return g.ended
}

func (g *Game) Winner() *tgbotapi.User {
	if g.disksCount[color.White] == g.disksCount[color.Black] {
		return nil
	}
	if g.disksCount[color.White] > g.disksCount[color.Black] {
		return &g.users[color.White]
	}
	return &g.users[color.Black]
}

func (g *Game) Loser() *tgbotapi.User {
	winner := g.Winner()
	if winner == nil {
		return nil
	}
	return g.OpponentOf(winner)
}

func (g *Game) OpponentOf(user *tgbotapi.User) *tgbotapi.User {
	if *user == *g.WhiteUser() {
		return g.BlackUser()
	}
	if *user == *g.BlackUser() {
		return g.WhiteUser()
	}
	log.Panicln("Invalid state: OpponentOf called with an argument unequal to both game users.")
	panic("")
}

func (g *Game) WinnerColor() string {
	winner := g.Winner()
	if winner == nil {
		log.Panicln("Invalid state: WinnerColor called when the game is a draw.")
	}
	if *winner == g.users[color.White] {
		return cell.White.Emoji()
	}
	return cell.Black.Emoji()
}

func (g *Game) InlineKeyboard(showLegalMoves bool) [][]tgbotapi.InlineKeyboardButton {
	keyboard := make([][]tgbotapi.InlineKeyboardButton, len(g.board))
	for y := range g.board {
		keyboard[y] = make([]tgbotapi.InlineKeyboardButton, len(g.board[y]))
		for x, cell := range g.board[y] {
			var buttonText string
			if showLegalMoves && g.placeableCoords.Contains(coord.New(x, y)) {
				buttonText = consts.LegalMoveEmoji
			} else {
				buttonText = cell.Emoji()
			}

			keyboard[y][x] = tgbotapi.NewInlineKeyboardButtonData(
				buttonText,
				fmt.Sprintf("%d_%d", x, y),
			)
		}
	}
	return keyboard
}

func (g *Game) EndInlineKeyboard() [][]tgbotapi.InlineKeyboardButton {
	keyboard := make([][]tgbotapi.InlineKeyboardButton, len(g.board))
	for y := range g.board {
		keyboard[y] = make([]tgbotapi.InlineKeyboardButton, len(g.board[y]))
		for x, cell := range g.board[y] {
			keyboard[y][x] = tgbotapi.NewInlineKeyboardButtonData(
				cell.Emoji(),
				"gameOver",
			)
		}
	}
	return keyboard
}

func (g *Game) WhiteStarted() bool {
	return g.whiteStarted
}

func (g *Game) MovesSequence() []coord.Coord {
	return g.movesSequence
}

func (g *Game) SetTurn(white bool) {
	g.turn = turn.Turn(!white)
	if len(g.movesSequence) == 0 {
		g.whiteStarted = white
	}
}

func (g *Game) PlaceDisk(where coord.Coord, user *tgbotapi.User) error {
	if err := g.checkPlacingDisk(where, user); err != nil {
		return err
	}
	g.PlaceDiskUnchecked(where)
	return nil
}

func (g *Game) PlaceDiskUnchecked(where coord.Coord) {
	g.board[where.Y][where.X] = g.turn.Cell()
	g.flipDisks(where)

	for i := 0; i < 2; i++ {
		g.passTurn()
		g.updatePlaceableCoords()
		if !g.placeableCoords.IsEmpty() {
			break
		}
	}

	if g.placeableCoords.IsEmpty() {
		g.ended = true
	}

	g.movesSequence = append(g.movesSequence, where)
}

func (g *Game) IsTurnOf(user *tgbotapi.User) bool {
	return *g.ActiveUser() == *user
}

func (g *Game) checkPlacingDisk(where coord.Coord, user *tgbotapi.User) error {
	if !g.IsTurnOf(user) {
		return fmt.Errorf("It's not your turn!")
	}
	if g.board[where.Y][where.X] != cell.Empty {
		return fmt.Errorf("That cell is not empty!")
	}
	if !g.placeableCoords.Contains(where) {
		return fmt.Errorf("You can't place a disk there!")
	}
	return nil
}

func (g *Game) flipDisks(where coord.Coord) {
	opponent := g.turn.Cell().Reversed()
	directionsToFlip := g.findDirectionsToFlip(where, false)
	for _, dir := range directionsToFlip {
		c := coord.Plus(where, offset[dir])
		for g.board[c.Y][c.X] == opponent {
			g.board[c.Y][c.X] = g.turn.Cell()
			c.Plus(offset[dir])
		}
	}
	g.updateDisksCount()
}

func (g *Game) findDirectionsToFlip(
	where coord.Coord,
	mustBeEmptyCell bool,
) []direction.Direction {
	opponent := g.turn.Cell().Reversed()
	res := make([]direction.Direction, 0, direction.Count)

	if mustBeEmptyCell && g.board[where.Y][where.X] != cell.Empty {
		return res
	}

	for dir := direction.NorthWest; dir < direction.Count; dir++ {
		c := coord.Plus(where, offset[dir])
		if isValidCoord(c, len(g.board)) && g.board[c.Y][c.X] == opponent {
		loop:
			for {
				c.Plus(offset[dir])

				if !isValidCoord(c, len(g.board)) {
					break
				}

				switch g.board[c.Y][c.X] {
				case g.turn.Cell():
					res = append(res, dir)
					break loop
				case cell.Empty:
					break loop
				}
			}
		}
	}
	return res
}

func isValidCoord(c coord.Coord, length int) bool {
	return c.X >= 0 && c.Y >= 0 && c.X < length && c.Y < length
}

func (g *Game) passTurn() {
	g.turn = !g.turn
}

func (g *Game) updatePlaceableCoords() {
	g.placeableCoords.Clear()
	for y := range g.board {
		for x := range g.board[y] {
			if c := coord.New(x, y); g.isPlaceableCoord(c) {
				g.placeableCoords.Insert(c)
			}
		}
	}
}

func (g *Game) isPlaceableCoord(where coord.Coord) bool {
	return len(g.findDirectionsToFlip(where, true)) > 0
}

func (g *Game) updateDisksCount() {
	white, black := 0, 0
	for _, row := range g.board {
		for _, c := range row {
			switch c {
			case cell.White:
				white++
			case cell.Black:
				black++
			}
		}
	}
	g.disksCount[color.White] = white
	g.disksCount[color.Black] = black
}
