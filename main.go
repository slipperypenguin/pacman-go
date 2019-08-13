package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"time"
)

// hold the Player data. 2D coords (row, col)
type Player struct {
	row int
	col int
}

// Ghost is the enemy that chases the player :O
type Ghost struct {
    row int
    col int
}

// Config holds the emoji configuration
// public members for Config struct required for json decoder
type Config struct {
	Player   string `json:"player"`
	Ghost    string `json:"ghost"`
	Wall     string `json:"wall"`
	Dot      string `json:"dot"`
	Pill     string `json:"pill"`
	Death    string `json:"death"`
	Space    string `json:"space"`
	UseEmoji bool   `json:"use_emoji"`
}

// Global vars
var ghosts []*Ghost // slice of pointers to Ghost objects
var player Player
var cfg Config // stores parsed json
var maze []string
var score int
var numDots int
var lives = 1

// Global flags
// flag.String(NAME, DEFAULT_VAL, HELP_DESCRIP)
// return pointer to a string that holds the value of the flag
// value is only filled after calling flag.Parse, in main()
var (
	configFile = flag.String("config-file", "config.json", "path to custom configuration file")
	mazeFile   = flag.String("maze-file", "maze01.txt", "path to a custom maze file")
)


// parse the json and store it in the cfg global variable
func loadConfig() error  {
	// note dereference operator, since flags are pointers
	file, err := os.Open(*configFile)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return err
	}

	return nil
}

// Load the maze file into memory
func loadMaze() error {
	file, err := os.Open(*mazeFile)
	if err != nil {
		return err
	}
	// use defer to call at the end of loadMaze()
	defer file.Close() // puts file.Close() in the call stack

	// read the file line-by-line and append to maze slice
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		maze = append(maze, line)
	}

	// capture player position as soon as we load maze.
	// traverse each character of the maze and create a player when it locates a `P`
	for row, line := range maze {
		for col, char := range line {
			switch char {
			case 'P':
				player = Player{row, col}
			case 'G':
				// note: & oper. means that instad of adding a Ghost obj to the slice, we are adding a pointer to it
				// add pointer to ghost obj. instead of adding the obj to the slice
				ghosts = append(ghosts, &Ghost{row, col})
			case '.':
				numDots++
			}
		}
	}

	return nil
}

func printScreen() {
	// clear screen before printing, so we see
	clearScreen()
	// _ is just placeholder for where the compiler expects a var name
	// using _ means that we are ingoring the value
	// if we omit _, then the range only returns the index
	for _, line := range maze { // range returns: index,
		for _, chr := range line {
			switch chr {
			case '#':
				fmt.Printf(cfg.Wall)
			case '.':
				fmt.Printf(cfg.Dot)
			case 'X':
				fmt.Printf(cfg.Pill)
			default:
				fmt.Printf(cfg.Space)
			}
		}
		fmt.Printf("\n")
	}

	moveCursor(player.row, player.col)
	fmt.Printf(cfg.Player)

	for _, g := range ghosts {
		moveCursor(g.row, g.col)
		fmt.Printf(cfg.Ghost)
	}

	// print score
  moveCursor(len(maze)+1, 0)
  fmt.Printf("Score: %v\nLives: %v\n", score, lives)
}

func clearScreen() {
	// use escape sequences
	fmt.Printf("\x1b[2J")
	moveCursor(0, 0)
}

func moveCursor(row, col int) {
	// move cursor to a given position
	// use escape sequences
	if cfg.UseEmoji {
		// handle horiz. displacement with emojis
		// scaling col by 2 so every char is in the right place
		fmt.Printf("\x1b[%d;%df", row+1, col*2+1)
	} else {
		fmt.Printf("\x1b[%d;%df", row+1, col+1)
	}
}

func readInput() (string, error) {
	// since we read input on a loop, we are ok in dropping all key presses ina queue and just focus on the last one (more responsive)
	buffer := make([]byte, 100) // create array of bytes with size 100
	cnt, err := os.Stdin.Read(buffer)
	if err != nil {
		return "", err // pass the error up the call stack
	}

	// if we read one byte, see if it is the esc key (hex code 0x1b)
	if cnt == 1 && buffer[0] == 0x1b {
		return "ESC", nil
	} else if cnt >= 3 {
		// escape sequence of arrow keys are 3 bytes long, starting with ESC+[ and then a letter from A to D
		if buffer[0] == 0x1b && buffer[1] == '[' {
			switch buffer[2] {
			case 'A':
				return "UP", nil
			case 'B':
				return "DOWN", nil
			case 'C':
				return "RIGHT", nil
			case 'D':
				return "LEFT", nil
			}
		}
	}

	return "", nil
}

