package painter

import (
	"fmt"
	"github.com/roman-mazur/architecture-lab-3/ui"
	"image"
	"image/color"
	"strconv"

	"golang.org/x/exp/shiny/screen"
)

func getCoordsByArgs(width int, height int, args []float64) ([]int, error) {
	if len(args)%2 != 0 {
		return nil, fmt.Errorf("invalid args count")
	}

	coords := make([]int, len(args))

	floatWidth := float64(width)
	floatHeight := float64(height)

	for index := range args {
		if index%2 == 0 {
			coords[index] = int(floatWidth * args[index])
		} else {
			coords[index] = int(floatHeight * args[index])
		}
	}

	return coords, nil
}

func convertArgs(args []string) ([]float64, error) {
	parsedArgs := make([]float64, len(args))
	for i, str := range args {
		num, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return nil, err
		}
		parsedArgs[i] = num
	}
	return parsedArgs, nil
}

type Operation interface {
	Do(t screen.Texture, state *UI) (ready bool)
}

type OperationList []Operation

func (ol OperationList) Do(t screen.Texture, state *UI) (ready bool) {
	for _, o := range ol {
		ready = o.Do(t, state) || ready
	}
	return
}

var UpdateOp = updateOp{}

type updateOp struct{}

func (op updateOp) Do(t screen.Texture, state *UI) bool {
	t.Fill(t.Bounds(), state.bg, screen.Src)
	t.Fill(image.Rectangle{
		Min: state.bgFigure[0],
		Max: state.bgFigure[1],
	}, color.Black, screen.Src)

	for _, cross := range state.crosses {
		cross.Draw(t)
	}

	return true
}

type OperationFunc func(t screen.Texture, state *UI)

func (f OperationFunc) Do(t screen.Texture, state *UI) bool {
	f(t, state)
	return false
}

func WhiteFill(t screen.Texture, state *UI) {
	state.bg = color.White
}

func GreenFill(t screen.Texture, state *UI) {
	state.bg = color.RGBA{G: 0xff, A: 0xff}
}

func Reset(t screen.Texture, state *UI) {
	state.bg = color.White
	state.bgFigure = [2]image.Point{{0, 0}, {0, 0}}
	state.crosses = []*ui.Cross{}
}

func DrawRectangle(args []string) OperationFunc {
	if len(args) != 4 {
		fmt.Println("Min amount of arguments in 4")
		return nil
	}
	floatArgs, err := convertArgs(args)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	return func(t screen.Texture, state *UI) {
		coords, err := getCoordsByArgs(t.Bounds().Dx(), t.Bounds().Dy(), floatArgs)
		if err == nil && len(coords) == 4 {
			state.bgFigure[0] = image.Point{X: coords[0], Y: coords[1]}
			state.bgFigure[1] = image.Point{X: coords[2], Y: coords[3]}
		}
	}
}

func Figure(args []string) OperationFunc {
	if len(args) != 2 {
		fmt.Println("Wrong amount of arguments to move figures")
		return nil
	}
	floatArgs, err := convertArgs(args)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return func(t screen.Texture, state *UI) {
		coords, err := getCoordsByArgs(t.Bounds().Dx(), t.Bounds().Dy(), floatArgs)
		if err == nil && len(coords) == 2 {
			cross := ui.GetCross(coords[0], coords[1])
			state.crosses = append(state.crosses, cross)

		}
	}
}

func Move(args []string) OperationFunc {
	if len(args) != 2 {
		fmt.Println("Min amount of arguments is 2")
	}
	floatArgs, err := convertArgs(args)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	return func(t screen.Texture, state *UI) {
		coords, err := getCoordsByArgs(t.Bounds().Dx(), t.Bounds().Dy(), floatArgs)
		if err == nil && len(coords) == 2 {
			cross := ui.GetCross(coords[0], coords[1])
			state.crosses = []*ui.Cross{cross}
		}
	}
}
