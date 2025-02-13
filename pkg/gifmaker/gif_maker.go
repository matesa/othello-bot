package gifmaker

import (
	"image"
	"image/draw"
	"image/gif"
	"log"
	"os"

	"github.com/ArminGh02/othello-bot/pkg/othellogame"
	"github.com/ArminGh02/othello-bot/pkg/othellogame/cell"
	"github.com/ArminGh02/othello-bot/pkg/util/coord"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	x0 = 77
	y0 = 121
)

const (
	diskLength = 39
	cellLength = 44
)

var (
	whiteDisk  = readPNG("resources/white-disk.png")
	blackDisk  = readPNG("resources/black-disk.png")
	boardImage = imageToPaletted(readPNG("resources/board.png"))
)

func Make(outputFilename string, movesSequence []coord.Coord, whiteStarts bool) {
	frames := getGameFrames(movesSequence, whiteStarts)
	frames = append(frames, frames[len(frames) - 1])
	delays := make([]int, len(frames))
	for i := range delays {
		delays[i] = 200
	}

	out, err := os.Create(outputFilename)
	if err != nil {
		log.Panicln(err)
	}
	defer out.Close()
	gif.EncodeAll(out, &gif.GIF{
		Image: frames,
		Delay: delays,
	})
}

func getGameFrames(movesSequence []coord.Coord, whiteStarts bool) []*image.Paletted {
	game := othellogame.New(&tgbotapi.User{}, &tgbotapi.User{})
	game.SetTurn(whiteStarts)

	res := make([]*image.Paletted, 0, len(movesSequence))
	for _, move := range movesSequence {
		res = append(res, getGameFrame(game))
		game.PlaceDiskUnchecked(move)
	}
	return res
}

func getGameFrame(game *othellogame.Game) *image.Paletted {
	getDiskImage := func(white bool) image.Image {
		if white {
			return whiteDisk
		}
		return blackDisk
	}

	res := cloneImage(boardImage)
	board := game.Board()
	for i := range board {
		for j := range board[i] {
			if board[i][j] == cell.Empty {
				continue
			}

			x := x0 + j*cellLength
			y := y0 + i*cellLength
			draw.Draw(
				res,
				image.Rect(x, y, x+diskLength, y+diskLength),
				getDiskImage(board[i][j] == cell.White),
				image.Point{},
				draw.Over,
			)
		}
	}
	return res
}
