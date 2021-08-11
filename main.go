package main

import (
	"math/rand"
	"sync"
	"time"

	termbox "github.com/nsf/termbox-go"
)

const (
	Up = iota
	Down
	Left
	Right
)

type Point struct {
	x, y int
}

var direction int = Up
var snakePos []Point        // store all Points in the snake
var wallPos []Point         // store all Points in the wall
var applePos []Point        // store all Points for apples
var speed = 200             // initial speed (delay between moves) in milliseconds
var updatedPosition = false // makes

func tbPrint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
}
}

func drawWalls() {
	w, h := termbox.Size()
	y := 0

	// store wall Points in wallPos so we can check this slice for collision
	for x := 0; x < w; x++ {
		wallPos = append(wallPos, Point{x, y})     // top wall
		wallPos = append(wallPos, Point{x, h - 1}) // bottom wall
	}
	y = 1
	for y < h-1 {
		wallPos = append(wallPos, Point{0, y})     // left wall
		wallPos = append(wallPos, Point{w - 1, y}) // right wall
		y++
	}
	for _, sp := range wallPos {
		termbox.SetCell(sp.x, sp.y, '#', termbox.ColorRed, termbox.ColorDefault)
	}

	termbox.Flush()
}

func drawSnake(wg *sync.WaitGroup, w, h int) {
	for {
		for i := 0; i < len(snakePos); i++ {
			termbox.SetCell(snakePos[i].x, snakePos[i].y, '@', termbox.ColorWhite, termbox.ColorDefault)
		}

		termbox.Flush()

		time.Sleep(time.Duration(speed) * time.Millisecond)

		// save the last position of the snake for later
		lastPos := Point{snakePos[len(snakePos)-1].x, snakePos[len(snakePos)-1].y}

		// move
		switch direction {
		case Up:
			// add new position in the direction we're going and remove the last position
			snakePos = append([]Point{Point{snakePos[0].x, snakePos[0].y - 1}}, snakePos[0:len(snakePos)-1]...)
		case Down:
			snakePos = append([]Point{Point{snakePos[0].x, snakePos[0].y + 1}}, snakePos[0:len(snakePos)-1]...)
		case Left:
			snakePos = append([]Point{Point{snakePos[0].x - 1, snakePos[0].y}}, snakePos[0:len(snakePos)-1]...)
		case Right:
			snakePos = append([]Point{Point{snakePos[0].x + 1, snakePos[0].y}}, snakePos[0:len(snakePos)-1]...)
		}
		updatedPosition = false

		if checkCrash(Point{snakePos[0].x, snakePos[0].y}) {
			gameOver := "Game Over!"
			tbPrint(w/2-len(gameOver)/2, h/2, termbox.ColorRed, termbox.ColorDefault, gameOver)
			termbox.Flush()
			time.Sleep(3 * time.Second)
			termbox.Interrupt()
			wg.Done()
			return
		}

		if checkApple(Point{snakePos[0].x, snakePos[0].y}) {
			// add the old last position again, so the snake grows
			snakePos = append(snakePos, lastPos)
			increaseSpeed()
			drawApple(w, h, snakePos)
		} else {
			// remove old last position from screen
			termbox.SetCell(lastPos.x, lastPos.y, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}

func drawApple(w, h int, snakePos []Point) {

	var x, y int

retry:
	for {
		x = rand.Intn(w-2) + 1
		y = rand.Intn(h-2) + 1

		for _, sp := range snakePos {
			if (Point{x, y}) == sp {
				continue retry
			}
		}
		break
	}
	termbox.SetCell(x, y, '', termbox.ColorGreen, termbox.ColorDefault)
	applePos = append(applePos, Point{x, y})
}

func checkCrash(p Point) bool {
	// Check walls
	for _, sp := range wallPos {
		if p == sp {
			return true
		}
	}

	// Check snake
	if len(snakePos) > 1 {
		for _, sp := range snakePos[1:] {
			if p == sp {
				return true
			}
		}
	}

	return false
}

func checkApple(p Point) bool {
	for i, sp := range applePos {
		if p == sp {
			// remove the apple we collided with
			applePos = append(applePos[:i], applePos[i+1:]...)
			return true
		}
	}

	return false
}

func increaseSpeed() {
	speed -= 10
	if speed <= 0 {
		speed = 10
	}
}

func controlSnake(wg *sync.WaitGroup, w, h int) {
loop:
	for {
		ev := termbox.PollEvent()
		switch ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break loop
			case termbox.KeyArrowUp:
				if (direction != Down || len(snakePos) == 1) && !updatedPosition {
					direction = Up
					updatedPosition = true
				}
			case termbox.KeyArrowDown:
				if (direction != Up || len(snakePos) == 1) && !updatedPosition {
					direction = Down
					updatedPosition = true
				}
			case termbox.KeyArrowLeft:
				if (direction != Right || len(snakePos) == 1) && !updatedPosition {
					direction = Left
					updatedPosition = true
				}
			case termbox.KeyArrowRight:
				if (direction != Left || len(snakePos) == 1) && !updatedPosition {
					direction = Right
					updatedPosition = true
				}
			}
		case termbox.EventInterrupt:
			break loop
		}

		if ev.Ch == '-' {
			speed += 10
		} else if ev.Ch == '+' {
			increaseSpeed()
		} else if ev.Ch == 'n' {
			drawApple(w, h, snakePos)
		}
	}
	wg.Done()
}

func main() {
	rand.Seed(time.Now().UnixNano())

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputEsc)

	w, h := termbox.Size()
	snakePos = append(snakePos, Point{w / 2, h / 2})

	var wg sync.WaitGroup

	drawWalls()
	drawApple(w, h, snakePos)

	// add only one to WaitGroup as whatever gorputing finishes first should exit the program
	wg.Add(1)

	go drawSnake(&wg, w, h)
	go controlSnake(&wg, w, h)

	wg.Wait()

	termbox.Close()
}
