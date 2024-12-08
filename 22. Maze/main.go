package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type Point struct {
	X, Y int
}

var (
	width, height int
	maze          [][]rune
	playerPos     Point
	exitPos       Point
	visibility    int = 3 // 제한된 시야
)

func initMaze(w, h int) {
	width, height = w, h
	maze = make([][]rune, height)
	for i := range maze {
		maze[i] = make([]rune, width)
		for j := range maze[i] {
			maze[i][j] = '#'
		}
	}
}

func generateMaze() {
	stack := []Point{{1, 1}}
	maze[1][1] = ' '
	rand.Seed(time.Now().UnixNano())

	directions := []Point{
		{0, -2}, // 위
		{0, 2},  // 아래
		{-2, 0}, // 왼쪽
		{2, 0},  // 오른쪽
	}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		rand.Shuffle(len(directions), func(i, j int) {
			directions[i], directions[j] = directions[j], directions[i]
		})

		for _, dir := range directions {
			nx, ny := current.X+dir.X, current.Y+dir.Y
			if nx > 0 && nx < width-1 && ny > 0 && ny < height-1 && maze[ny][nx] == '#' {
				maze[ny][nx] = ' '
				maze[current.Y+dir.Y/2][current.X+dir.X/2] = ' '
				stack = append(stack, Point{nx, ny})
			}
		}
	}

	// 플레이어 시작 위치와 출구 설정
	playerPos = Point{1, 1}
	maze[playerPos.Y][playerPos.X] = '@'
	exitPos = Point{width - 2, height - 2}
	maze[exitPos.Y][exitPos.X] = 'E'
}

func displayMaze() {
	for y := range maze {
		for x := range maze[y] {
			if inVisibilityRange(x, y) {
				fmt.Print(string(maze[y][x]))
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}
}

func inVisibilityRange(x, y int) bool {
	return abs(x-playerPos.X) <= visibility && abs(y-playerPos.Y) <= visibility
}

func movePlayer(dx, dy int) {
	newX, newY := playerPos.X+dx, playerPos.Y+dy
	if newX > 0 && newX < width && newY > 0 && newY < height && maze[newY][newX] != '#' {
		maze[playerPos.Y][playerPos.X] = ' '
		playerPos = Point{newX, newY}
		maze[playerPos.Y][playerPos.X] = '@'
		if playerPos == exitPos {
			fmt.Println("\n축하합니다! 미로를 탈출했습니다!")
			os.Exit(0)
		}
	}
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("=== 터미널 기반 랜덤 미로 탐험 게임 ===")
	fmt.Println("난이도를 선택하세요:")
	fmt.Println("1. Easy (10x10)")
	fmt.Println("2. Medium (20x20)")
	fmt.Println("3. Hard (30x30)")
	fmt.Print("> ")
	scanner.Scan()
	choice := scanner.Text()

	switch choice {
	case "1":
		initMaze(10, 10)
	case "2":
		initMaze(20, 20)
	case "3":
		initMaze(30, 30)
	default:
		fmt.Println("잘못된 입력입니다. 기본 난이도(Easy)로 시작합니다.")
		initMaze(10, 10)
	}

	generateMaze()

	fmt.Println("\n미로를 탈출하세요!")
	fmt.Println("이동: w (위), a (왼쪽), s (아래), d (오른쪽), q (종료)\n")

	for {
		displayMaze()
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		switch input {
		case "w":
			movePlayer(0, -1)
		case "a":
			movePlayer(-1, 0)
		case "s":
			movePlayer(0, 1)
		case "d":
			movePlayer(1, 0)
		case "q":
			fmt.Println("게임을 종료합니다.")
			return
		default:
			fmt.Println("잘못된 입력입니다.")
		}
	}
}
