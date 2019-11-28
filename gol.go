
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

//function to check the alive cells around the given point
func NeighAlive(world [][]byte,y, x int, imageWidth int) int {
	var aliveNum = 0
	//h := workerHeight
	w := imageWidth
	//for loop to check the 8 neighbors of the given point
	for i := y-1; i < y+2; i++ {
		for j := x-1; j < x+2; j++ {
			if i == y && j == x {continue}
			if world[i][(j+w)%w] == 255 {aliveNum++}
		}
	}
	return aliveNum
}

//// take out the game logic from the distributor and make it a function that would come in handy
//func CheckNeighborAndFlip( world[][]byte, temp[][]byte, imageHeight, imageWidth int) {
//
//	for y := 1; y <= imageHeight; y++ {
//		for x := 0; x < imageWidth; x++ {
//			//actual Game of Life logic: flips alive cells to dead and dead cells to alive.
//			//set the variable name
//			var aliveNum = 0
//			aliveNum = Num(world, imageHeight, imageWidth)
//			////for loop to check the 8 neighbors of the given point
//			//for i := y - 1; i < y+2; i++ {
//			//	for j := x - 1; j < x+2; j++ {
//			//		if i == y && j == x {
//			//			continue
//			//		}
//			//		if world[(i+h)%h][(j+w)%w] == 0xFF {
//			//			aliveNum++
//			//		}
//			//	}
//			//}
//			//if the given point is alive
//			if world[y][x] == 255 {
//				if aliveNum < 2 || aliveNum > 3 {
//					temp[y][x] = 0
//				} else {
//					temp[y][x] = world[y][x]
//				}
//			}
//			//if the given point is dead
//			if world[y][x] == 0 {
//				if aliveNum == 3 {
//					temp[y][x] = 255
//				} else {
//					temp[y][x] = 0
//				}
//			}
//		}
//	}
//	//for y := 0; y < p.imageHeight; y++ {
//	//	for x := 0; x < p.imageWidth; x++ {
//	//		world[y][x] = temp[y][x]
//	//	}
//	//}
//}

//a function form our new world for corresponding threads
func buildNewWorld (world [][]byte, SectionHeight, imageHeight, imageWidth, totalThreads, currentThreads int) [][] byte{
	newWorld := make([][]byte, SectionHeight+2)
	for j := 0;j<SectionHeight+2; j++ {
		newWorld[j] = make([]byte, imageWidth)
	}

	if currentThreads==0{
		for x := 0; x < imageWidth; x++ {
			newWorld[0][x]=world[imageHeight-1][x]
		}
	}else{
		for x := 0; x < imageWidth; x++ {
			newWorld[0][x]=world[currentThreads*SectionHeight-1][x]
		}
	}

	for y := 1; y <= SectionHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			newWorld[y][x]=world[currentThreads*SectionHeight+y-1][x]
		}
	}

	if currentThreads==totalThreads-1{
		for x := 0; x < imageWidth; x++ {
			newWorld[SectionHeight+1][x]=world[0][x]
		}
	}else {
		for x := 0; x < imageWidth; x++ {
			newWorld[SectionHeight+1][x]=world[(currentThreads+1)*SectionHeight][x]
		}
	}

	return newWorld
}


func worker(world [][]byte, imageHeight int,imageWidth int,  out chan<- [][]byte){
	temp := make([][]byte, imageHeight+2)
	for i:= range world {
		temp[i] = make([]byte, imageWidth)
	}

	for y := 1; y <= imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			var aliveNum = 0
			aliveNum = NeighAlive(world,y,x,imageWidth)
			//if the given point is alive
			if world[y][x] == 255 {
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
	out <- temp
}




// distributor divides the work between workers and interacts with other goroutines.
func distributor(p golParams, d distributorChans, alive chan []cell, state chan rune, pause chan rune, quit chan rune) {

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

	ticker := time.NewTicker(2 * time.Second)

	//Calculate the new state of Game of Life after the given number of turns.
	for turns := 0; turns < p.turns; turns++ {




		SectionHeight := p.imageHeight / p.threads

		var out [8] chan [][]byte

		for i := 0; i < p.threads; i++ {
			out[i] = make(chan [][]byte)
			newWorld := buildNewWorld(world, SectionHeight, p.imageHeight, p.imageWidth, p.threads, i)
			go worker(newWorld, SectionHeight, p.imageWidth, out[i])
		}
		for i := 0; i < p.threads; i++ {
			tempOut := <-out[i]
			//println("tempOut  i=",i)
			for y := 0; y < SectionHeight; y++ {
				for x := 0; x < p.imageWidth; x++ {
					//print(tempOut[y+1][x])
					world[i*SectionHeight+y][x] = tempOut[y+1][x]
				}
			}
		}

		select  {

		case <- state :
			d.io.command <- ioOutput
			d.io.filename <- strings.Join([]string{strconv.Itoa(p.imageWidth), strconv.Itoa(p.imageHeight), strconv.Itoa(turns)}, "x")
			//send the world to finalBoard so we will be able to print it
			d.io.finalBoard <- world


			    case <- pause :
			println(" the current turn is ", turns)
			//busy waiting that pause the program

				//busy waiting
				//if 'p' != <- pause  {
				//	//do nothing until we have a second press and it is p
				//}
				if  <-pause  == 'p' {
					fmt.Print("continuing ...\n")
					break
				}

				case <- quit :

			d.io.command <- ioOutput
			d.io.filename <- strings.Join([]string{strconv.Itoa(p.imageWidth), strconv.Itoa(p.imageHeight), strconv.Itoa(turns)}, "x")
			//send the world to finalBoard so we will be able to print it
			d.io.finalBoard <- world
			StopControlServer()
			os.Exit(0)

			case <- ticker.C :
				var finalAlive []cell
				// Go through the world and append the cells that are still alive.
				for y := 0; y < p.imageHeight; y++ {
					for x := 0; x < p.imageWidth; x++ {
						if world[y][x] != 0 {
							finalAlive = append(finalAlive, cell{x: x, y: y})
						}
					}
				}
				fmt.Println("NuMbEr oF aLiVe CeLls :", len(finalAlive))
			default:
			continue
		}
	}

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

