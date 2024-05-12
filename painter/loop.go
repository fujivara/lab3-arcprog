package painter

import (
	"github.com/roman-mazur/architecture-lab-3/ui"
	"image"
	"image/color"
	"sync"

	"golang.org/x/exp/shiny/screen"
)

type Receiver interface {
	Update(t screen.Texture)
}

type UI struct {
	bg       color.Color
	bgFigure [2]image.Point
	crosses  []*ui.Cross
}

type Loop struct {
	Receiver Receiver

	buffer screen.Texture
	state  UI

	mq messageQueue

	stop    chan struct{}
	stopReq bool
}

var size = image.Pt(800, 800)

func (l *Loop) Start(s screen.Screen) {
	l.buffer, _ = s.NewTexture(size)
	l.stop = make(chan struct{})
	l.state = UI{
		bg:       color.Black,
		bgFigure: [2]image.Point{{0, 0}, {0, 0}},
		crosses:  []*ui.Cross{},
	}

	go func() {
		for !(l.stopReq && l.mq.empty()) {
			op := l.mq.pull()
			update := op.Do(l.buffer, &l.state)

			if update {
				l.Receiver.Update(l.buffer)
			}
		}
		close(l.stop)
	}()
}

func (l *Loop) Post(op Operation) {
	l.mq.push(op)
}

func (l *Loop) StopAndWait() {
	l.Post(OperationFunc(func(screen.Texture, *UI) {
		l.stopReq = true
	}))
	<-l.stop
}

type messageQueue struct {
	operations []Operation
	mu         sync.Mutex
	signal     chan struct{}
}

func (mq *messageQueue) push(op Operation) {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	mq.operations = append(mq.operations, op)
	if mq.signal != nil {
		close(mq.signal)
	}
}

func (mq *messageQueue) pull() Operation {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	if len(mq.operations) == 0 {
		mq.mu.Unlock()
		mq.signal = make(chan struct{})
		<-mq.signal
		mq.signal = nil
		mq.mu.Lock()
	}
	op := mq.operations[0]
	mq.operations[0] = nil
	mq.operations = mq.operations[1:]
	return op
}

func (mq *messageQueue) empty() bool {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	return len(mq.operations) == 0
}
