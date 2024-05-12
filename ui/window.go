package ui

import (
	"golang.org/x/exp/shiny/imageutil"
	"image"
	"image/color"
	"log"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/draw"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

type Cross struct {
	x, y, size, width int
}

func getCross(x int, y int) *Cross {
	return &Cross{
		x:     x,
		y:     y,
		size:  400,
		width: 100,
	}
}

func (cross *Cross) draw(t screen.Texture) {
	t.Fill(
		image.Rect(
			cross.x-cross.width/2,
			cross.x-cross.width/2,
			cross.y-cross.size/2,
			cross.y-cross.size/2,
		),
		color.RGBA{R: 255, G: 255, A: 255},
		draw.Src,
	)

	t.Fill(
		image.Rect(
			cross.x-cross.size/2,
			cross.x-cross.size/2,
			cross.y-cross.width/2,
			cross.y-cross.width/2,
		),
		color.RGBA{R: 255, G: 255, A: 255},
		draw.Src,
	)
}

func (cross *Cross) visualize(pw *Visualizer) {
	pw.w.Fill(
		image.Rect(
			cross.x,
			cross.y-cross.size/2+cross.width/2,
			cross.x-cross.size,
			cross.y-cross.size/2-cross.width/2,
		),
		color.RGBA{R: 255, G: 255, A: 255},
		draw.Src,
	)

	pw.w.Fill(
		image.Rect(
			cross.x+cross.size/2+cross.width/2,
			cross.y,
			cross.x+cross.size/2-cross.width/2,
			cross.y-cross.size,
		),
		color.RGBA{R: 255, G: 255, A: 255},
		draw.Src,
	)
}

func (cross *Cross) position(x int, y int) {
	cross.x = x - cross.size/2
	cross.y = y - cross.size/2
}

type Visualizer struct {
	Title         string
	Debug         bool
	OnScreenReady func(s screen.Screen)

	w    screen.Window
	tx   chan screen.Texture
	done chan struct{}

	sz  size.Event
	pos image.Rectangle

	startSize int
	crosses   []*Cross
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
	w.Send(paint.Event{})
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
	pw.crosses = []*Cross{{200, 200, 400, 100}}
	pw.startSize = 400

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
				pw.positionAll(int(e.X), int(e.Y))
				pw.w.Send(paint.Event{})
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
	pw.w.Fill(pw.sz.Bounds(), color.White, draw.Src)

	for _, cross := range pw.crosses {
		cross.visualize(pw)
	}

	for _, border := range imageutil.Border(pw.sz.Bounds(), 10) {
		pw.w.Fill(border, color.Black, draw.Src)
	}
}

func (pw *Visualizer) positionAll(x int, y int) {
	for _, cross := range pw.crosses {
		cross.position(x, y)
	}
}

func (pw *Visualizer) add(x int, y int) {
	cross := &Cross{
		x - pw.startSize/2,
		y - pw.startSize,
		pw.startSize,
		100,
	}
	pw.crosses = append(pw.crosses, cross)
}
