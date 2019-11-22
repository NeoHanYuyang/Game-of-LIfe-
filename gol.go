package main

import (
	"fmt"
	"strconv"
	"strings"
)



//create a worker function that can work in parallel
func worker(start, end, n int, cc, up, down chan[]byte, p golParams, world[][]byte, temp[][]byte) {

}

// take out the game logic from the distributor and make it a function that would come in handy
func CheckNeighborAndFlip(p golParams, world[][]byte, temp[][]byte) {
	// Calculate the new state of Game of Life after the given number of turns.
	for turns := 0; turns < p.turns; turns++ {
		for y := 0; y < p.imageHeight; y++ {
			for x := 0; x < p.imageWidth; x++ {
				//actual Game of Life logic: flips alive cells to dead and dead cells to alive.
				//set the variable name
				aliveNum := 0
				h := p.imageHeight
				w := p.imageWidth
				//for loop to check the 8 neighbors of the given point
				for i := y - 1; i < y+2; i++ {
					for j := x - 1; j < x+2; j++ {
						if i == y && j == x {
							continue
						}
						if world[(i+h)%h][(j+w)%w] == 0xFF {
							aliveNum++
						}
					}
				}
				//if the given point is alive
				if world[y][x] != 0 {
					if aliveNum < 2 || aliveNum > 3 {
						temp[y][x] = 0
					} else {
						temp[y][x] = world[y][x]
					}
				}
				//if the given point is dead
				if world[y][x] == 0 {
					if aliveNum == 3 {
						temp[y][x] = 255
					} else {
						temp[y][x] = 0
					}
				}
			}
		}
		for y := 0; y < p.imageHeight; y++ {
			for x := 0; x < p.imageWidth; x++ {
				world[y][x] = temp[y][x]
			}
		}
	}
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p golParams, d distributorChans, alive chan []cell) {

	// Create the 2D slice to store the world.
	world := make([][]byte, p.imageHeight)
	for i := range world {
		world[i] = make([]byte, p.imageWidth)
	}

	// Request the io goroutine to read in the image with the given filename.
	d.io.command <- ioInput
	d.io.filename <- strings.Join([]string{strconv.Itoa(p.imageWidth), strconv.Itoa(p.imageHeight)}, "x")

	// The io goroutine sends the requested image byte by byte, in rows.
	for y := 0; y < p.imageHeight; y++ {
		for x := 0; x < p.imageWidth; x++ {
			val := <-d.io.inputVal
			if val != 0 {
				fmt.Println("Alive cell at", x, y)
				world[y][x] = val
			}
		}
	}

	//create channels to pass array around
	//cc   := make(chan byte)
	//up   := make(chan byte)
	//down := make(chan byte)

	//make a temp world to store the point
	temp := make([][]byte, p.imageHeight)
	for i := range temp {
		temp[i] = make([]byte, p.imageWidth)
	}
	

	// Calculate the new state of Game of Life after the given number of turns.
	//for turns := 0; turns < p.turns; turns++ {
	//
	//
		//function call game logic
		CheckNeighborAndFlip(p, world, temp)
	//}

	// Create an empty slice to store coordinates of cells that are still alive after p.turns are done.
	var finalAlive []cell
	// Go through the world and append the cells that are still alive.
	for y := 0; y < p.imageHeight; y++ {
		for x := 0; x < p.imageWidth; x++ {
			if world[y][x] != 0 {
				finalAlive = append(finalAlive, cell{x: x, y: y})
			}
		}
	}

	//output the board by receiving from signaled output
	d.io.command <- ioOutput
	d.io.filename <- strings.Join([]string{strconv.Itoa(p.imageWidth), strconv.Itoa(p.imageHeight), strconv.Itoa(p.turns)}, "x")

	//send the world to finalBoard so we will be able to print it
	d.io.finalBoard <- world

	// Make sure that the Io has finished any output before exiting.
	d.io.command <- ioCheckIdle
	<-d.io.idle

	// Return the coordinates of cells that are still alive.
	alive <- finalAlive
}
