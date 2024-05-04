package ui

import (
	"image"
	"image/color"
	"log"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/imageutil"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/draw"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

type Visualizer struct {
	Title         string
	Debug         bool
	OnScreenReady func(s screen.Screen)

	w    screen.Window
	tx   chan screen.Texture
	done chan struct{}

	sz  size.Event
	pos image.Rectangle
}

func (pw *Visualizer) Main() {
	pw.tx = make(chan screen.Texture)
	pw.done = make(chan struct{})
	pw.pos.Max.X = 200
	pw.pos.Max.Y = 200
	driver.Main(pw.run)
}

func (pw *Visualizer) Update(t screen.Texture) {
	pw.tx <- t
}

func (pw *Visualizer) run(s screen.Screen) {
	w, err := s.NewWindow(&screen.NewWindowOptions{
		Title:  pw.Title,
		Width:  800,
		Height: 800,
	})
	if err != nil {
		log.Fatal("Failed to initialize the app window:", err)
	}
	defer func() {
		w.Release()
		close(pw.done)
	}()

	if pw.OnScreenReady != nil {
		pw.OnScreenReady(s)
	}

	pw.w = w

	events := make(chan any)
	go func() {
		for {
			e := w.NextEvent()
			if pw.Debug {
				log.Printf("new event: %v", e)
			}
			if detectTerminate(e) {
				close(events)
				break
			}
			events <- e
		}
	}()

	var t screen.Texture

	for {
		select {
		case e, ok := <-events:
			if !ok {
				return
			}
			pw.handleEvent(e, t)

		case t = <-pw.tx:
			w.Send(paint.Event{})
		}
	}
}

func detectTerminate(e any) bool {
	switch e := e.(type) {
	case lifecycle.Event:
		if e.To == lifecycle.StageDead {
			return true // Window destroy initiated.
		}
	case key.Event:
		if e.Code == key.CodeEscape {
			return true // Esc pressed.
		}
	}
	return false
}

func (pw *Visualizer) handleEvent(e any, t screen.Texture) {
	switch e := e.(type) {

	case size.Event: // Оновлення даних про розмір вікна.
		pw.sz = e

	case error:
		log.Printf("ERROR: %s", e)

	case mouse.Event:
		if t == nil {
			if e.Button == mouse.ButtonRight && e.Direction == mouse.DirPress {
				pw.w.Fill(pw.sz.Bounds(), color.White, draw.Src)

				squareWidth := 400
				squareHeight := 200

				centerPoint := image.Pt(int(e.X), int(e.Y))

				posX1 := centerPoint.X - squareWidth/2
				posY1 := centerPoint.Y - squareHeight/2

				posX2 := centerPoint.X - squareHeight/2
				posY2 := centerPoint.Y - squareWidth/2

				drawCrossByCoords(pw, posX1, posY1, posX2, posY2, squareHeight, squareWidth)
			}

		}

	case paint.Event:
		// Малювання контенту вікна.
		if t == nil {
			pw.drawDefaultUI()
		} else {
			// Використання текстури отриманої через виклик Update.
			pw.w.Scale(pw.sz.Bounds(), t, t.Bounds(), draw.Src, nil)
		}
		pw.w.Publish()
	}
}

func (pw *Visualizer) drawDefaultUI() {
	pw.w.Fill(pw.sz.Bounds(), color.White, draw.Src) // Білий фон.

	// Розміри квадратів.
	squareWidth := 400
	squareHeight := 200

	// Розміри вікна.
	winWidth := pw.sz.Bounds().Dx()
	winHeight := pw.sz.Bounds().Dy()

	// Обчислення позицій по центру вікна.
	posX1 := (winWidth - squareWidth) / 2
	posY1 := (winHeight - squareHeight) / 2

	posX2 := (winWidth - squareHeight) / 2
	posY2 := (winHeight - squareWidth) / 2

	drawCrossByCoords(pw, posX1, posY1, posX2, posY2, squareHeight, squareWidth)
}

func drawCrossByCoords(
	pw *Visualizer,
	posX1 int,
	posY1 int,
	posX2 int,
	posY2 int,
	squareHeight int,
	squareWidth int,
) {
	yellow := color.RGBA{R: 255, G: 255, B: 0, A: 255}

	// Малювання першого квадрата.
	square1 := image.Rect(posX1, posY1, posX1+squareWidth, posY1+squareHeight)
	pw.w.Fill(square1, yellow, draw.Src)

	// Малювання другого квадрата.
	square2 := image.Rect(posX2, posY2, posX2+squareHeight, posY2+squareWidth)
	pw.w.Fill(square2, yellow, draw.Src)

	// Малювання білої рамки.
	for _, br := range imageutil.Border(pw.sz.Bounds(), 10) {
		pw.w.Fill(br, color.White, draw.Src)
	}
}
