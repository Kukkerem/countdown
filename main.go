package main

import (
	"fmt"
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

const (
	usage = `usage:
countdown 25s [-up] [title]
countdown 1m50s [-up] [title]
countdown 2h45m50s [-up] [title]
`
	tick = time.Second
)

var (
	timer          *time.Timer
	ticker         *time.Ticker
	queues         chan termbox.Event
	startDone      bool
	startX, startY int
	titleX, titleY int
	title          string
)

func draw(d time.Duration, title Text) {
	w, h := termbox.Size()
	clear()

	str := format(d)
	text := toText(str)

	if !startDone {
		startDone = true
		startX, startY = w/2-text.width()/2, h/2-text.height()/2
		titleX, titleY = w/2-title.width()/2, h/2-2*title.height()-text.height()/2
	}

	x, y := startX, startY
	tx, ty := titleX, titleY
	for _, t := range title {
		echo(t, tx, ty)
	}

	for _, s := range text {
		echo(s, x, y)
		x += s.width()
	}

	flush()
}

func format(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h < 1 {
		return fmt.Sprintf("%02d:%02d", m, s)
	}
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func start(d time.Duration) {
	timer = time.NewTimer(d)
	ticker = time.NewTicker(tick)
}

func stop() {
	timer.Stop()
	ticker.Stop()
}

func countdown(timeLeft time.Duration, countUp bool) {
	var exitCode int
	t := make(Text, 0)
	t = append(t, []string{title})

	start(timeLeft)

	if countUp {
		timeLeft = 0
	}

	draw(timeLeft, t)

loop:
	for {
		select {
		case ev := <-queues:
			if ev.Type == termbox.EventKey && (ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC) {
				exitCode = 1
				break loop
			}
			if ev.Ch == 'p' || ev.Ch == 'P' {
				stop()
			}
			if ev.Ch == 'c' || ev.Ch == 'C' {
				start(timeLeft)
			}
		case <-ticker.C:
			if countUp {
				timeLeft += time.Duration(tick)
			} else {
				timeLeft -= time.Duration(tick)
			}
			draw(timeLeft, t)
		case <-timer.C:
			break loop
		}
	}

	termbox.Close()
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

func main() {
	if len(os.Args) < 2 || len(os.Args) > 4 {
		stderr(usage)
		os.Exit(2)
	}

	duration, err := time.ParseDuration(os.Args[1])
	if err != nil {
		stderr("error: invalid duration: %v\n", os.Args[1])
		os.Exit(2)
	}
	timeLeft := duration

	err = termbox.Init()
	if err != nil {
		panic(err)
	}

	queues = make(chan termbox.Event)
	go func() {
		for {
			queues <- termbox.PollEvent()
		}
	}()
	countUp := len(os.Args) >= 3 && os.Args[2] == "-up"

	switch len(os.Args) {
	case 3:
		if os.Args[2] != "-up" {
			title = os.Args[2]
		}
	case 4:
		title = os.Args[3]
	}

	countdown(timeLeft, countUp)
}