// handle movement
func makeMove(oldRow, oldCol int, dir string) (newRow, newCol int)  {
	newRow, newCol = oldRow, oldCol

	switch dir {
	case "UP":
		newRow = newRow - 1
		if newRow < 0 {
			newRow = len(maze) - 1
		}
	case "DOWN":
		newRow = newRow + 1
		if newRow == len(maze)-1 {
			newRow = 0
		}
	case "LEFT":
		newCol = newCol - 1
		if newCol < 0 {
			newCol = len(maze[0]) - 1
		}
	case "RIGHT":
		newCol = newCol + 1
		if newCol == len(maze[0]) {
			newCol = 0
		}
	}

	if maze[newRow][newCol] == '#' {
		newRow = oldRow
		newCol = oldCol
	}

	return
}

func movePlayer(dir string) {
	player.row, player.col = makeMove(player.row, player.col, dir)
	switch maze[player.row][player.col] {
	case '.':
		numDots--
		score++
		// remove dot from maze
		maze[player.row] = maze[player.row][0:player.col] + " " + maze[player.row][player.col+1:]
		// ^^ make momake new string composed by two slices of the original string
		// ^^ the two slices make the same original string, except for one position that we replace with a space.
	}
}

func moveGhosts() {
    for _, g := range ghosts {
        dir := drawDirection()
        g.row, g.col = makeMove(g.row, g.col, dir)
    }
}

// generate direction for ghosts
func drawDirection() string {
    dir := rand.Intn(4)
		// map the integer numbers to the actual movements using a map
		// map move maps an integer to a string
    move := map[int]string{
        0: "UP",
        1: "DOWN",
        2: "RIGHT",
        3: "LEFT",
    }
    return move[dir]
}

// not using `init` because we want to parse flags before changing console to cbreak
// otherwise, erroneous flag will call os.Exit(), so cleanup() wont run and will break things
func initialize() {
	// init() is a Go function called before main()
	// enable cbreak terminal mode (handle escape sequence)
	cbTerm := exec.Command("/bin/stty", "cbreak", "-echo")
	cbTerm.Stdin = os.Stdin

	err := cbTerm.Run()
	if err != nil {
		// terminate the program after printing this log
		log.Fatalf("Unable to activate cbreak mode terminal: %v\n", err)
	}
}

func cleanup() {
	// restore cooked mode in terminal (handle movement keys)
	cookedTerm := exec.Command("/bin/stty", "-cbreak", "echo")
	cookedTerm.Stdin = os.Stdin

	err := cookedTerm.Run()
	if err != nil {
		log.Fatalf("Unable to activate cooked mode terminal: %v\n", err)
	}
}

func main() {
	// handle flag args
	flag.Parse()

	// initialize game
	initialize()
	defer cleanup()

	// load resources
	err := loadMaze()
	if err != nil {
		log.Printf("Error loading maze: %v\n", err)
		return
	}

	err = loadConfig()
	if err != nil {
		log.Printf("Error loading configuration: %v\n", err)
		return
	}

	// process input (async)
	input := make(chan string)
	go func(ch chan<- string) {
		for {
			// pass input channel as a parameter to an async func thats invoked with the go statement
			input, err := readInput()
			if err != nil {
				log.Printf("Error reading input: %v", err)
				ch <- "ESC"
			}
			ch <- input
		}
	}(input)


	// game loop
	for {
		// update screen
		printScreen()

		// process movement
		select {
		// if the input channel has something to read, it will. otherwise, default break
		case inp := <-input:
			if inp == "ESC" {
				lives = 0
			}
			movePlayer(inp)
		default:
		}

		moveGhosts()

		// process collisions
		for _, g := range ghosts {
			if player.row == g.row && player.col == g.col {
				lives = 0
			}
		}

		// check game over
		if numDots == 0 || lives == 0 {
			if lives == 0 {
				moveCursor(player.row, player.col)
				fmt.Printf(cfg.Death)
				moveCursor(len(maze)+2, 0)
			}
			break
		}

		// repeat
		time.Sleep(200 * time.Millisecond) //delay game since not waiting on input
	}
}
