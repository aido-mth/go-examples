package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// if 1, then cell is alive 0 otherwise
var (
	currentField [][]bool
	prevField    [][]bool
)

func main() {
	port := os.Args[2]
	currentField = make([][]bool, 100)
	prevField = make([][]bool, 100)
	for i := range currentField {
		currentField[i] = make([]bool, 100)
		prevField[i] = make([]bool, 100)
	}
	randomizeField()
	http.HandleFunc("/", handler)              // each request calls handler
	http.HandleFunc("/favicon.ico", doNothing) // to avoid calling handler twice
	log.Fatal(http.ListenAndServe("localhost:"+port, nil))
}

func doNothing(w http.ResponseWriter, r *http.Request) {}
func handler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// Pixel scale
	scale, err := strconv.Atoi(query.Get("scale"))
	if err != nil || scale <= 0 || scale > 20 {
		scale = 1
	}

	w.Header().Add("Content-Type", "image/png")
	start := time.Now()
	img := drawEpoch(scale)
	fmt.Printf("Epoch time: %dms\n", time.Since(start))
	_ = png.Encode(w, img)
}

func drawEpoch(scale int) (img *image.RGBA) {
	width := 100 * scale
	height := 100 * scale
	// Creating empty image
	img = image.NewRGBA(image.Rect(0, 0, width, height))

	// Filling image
	var n sync.WaitGroup
	_ = copy(prevField, currentField)
	for i := range prevField {
		n.Add(1)
		go drawLine(i, currentField[i], img, scale, &n)
	}
	n.Wait()
	return img
}

func randomizeField() {
	for _, row := range currentField {
		for i := range row {
			s1 := rand.NewSource(time.Now().UnixNano())
			r1 := rand.New(s1)
			num := r1.Intn(100)
			row[i] = num == 1 || num == 9 || num == 14 || num == 29 || num == 46 || num == 93
		}
	}
}

func drawLine(rowNum int, currentRow []bool, img *image.RGBA, scale int, n *sync.WaitGroup) {
	defer n.Done()
	for j, cell := range currentRow {
		currentRow[j] = checkCell(rowNum, j, cell)
		for row := 0; row < scale; row++ {
			for col := 0; col < scale; col++ {
				if currentRow[j] {
					img.Set(rowNum*scale+row, j*scale+col, color.RGBA{R: 0, G: 0, B: 0, A: 0xff})
				} else {
					img.Set(rowNum*scale+row, j*scale+col, color.RGBA{R: 255, G: 255, B: 255, A: 0xff})
				}
			}
		}
	}
}

func checkCell(rowNum int, colNum int, cell bool) bool {
	count := 0
	if rowNum-1 >= 0 {
		if colNum-1 >= 0 {
			count += boolToInt(prevField[rowNum-1][colNum-1])
		}
		count += boolToInt(prevField[rowNum-1][colNum])
		if colNum+1 < 100 {
			count += boolToInt(prevField[rowNum-1][colNum+1])
		}
	}
	if colNum-1 >= 0 {
		count += boolToInt(prevField[rowNum][colNum-1])
	}
	if colNum+1 < 100 {
		count += boolToInt(prevField[rowNum][colNum+1])
	}

	if rowNum+1 < 100 {
		if colNum-1 >= 0 {
			count += boolToInt(prevField[rowNum+1][colNum-1])
		}
		if colNum+1 < 100 {
			count += boolToInt(prevField[rowNum+1][colNum+1])
		}
		count += boolToInt(prevField[rowNum+1][colNum])
	}

	if !cell {
		return count == 3
	}
	return count == 2 || count == 3
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
